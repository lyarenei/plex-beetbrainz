package plex

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"plex-beetbrainz/common"
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
	rq, err := parseRequest(r)
	if err != nil {
		log.Printf("Failed to decode the Plex request: %v", err)
		return
	}

	common.HandleRequest(rq)
}

func parseRequest(r *http.Request) (*common.Request, error) {
	err := r.ParseMultipartForm(16)
	if err != nil {
		return nil, err
	}

	data := []byte(r.FormValue("payload"))
	var plexRequest PlexRequest
	err = json.Unmarshal(data, &plexRequest)
	if err != nil {
		return nil, err
	}

	rq := &common.Request{
		Event:     toRequestEvent(plexRequest.Event),
		User:      plexRequest.Account.Title,
		MediaType: plexRequest.Item.Type,
		Item:      plexRequest.Item.AsMediaItem(),
	}

	return rq, nil
}

func toRequestEvent(e string) string {
	switch e {
	case "media.play":
		return "play"
	case "media.resume":
		return "resume"
	default:
		return "scrobble"
	}
}
