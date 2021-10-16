package plex

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

type PlexItem struct {
	Type        string `json:"type"`
	Title       string `json:"title"`
	Parent      string `json:"parentTitle"`
	Grandparent string `json:"grandparentTitle"`
}

func (item *PlexItem) String() string {
	return fmt.Sprintf("%s - %s (%s)", item.Grandparent, item.Title, item.Parent)
}

func (item *PlexItem) AsMediaItem() *types.MediaItem {
	return &types.MediaItem{
		Artist: item.Grandparent,
		Album:  item.Parent,
		Track:  item.Title,
	}
}

type PlexAccount struct {
	Title string `json:"title"`
}

type PlexRequest struct {
	Event   string      `json:"event"`
	Account PlexAccount `json:"Account"`
	Item    PlexItem    `json:"Metadata"`
}

func HandleRequest(w http.ResponseWriter, r *http.Request) {
	payload, err := parseRequest(r)
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

	apiToken := env.GetApiToken(payload.Account.Title)
	if apiToken == "" {
		log.Printf("No API token configured for user '%s'", payload.Account.Title)
		return
	}

	log.Printf("Processing request for item: '%s'...", payload.Item.String())
	tm, err := beets.GetMetadataForItem(payload.Item.AsMediaItem())
	if err != nil {
		log.Printf("Failed to process item '%s': %v", payload.Item.String(), err)
		return
	}

	if payload.Event == "media.play" || payload.Event == "media.resume" {
		err := lb.PlayingNow(apiToken, tm)
		if err != nil {
			log.Printf("Playing now request for item '%s' failed: %v", payload.Item.String(), err)
		} else {
			log.Printf("User %s is now listening to '%s'", payload.Account.Title, payload.Item.String())
		}
		return
	}

	err = lb.SubmitListen(apiToken, tm)
	if err != nil {
		log.Printf("Listen submission for item '%s' failed: %v", payload.Item.String(), err)
	} else {
		log.Printf("User %s has listened to '%s'", payload.Account.Title, payload.Item.String())
	}
}

func parseRequest(r *http.Request) (*PlexRequest, error) {
	r.ParseMultipartForm(16)
	data := []byte(r.FormValue("payload"))

	var plexRequest PlexRequest
	err := json.Unmarshal(data, &plexRequest)
	if err != nil {
		return &plexRequest, err
	}

	return &plexRequest, nil
}

func isEventAccepted(event string) bool {
	return event != "media.play" &&
		event != "media.scrobble" &&
		event != "media.resume"
}
