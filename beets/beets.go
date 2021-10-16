package beets

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	lb "plex-beetbrainz/listenbrainz"
	"plex-beetbrainz/types"
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

func GetMetadataForItem(item *types.MediaItem) (*lb.TrackMetadata, error) {
	tm := lb.TrackMetadata{
		AdditionalInfo: &lb.AdditionalInfo{
			ListeningFrom: "Plex Media Server",
		},
		ArtistName:  item.Artist,
		ReleaseName: item.Album,
		TrackName:   item.Track,
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

func matchWithBeets(item *types.MediaItem) (*BeetsData, error) {
	beetsResults, err := getBeetsData(item.Track)
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

func getBeetsData(title string) ([]*BeetsData, error) {
	beetsIP := os.Getenv("BEETS_IP")
	beetsPort := os.Getenv("BEETS_PORT")
	if beetsPort == "" {
		beetsPort = "8337"
	}

	url := "http://" + beetsIP + ":" + beetsPort + "/item/query/title:" + url.PathEscape(title)
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

func matchBeetsData(beetsResults []*BeetsData, refItem *types.MediaItem) (*BeetsData, error) {
	for _, bd := range beetsResults {
		if refItem.Track == bd.Title &&
			refItem.Album == bd.Album {
			if refItem.Artist == "Various Artists" {
				log.Printf("Item '%s' partially matches with: '%s'", refItem.String(), bd.String())
				return bd, nil
			} else if refItem.Artist == bd.Artist {
				log.Printf("Item '%s' exactly matches with: '%s'", refItem.String(), bd.String())
				return bd, nil
			}
		}
	}

	log.Printf("No match in beets db for item '%s'", refItem.String())
	return nil, errors.New("no match for item")
}
