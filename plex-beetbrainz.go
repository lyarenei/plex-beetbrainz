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
		log.Printf(payload.Item.Title, " is not a music item, skipping.")
		return
	}

	if isEventAccepted(payload.Event) {
		log.Printf("Event '%s' is not accepted, ignoring request.", payload.Event)
		return
	}

	log.Printf("Received request from Plex for item: %s", payload.Item.Title)
	var beetsResults []*BeetsData
	var beetsData *BeetsData = nil
	if os.Getenv("BEETS_IP") != "" {
		log.Printf("Beets IP configured - getting additional track metadata for item '%s'", payload.Item.Title)
		beetsResults, err = getBeetsData(payload.Item.Title)

		logStr := ""
		if len(beetsResults) > 1 {
			for _, br := range beetsResults {
				logStr += fmt.Sprintf("\n\t%s - %s (%s)", br.Artist, br.Title, br.Album)
			}
			log.Printf("Received multiple beets results for '%s':%s", payload.Item.Title, logStr)
			beetsData = matchBeetsData(beetsResults, payload.Item)
		} else if len(beetsResults) > 0 {
			beetsData = beetsResults[0]
		} else {
			log.Printf("No beets data received for user '%s'", payload.Item.Title)
		}
	}

	var t *TrackMetadata
	if err != nil || beetsData == nil {
		log.Printf("Plex data will be used for Listen submission.")
		t = &TrackMetadata{
			ArtistName:  payload.Item.Grandparent,
			ReleaseName: payload.Item.Parent,
			TrackName:   payload.Item.Title,
		}
	} else {
		log.Printf("Beets data will be used for Listen submission.")
		t = &TrackMetadata{
			AdditionalInfo: &AdditionalInfo{
				ListeningFrom: "Plex Media Server",
				ReleaseMbid:   beetsData.ReleaseId,
				ArtistMbids:   []string{beetsData.ArtistId},
				RecordingMbid: beetsData.RecordingId,
			},
			ArtistName:  beetsData.Artist,
			ReleaseName: beetsData.Album,
			TrackName:   beetsData.Title,
		}
	}

	apiToken := getApiToken(payload.Account.Title)
	if apiToken == "" {
		log.Printf("No API token configured for user '%s'", payload.Account.Title)
		return
	}

	if payload.Event == "media.play" || payload.Event == "media.resume" {
		playingNow(apiToken, t)
		return
	}

	if submitListen(apiToken, t) {
		log.Printf("Listen submission successful for user '%s' (item '%s')",
			payload.Account.Title, payload.Item.Title)
	}
}

func matchBeetsData(beetsResults []*BeetsData, refItem PlexItem) *BeetsData {
	for _, bd := range beetsResults {
		if refItem.Title == bd.Title &&
			refItem.Parent == bd.Album &&
			refItem.Grandparent == bd.Artist {
			log.Printf("Item '%s - %s (%s)' matches with: %s - %s (%s)",
				refItem.Grandparent,
				refItem.Title,
				refItem.Parent,
				bd.Artist,
				bd.Title,
				bd.Album)
			return bd
		}
	}

	log.Printf("No match in beets db for item '%s - %s (%s)'",
		refItem.Grandparent,
		refItem.Title,
		refItem.Parent)
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
