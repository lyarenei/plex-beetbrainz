package tautulli

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"plex-beetbrainz/common"
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

	rq, err := parseRequest(r)
	if err != nil {
		log.Printf("Failed to decode Tautulli request: %v", err)
		return
	}

	common.HandleRequest(rq)
}

func parseRequest(r *http.Request) (*common.Request, error) {
	var tr TautulliRequest
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&tr)
	if err != nil {
		return nil, err
	}

	rq := &common.Request{
		Event:     toRequestEvent(tr.Action),
		User:      tr.UserName,
		MediaType: tr.MediaType,
		Item:      tr.AsMediaItem(),
	}

	return rq, nil
}

func toRequestEvent(e string) string {
	if e == "watched" {
		return "scrobble"
	}

	return e
}
