package main

import (
	"errors"
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

func matchWithBeets(item PlexItem) (*BeetsData, error) {
	beetsResults, err := getBeetsData(item.Title)
	if err != nil {
		return nil, err
	}

	logStr := ""
	if len(beetsResults) > 1 {
		for _, br := range beetsResults {
			logStr += fmt.Sprintf("\n\t%s", br.String())
		}

		log.Printf("Received multiple beets results for '%s':%s", item.String(), logStr)
		return matchBeetsData(beetsResults, item)
	} else if len(beetsResults) > 0 {
		return beetsResults[0], nil
	} else {
		log.Printf("No beets data received for item '%s'", item.String())
		return nil, errors.New("no beets data received")
	}
}

func processItem(item PlexItem) (*TrackMetadata, error) {
	tm := TrackMetadata{
		AdditionalInfo: &AdditionalInfo{
			ListeningFrom: "Plex Media Server",
		},
		ArtistName:  item.Grandparent,
		ReleaseName: item.Parent,
		TrackName:   item.Title,
	}

	var beetsData *BeetsData
	if os.Getenv("BEETS_IP") != "" {
		log.Printf("Getting additional track metadata for item '%s'...", item.String())
		beetsData, _ = matchWithBeets(item)
	}

	if beetsData != nil {
		log.Printf("Adding Beets data to Listen submission")
		tm.AdditionalInfo.ReleaseMbid = beetsData.ReleaseId
		tm.AdditionalInfo.ArtistMbids = []string{beetsData.ArtistId}
		tm.AdditionalInfo.RecordingMbid = beetsData.RecordingId

		tm.ArtistName = beetsData.Artist
		tm.ReleaseName = beetsData.Album
		tm.TrackName = beetsData.Title
	}

	return &tm, nil
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
		log.Printf("Item '%s' is not a music item, skipping...", payload.Item.String())
		return
	}

	if isEventAccepted(payload.Event) {
		log.Printf("Event '%s' is not accepted, ignoring request...", payload.Event)
		return
	}

	apiToken := getApiToken(payload.Account.Title)
	if apiToken == "" {
		log.Printf("No API token configured for user '%s'", payload.Account.Title)
		return
	}

	log.Printf("Processing request for item: '%s'...", payload.Item.String())
	tm, err := processItem(payload.Item)
	if err != nil {
		log.Printf("Failed to process item '%s': %v", payload.Item.String(), err)
		return
	}

	if payload.Event == "media.play" || payload.Event == "media.resume" {
		err := playingNow(apiToken, tm)
		if err != nil {
			log.Printf("Playing now request for item '%s' failed: %v", payload.Item.String(), err)
		} else {
			log.Printf("User %s is now listening to '%s'", payload.Account.Title, payload.Item.String())
		}
		return
	}

	err = submitListen(apiToken, tm)
	if err != nil {
		log.Printf("Listen submission for item '%s' failed: %v", payload.Item.String(), err)
	} else {
		log.Printf("User %s has listened to '%s'", payload.Account.Title, payload.Item.String())
	}
}

func matchBeetsData(beetsResults []*BeetsData, refItem PlexItem) (*BeetsData, error) {
	for _, bd := range beetsResults {
		if refItem.Title == bd.Title &&
			refItem.Parent == bd.Album {
			if refItem.Grandparent == "Various Artists" {
				log.Printf("Item '%s' partially matches with: '%s'", refItem.String(), bd.String())
				return bd, nil
			} else if refItem.Grandparent == bd.Artist {
				log.Printf("Item '%s' exactly matches with: '%s'", refItem.String(), bd.String())
				return bd, nil
			}
		}
	}

	log.Printf("No match in beets db for item '%s'", refItem.String())
	return nil, errors.New("no match for item")
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

	log.Printf("Starting Beetbrainz, listening on: %s", l.Addr().String())
	log.Fatal(http.Serve(l, sm))
}
