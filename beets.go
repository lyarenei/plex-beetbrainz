package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
)

type BeetsData struct {
	Title          string `json:"title"`
	Album          string `json:"album"`
	Artist         string `json:"artist"`
	RecordingId    string `json:"mb_trackid"`
	ReleaseId      string `json:"mb_albumid"`
	ArtistId       string `json:"mb_artistid"`
	ReleaseGroupId string `json:"mb_releasegroupid"`
	WorkId         string `json:"mb_workid"`
}

func (data *BeetsData) String() string {
	return fmt.Sprintf("%s - %s (%s)", data.Artist, data.Title, data.Album)
}

func getBeetsData(title string) ([]*BeetsData, error) {
	beetsIp := os.Getenv("BEETS_IP")
	beetsPort := os.Getenv("BEETS_PORT")
	if beetsPort == "" {
		beetsPort = "8337"
	}

	url := "http://" + beetsIp + ":" + beetsPort + "/item/query/title:" + url.PathEscape(title)
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Beets request failed: %v", err)
		return []*BeetsData{}, err
	}

	defer resp.Body.Close()

	var rj map[string][]*BeetsData
	err = json.NewDecoder(resp.Body).Decode(&rj)
	if err != nil {
		log.Printf("Failed to decode the response from beets: %v", err)
		return []*BeetsData{}, err
	}

	return rj["results"], nil
}
