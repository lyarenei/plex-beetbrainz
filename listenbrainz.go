package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type AdditionalInfo struct {
	ListeningFrom string   `json:"listening_from,omitempty"`
	ReleaseMbid   string   `json:"release_mbid,omitempty"`
	ArtistMbids   []string `json:"artist_mbids,omitempty"`
	RecordingMbid string   `json:"recording_mbid,omitempty"`
	Tags          []string `json:"tags,omitempty"`
}

type TrackMetadata struct {
	AdditionalInfo *AdditionalInfo `json:"additional_info,omitempty"`
	ArtistName     string          `json:"artist_name"`
	TrackName      string          `json:"track_name"`
	ReleaseName    string          `json:"release_name"`
}

type ListenPayload struct {
	ListenedAt    int64          `json:"listened_at,omitempty"`
	TrackMetadata *TrackMetadata `json:"track_metadata"`
}

type Listen struct {
	ListenType string           `json:"listen_type"`
	Payload    []*ListenPayload `json:"payload"`
}

func listenbrainzSubmitRequest(apiToken string, listen Listen) bool {
	apiUrl := "https://api.listenbrainz.org/1/submit-listens"

	bdata, err := json.Marshal(listen)
	// log.Println(string(bdata))
	if err != nil {
		log.Printf("Failed to encode listen into json: %v", err)
		return false
	}

	r, err := http.NewRequest("POST", apiUrl, bytes.NewBuffer(bdata))
	if err != nil {
		log.Printf("Failed to create a new request: %v", err)
		return false
	}

	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Authorization", "Token "+apiToken)

	client := &http.Client{}
	resp, err := client.Do(r)
	if err != nil {
		log.Printf("Request to Listenbrainz failed: %v", err)
		return false
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		b, _ := ioutil.ReadAll(resp.Body)
		log.Printf("Listen submission failed: %s", string(b))
		return false
	}

	return true
}

func playingNow(apiToken string, trackMetadata *TrackMetadata) bool {
	l := Listen{
		ListenType: "playing_now",
		Payload: []*ListenPayload{
			{
				TrackMetadata: trackMetadata,
			},
		},
	}

	return listenbrainzSubmitRequest(apiToken, l)
}

func submitListen(apiToken string, trackMetadata *TrackMetadata) bool {
	l := Listen{
		ListenType: "single",
		Payload: []*ListenPayload{
			{
				ListenedAt:    time.Now().Unix(),
				TrackMetadata: trackMetadata,
			},
		},
	}

	return listenbrainzSubmitRequest(apiToken, l)
}
