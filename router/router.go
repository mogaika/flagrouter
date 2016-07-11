package router

import (
	"errors"
	"log"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
)

const (
	PRIORITY_LOW = iota
	PRIORITY_HIGH
)

const (
	STATUS_QUEUED = iota
	STATUS_ERROR
	STATUS_DELIVERED
)

const (
	THREAD_STOP = iota
	THREAD_TICK
)

var ErrAlreadyQueued = errors.New("Flag already queued")

type Flag struct {
	Id        int
	Priority  byte      `gorm:"not null"`
	Flag      string    `gorm:"unique;not null"`
	Timestamp time.Time `gorm:"not null"`
	Status    byte      `gorm:"not null"`
}

type Router struct {
	DB                  *gorm.DB
	LastDeliveryTime    time.Time
	FlagSendPeriod      time.Duration
	threadCommunication chan int
	DeliveryFunction    func(*Flag) error
}

func NewRouter(databaseFile string, deliveryFunction func(*Flag) error, flagSendPeriod time.Duration) (*Router, error) {
	db, err := gorm.Open("sqlite3", databaseFile)

	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&Flag{}).Error; err != nil {
		return nil, err
	}

	r := &Router{
		DB:                  db,
		LastDeliveryTime:    time.Now(),
		threadCommunication: make(chan int, 32),
		DeliveryFunction:    deliveryFunction,
		FlagSendPeriod:      flagSendPeriod,
	}

	go r.DeliveryThread()

	go func(tc chan int) {
		for {
			time.Sleep(time.Millisecond * 150)
			tc <- THREAD_TICK
		}
	}(r.threadCommunication)

	return r, nil
}

func (r *Router) AddToQueue(priority byte, flag string) error {
	err := r.DB.Create(&Flag{
		Priority:  priority,
		Flag:      flag,
		Timestamp: time.Now(),
		Status:    STATUS_QUEUED,
	}).Error

	r.threadCommunication <- THREAD_TICK
	return err
}

func (r *Router) ProcessFlagQueue() error {
	f := &Flag{}
	if err := r.DB.Where("status = ?", STATUS_QUEUED).Order("priority DESC").Order("timestamp ASC").First(f).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil
		} else {
			return err
		}
	}

	deliveryerr := r.DeliveryFunction(f)
	r.LastDeliveryTime = time.Now()

	if deliveryerr != nil {
		f.Status = STATUS_ERROR
	} else {
		f.Status = STATUS_DELIVERED
	}
	return r.DB.Select("status").Save(f).Error
}

func (r *Router) DeliveryThread() {
	log.Printf("Entered in delivery thread")

	for {
		switch <-r.threadCommunication {
		case THREAD_TICK:
			if r.LastDeliveryTime.Add(r.FlagSendPeriod).Before(time.Now()) {
				if err := r.ProcessFlagQueue(); err != nil {
					log.Printf("Timed process flag error: %v", err)
				}
			}
		case THREAD_STOP:
			break
		}
	}
	log.Printf("Delivery thread stop")
}

func (r *Router) Stop() {
	r.threadCommunication <- THREAD_STOP
}
