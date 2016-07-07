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

func DeliveryFunction(flag *router.Flag) {
	log.Printf("Flag delivered %+#v", flag)
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
		if err := r.AddToQueue(router.PRIORITY_HIGH, fmt.Sprintf("YES_THIS_IS_FLAG_%d", rand.Int63()), "mogaika_script", "pwn2"); err != nil {
			log.Fatalf("Error inserting: %v", err)
		}
	}

	r.Stop()
}
