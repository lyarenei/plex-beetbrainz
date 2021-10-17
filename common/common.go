package common

import (
	"log"

	"plex-beetbrainz/beets"
	env "plex-beetbrainz/environ"
	lb "plex-beetbrainz/listenbrainz"
	"plex-beetbrainz/types"
)

type Request struct {
	Event     string
	MediaType string
	User      string
	Item      *types.MediaItem
	Metadata  *lb.TrackMetadata
}

func HandleRequest(r *Request) {
	if r.MediaType != "track" {
		log.Printf("Item '%s' is not a music item, skipping...", r.Item.String())
		return
	}

	if isEventAccepted(r.Event) {
		log.Printf("Event '%s' is not accepted, ignoring request...", r.Event)
		return
	}

	apiToken := env.GetApiToken(r.User)
	if apiToken == "" {
		log.Printf("No API token configured for user '%s'", r.User)
		return
	}

	log.Printf("Processing request for item: '%s'...", r.Item.String())
	tm, err := beets.GetMetadataForItem(r.Item)
	if err != nil {
		log.Printf("Failed to process item '%s': %v", r.Item.String(), err)
		return
	}

	if r.Event == "play" || r.Event == "resume" {
		err := lb.PlayingNow(apiToken, tm)
		if err != nil {
			log.Printf("Playing now request for item '%s' failed: %v", r.Item.String(), err)
		} else {
			log.Printf("User %s is now listening to '%s'", r.User, r.Item.String())
		}
		return
	}

	err = lb.SubmitListen(apiToken, tm)
	if err != nil {
		log.Printf("Listen submission for item '%s' failed: %v", r.Item.String(), err)
	} else {
		log.Printf("User %s has listened to '%s'", r.User, r.Item.String())
	}
}

func isEventAccepted(event string) bool {
	return event != "play" &&
		event != "scrobble" &&
		event != "resume"
}
