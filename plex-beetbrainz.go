package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
)

func isEventAccepted(event string) bool {
	return event != "media.play" &&
		event != "media.scrobble" &&
		event != "media.resume"
}

func handlePlex(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(16)
	data := []byte(r.FormValue("payload"))

	payload, err := parseRequest(data)
	if err != nil {
		log.Printf("Failed to decode the Plex request: %v", err)
		return
	}

	if payload.Item.Type != "track" {
		log.Printf("Item '%s' is not a music item, skipping", payload.Item.String())
		return
	}

	if isEventAccepted(payload.Event) {
		log.Printf("Event '%s' is not accepted, ignoring request", payload.Event)
		return
	}

	log.Printf("Received request from Plex for item: '%s'", payload.Item.String())
	var beetsResults []*BeetsData
	var beetsData *BeetsData = nil
	if os.Getenv("BEETS_IP") != "" {
		log.Printf("Getting additional track metadata for item '%s'", payload.Item.String())
		beetsResults, err = getBeetsData(payload.Item.Title)

		logStr := ""
		if len(beetsResults) > 1 {
			for _, br := range beetsResults {
				logStr += fmt.Sprintf("\n\t%s", br.String())
			}
			log.Printf("Received multiple beets results for '%s':%s", payload.Item.String(), logStr)
			beetsData = matchBeetsData(beetsResults, payload.Item)
		} else if len(beetsResults) > 0 {
			beetsData = beetsResults[0]
		} else {
			log.Printf("No beets data received for user '%s'", payload.Item.String())
		}
	}

	tm := TrackMetadata{
		AdditionalInfo: &AdditionalInfo{
			ListeningFrom: "Plex Media Server",
		},
		ArtistName:  payload.Item.Grandparent,
		ReleaseName: payload.Item.Parent,
		TrackName:   payload.Item.Title,
	}

	if err == nil && beetsData != nil {
		log.Printf("Adding Beets data to Listen submission")
		tm.AdditionalInfo.ReleaseMbid = beetsData.ReleaseId
		tm.AdditionalInfo.ArtistMbids = []string{beetsData.ArtistId}
		tm.AdditionalInfo.RecordingMbid = beetsData.RecordingId
	}

	apiToken := getApiToken(payload.Account.Title)
	if apiToken == "" {
		log.Printf("No API token configured for user '%s'", payload.Account.Title)
		return
	}

	if payload.Event == "media.play" || payload.Event == "media.resume" {
		err := playingNow(apiToken, &tm)
		if err != nil {
			log.Printf("Playing now request for item '%s' failed: %v", payload.Item.String(), err)
		} else {
			log.Printf("User %s is now listening to '%s'", payload.Account.Title, payload.Item.String())
		}
		return
	}

	err = submitListen(apiToken, &tm)
	if err != nil {
		log.Printf("Listen submission for item '%s' failed: %v", payload.Item.String(), err)
	} else {
		log.Printf("User %s has listened to '%s'", payload.Account.Title, payload.Item.String())
	}
}

func matchBeetsData(beetsResults []*BeetsData, refItem PlexItem) *BeetsData {
	for _, bd := range beetsResults {
		if refItem.Title == bd.Title &&
			refItem.Parent == bd.Album &&
			refItem.Grandparent == bd.Artist {
			log.Printf("Item '%s' matches with: '%s'", refItem.String(), bd.String())
			return bd
		}
	}

	log.Printf("No match in beets db for item '%s'", refItem.String())
	return nil
}

func getApiToken(user string) string {
	ut := os.Getenv("USER_TOKENS")
	pairs := strings.Split(ut, ",")
	for _, pair := range pairs {
		values := strings.Split(pair, ":")
		if len(values) > 1 && strings.EqualFold(values[0], user) {
			return values[1]
		}
	}

	return ""
}

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
	sm.HandleFunc("/plex", handlePlex)

	l, err := net.Listen("tcp4", address)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Beetbrainz started, listening on: %s", l.Addr().String())
	log.Fatal(http.Serve(l, sm))
}
