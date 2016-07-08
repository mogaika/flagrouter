package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/mogaika/flagrouter/router"
)

func DeliveryFunction(flag *router.Flag) error {
	log.Printf("Flag delivered %+#v", flag)
	return nil
}

func main() {
	databaseFile := flag.String("db", "db.sqlite", "sqlite database file")
	//serverAddr := flag.String("addr", "0.0.0.0", "listen ip")
	//serverLowTcpPort := flag.Int("tcp", 100, "tcp low priority port")

	r, err := router.NewRouter(*databaseFile, DeliveryFunction, time.Second)
	if err != nil {
		log.Fatalf("Cannot init router: %v", err)
	}

	rand.Seed(int64(time.Now().Nanosecond()))

	reader := bufio.NewReader(os.Stdin)
	for {
		reader.ReadString('\n')
		var priority byte = router.PRIORITY_HIGH
		if rand.Int()%10 < 5 {
			priority = router.PRIORITY_LOW
		}
		if err := r.AddToQueue(priority, fmt.Sprintf("YES_THIS_IS_FLAG_%d", rand.Int63()), "mogaika_script", "pwn2"); err != nil {
			log.Fatalf("Error inserting: %v", err)
		}
	}

	r.Stop()
}
