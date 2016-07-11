package http

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/mogaika/flagrouter/provider"
	"github.com/mogaika/flagrouter/router"
)

var httpProvider = &HttpProvider{}

type HttpProvider struct {
	fr *router.Router
}

func handleFlagPriorited(w http.ResponseWriter, req *http.Request, priority byte) {
	flag, ex := mux.Vars(req)["flag"]
	if ex {
		if err := httpProvider.fr.AddToQueue(priority, flag); err != nil {
			http.Error(w, fmt.Sprintf("Error when pushing flag to queue:\n%v", err), http.StatusInternalServerError)
		} else {
			http.Error(w, "OK", http.StatusOK)
		}
	} else {
		http.Error(w, "Missed flag argument", http.StatusBadRequest)
	}
}

func handlerFlagHi(w http.ResponseWriter, req *http.Request) {
	handleFlagPriorited(w, req, router.PRIORITY_HIGH)
}

func handlerFlagLo(w http.ResponseWriter, req *http.Request) {
	handleFlagPriorited(w, req, router.PRIORITY_LOW)
}

func threadServer(addr string) {
	hr := mux.NewRouter()
	hr.HandleFunc("/flag_hi/{flag}", handlerFlagHi)
	hr.HandleFunc("/flag_lo/{flag}", handlerFlagLo)

	log.Printf("Try to start [http] listener on %s", addr)
	if err := http.ListenAndServe(addr, hr); err != nil {
		log.Printf("Error creating http server: %v", err)
	}
}

func (p *HttpProvider) Init(r *router.Router) error {
	httpProvider.fr = r

	go threadServer(":9999")

	return nil
}

func init() {
	provider.RegisterProvider("http", &HttpProvider{})
}
