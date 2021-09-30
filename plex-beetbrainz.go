package main

import (
	"log"
	"net"
	"net/http"
	"os"
	"strings"
)

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

	if payload.Event != "media.play" && payload.Event != "media.scrobble" {
		log.Printf("Received irrelevant event to process (not play or scrobble): %s", payload.Event)
		return
	}

	log.Printf("Received request from Plex for item: %s", payload.Item.Title)
	var beetsResults []*BeetsData
	var beetsData *BeetsData = nil
	if os.Getenv("BEETS_IP") != "" {
		log.Printf("Beets IP configured - getting additional track metadata for item '%s'", payload.Item.Title)
		beetsResults, err = getBeetsData(payload.Item.Title)

		if len(beetsResults) > 1 {
			log.Printf("Received multiple beets results for '%s'", payload.Item.Title)
			beetsData = matchBeetsData(beetsResults, payload.Item.Title)
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

	if payload.Event == "media.play" {
		playingNow(apiToken, t)
		return
	}

	if submitListen(apiToken, t) {
		log.Printf("Listen submission successful for user '%s' (item '%s')",
			payload.Account.Title, payload.Item.Title)
	}
}

func matchBeetsData(beetsResults []*BeetsData, refItem string) *BeetsData {
	for _, bd := range beetsResults {
		if refItem == bd.Title {
			return bd
		}
	}

	log.Printf("No beets data matched to '%s'", refItem)
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
	sm := http.NewServeMux()
	sm.HandleFunc("/plex", handlePlex)

	l, err := net.Listen("tcp4", ":5000")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Beetbrainz started, listening on: %s", l.Addr().String())
	log.Fatal(http.Serve(l, sm))
}