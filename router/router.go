package router

import (
	"log"
	"sync"
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

type Flag struct {
	Id        int
	Priority  byte      `gorm:"not null"`
	Flag      string    `gorm:"unique;not null"`
	Timestamp time.Time `gorm:"not null"`
	Source    string
	Task      string
	Status    byte `gorm:"not null"`
}

type Router struct {
	DB                  *gorm.DB
	LastDeliveryTime    time.Time
	Lock                sync.Mutex
	FlagSendPeriod      time.Duration
	threadCommunication chan int
	DeliveryFunction    func(*Flag)
}

func NewRouter(databaseFile string, deliveryFunction func(*Flag), flagSendPeriod time.Duration) (*Router, error) {
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
			time.Sleep(time.Millisecond * 50)
			tc <- THREAD_TICK
		}
	}(r.threadCommunication)

	return r, nil
}

func (r *Router) AddToQueue(priority byte, flag string, source string, task string) error {
	err := r.DB.Create(&Flag{
		Priority:  priority,
		Flag:      flag,
		Timestamp: time.Now(),
		Source:    source,
		Task:      task,
		Status:    STATUS_QUEUED,
	}).Error

	r.threadCommunication <- THREAD_TICK
	return err
}

func (r *Router) ProcessFlagQueue() error {
	f := &Flag{}
	if err := r.DB.Where("status = ?", STATUS_QUEUED).Order("priority DESC").Order("timestamp").First(f).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil
		} else {
			return err
		}
	}
	r.DeliveryFunction(f)
	r.LastDeliveryTime = time.Now()
	return r.DB.Delete(f).Error
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