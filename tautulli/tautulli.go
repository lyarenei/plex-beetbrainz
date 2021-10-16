package tautulli

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"plex-beetbrainz/beets"
	env "plex-beetbrainz/environ"
	lb "plex-beetbrainz/listenbrainz"
	"plex-beetbrainz/types"
)

type TautulliRequest struct {
	Action      string `json:"action"`
	UserName    string `json:"user_name"`
	ArtistName  string `json:"artist_name"`
	AlbumName   string `json:"album_name"`
	TrackName   string `json:"track_name"`
	TrackArtist string `json:"track_artist"`
	MediaType   string `json:"media_type"`
}

func (r *TautulliRequest) String() string {
	return fmt.Sprintf("%s - %s (%s)", r.ArtistName, r.TrackName, r.AlbumName)
}

func (r *TautulliRequest) AsMediaItem() *types.MediaItem {
	return &types.MediaItem{
		Artist: r.ArtistName,
		Album:  r.AlbumName,
		Track:  r.TrackName,
	}
}

func HandleRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		log.Printf("Request method '%s' is not allowed.", r.Method)
		return
	}

	tr, err := parseRequest(r)
	if err != nil {
		log.Printf("Failed to decode Tautulli request: %v", err)
		return
	}

	if tr.MediaType != "track" {
		log.Printf("Item '%s' is not a music item, skipping...", tr.String())
		return
	}

	if isEventAccepted(tr.Action) {
		log.Printf("Event '%s' is not accepted, ignoring request...", tr.Action)
		return
	}

	apiToken := env.GetApiToken(tr.UserName)
	if apiToken == "" {
		log.Printf("No API token configured for user '%s'", tr.UserName)
		return
	}

	log.Printf("Processing request for item: '%s'...", tr.String())
	tm, err := beets.GetMetadataForItem(tr.AsMediaItem())
	if err != nil {
		log.Printf("Failed to process item '%s': %v", tr.String(), err)
		return
	}

	if tr.Action == "play" || tr.Action == "resume" {
		err := lb.PlayingNow(apiToken, tm)
		if err != nil {
			log.Printf("Playing now request for item '%s' failed: %v", tr.String(), err)
		} else {
			log.Printf("User %s is now listening to '%s'", tr.UserName, tr.String())
		}
		return
	}

	err = lb.SubmitListen(apiToken, tm)
	if err != nil {
		log.Printf("Listen submission for item '%s' failed: %v", tr.String(), err)
	} else {
		log.Printf("User %s has listened to '%s'", tr.UserName, tr.String())
	}
}

func parseRequest(r *http.Request) (*TautulliRequest, error) {
	var tr TautulliRequest
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&tr)
	if err != nil {
		return nil, err
	}

	return &tr, nil
}

func isEventAccepted(event string) bool {
	return event != "play" &&
		event != "watched" &&
		event != "resume"
}
