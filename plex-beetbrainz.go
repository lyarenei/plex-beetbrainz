package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"plex-beetbrainz/plex"
	"plex-beetbrainz/tautulli"

	goplex "github.com/jrudio/go-plex-client"
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

	pollingModeEnabled, exists := os.LookupEnv("POLLING_MODE")
	if exists && strings.EqualFold(strings.ToLower(pollingModeEnabled), "true") {
		log.Print("POLLING_MODE set to true, will run in polling mode")

		plexAddr, _ := os.LookupEnv("PLEX_ADDR")
		plexPort, _ := os.LookupEnv("PLEX_PORT")
		plexToken, _ := os.LookupEnv("PLEX_TOKEN")
		plexURL := fmt.Sprintf("http://%s:%s", plexAddr, plexPort)

		plexConn, err := goplex.New(plexURL, plexToken)
		if err != nil {
			log.Fatalf("Failed to connect to Plex server: %v", err)
		}

		poller, err := plex.NewPoller(plexConn)
		if err != nil {
			log.Fatalf("Failed to start in polling mode: %v", err)
		}

		log.Fatal(poller.Start())
	}

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
