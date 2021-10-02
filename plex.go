package main

import (
	"encoding/json"
	"fmt"
)

type PlexItem struct {
	Type        string `json:"type"`
	Title       string `json:"title"`
	Parent      string `json:"parentTitle"`
	Grandparent string `json:"grandparentTitle"`
}

func (item PlexItem) String() string {
	return fmt.Sprintf("%s - %s (%s)", item.Grandparent, item.Title, item.Parent)
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
