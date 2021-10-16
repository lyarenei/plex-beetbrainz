package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"plex-beetbrainz/plex"
	"plex-beetbrainz/tautulli"
)

func main() {
	addr, exists := os.LookupEnv("BEETBRAINZ_IP")
	if !exists {
		addr = "0.0.0.0"
	}

	port, exists := os.LookupEnv("BEETBRAINZ_PORT")
	if !exists {
		port = "5000"
	}
	address := fmt.Sprintf("%s:%s", addr, port)

	sm := http.NewServeMux()
	sm.HandleFunc("/plex", plex.HandleRequest)
	sm.HandleFunc("/tautulli", tautulli.HandleRequest)

	l, err := net.Listen("tcp4", address)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Starting Beetbrainz, listening on: %s", l.Addr().String())
	log.Fatal(http.Serve(l, sm))
}
