package main

import (
	"bufio"
	"flag"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/mogaika/flagrouter/provider"
	_ "github.com/mogaika/flagrouter/provider/http"
	_ "github.com/mogaika/flagrouter/provider/tcp"
	_ "github.com/mogaika/flagrouter/provider/udp"
	"github.com/mogaika/flagrouter/router"
)

func DeliveryFunction(flag *router.Flag) error {
	log.Printf("Flag delivered %+#v", flag)
	return nil
}

func main() {
	databaseFile := flag.String("db", "db.sqlite", "sqlite database file")

	r, err := router.NewRouter(*databaseFile, DeliveryFunction, time.Second)
	if err != nil {
		log.Fatalf("Cannot init router: %v", err)
	}

	provider.InitProviders(r)

	rand.Seed(int64(time.Now().Nanosecond()))

	reader := bufio.NewReader(os.Stdin)
	for {
		if str, err := reader.ReadString('\n'); err != nil || str == "q\n" || str == "quit\n" || str == "exit\n" {
			break
		}
	}

	r.Stop()
}
