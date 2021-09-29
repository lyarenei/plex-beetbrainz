package main

import (
	"encoding/json"
)

type PlexItem struct {
	Type        string `json:"type"`
	Title       string `json:"title"`
	Parent      string `json:"parentTitle"`
	Grandparent string `json:"grandparentTitle"`
}

type PlexAccount struct {
	Title string `json:"title"`
}

type PlexRequest struct {
	Event   string      `json:"event"`
	Account PlexAccount `json:"Account"`
	Item    PlexItem    `json:"Metadata"`
}

func parseRequest(data []byte) (*PlexRequest, error) {
	var plexRequest PlexRequest
	err := json.Unmarshal(data, &plexRequest)
	if err != nil {
		return &plexRequest, err
	}

	return &plexRequest, nil
}
