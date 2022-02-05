package plex

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"plex-beetbrainz/beets"
	env "plex-beetbrainz/environ"
	lb "plex-beetbrainz/listenbrainz"
	"plex-beetbrainz/types"

	goplex "github.com/jrudio/go-plex-client"
)

const defaultPollingModeFreq = 2 * time.Second
const plexUpdateOffsetFreq = 10 * time.Second

type trackedMetadata struct {
	goplex.Metadata
	submitted  bool
	increments int
}

type PlexPoller struct {
	conn       *goplex.Plex
	polFreq    time.Duration
	playingNow map[string]*trackedMetadata
}

func NewPoller(conn *goplex.Plex) (*PlexPoller, error) {
	pollingModeFreq, exists := os.LookupEnv("POLLING_MODE_FREQ")
	polFreq := defaultPollingModeFreq
	if exists {
		num, err := strconv.Atoi(pollingModeFreq)
		if err != nil {
			log.Printf("Polling frequency is not a number, polling will be set to a default value")
		} else {
			polFreq = time.Duration(num) * time.Second
		}
	}

	return &PlexPoller{
		conn:       conn,
		polFreq:    polFreq,
		playingNow: make(map[string]*trackedMetadata),
	}, nil
}

func (pp PlexPoller) Start() error {
	var err error
	for {
		if err != nil {
			break
		}

		pp.pollForSessions()
		time.Sleep(pp.polFreq)
	}

	return err
}

func (pp PlexPoller) pollForSessions() {
	sess, err := pp.conn.GetSessions()
	if err != nil {
		log.Printf("Failed to get sessions from Plex server: %v", err)
	}

	for _, s := range sess.MediaContainer.Metadata {
		err := pp.processTrack(s)
		if err != nil {
			log.Printf("Failed to process track from sessions: %v", err)
		}
	}
}

func (pp PlexPoller) processTrack(m goplex.Metadata) error {
	if m.Type != "track" {
		return fmt.Errorf("%v is not an audio track", m)
	}

	apiToken := env.GetApiToken(m.User.Title)
	if apiToken == "" {
		log.Printf("no listenbrainz API token configured for user '%s'", m.User.Title)
		return nil
	}

	ct, exists := pp.playingNow[m.User.ID]
	if !exists {
		newItem := metadataToMediaItem(m)
		newTrackMeta, err := beets.GetMetadataForItem(newItem)
		if err != nil {
			return fmt.Errorf("failed to get metadata from beets for item '%s': %v", newItem, err)
		}

		pp.playingNow[m.User.ID] = &trackedMetadata{Metadata: m, submitted: false}
		pp.incrementViewOffset(m.User.ID)
		return sendPlayingNow(apiToken, newTrackMeta, newItem, m.User.Title)
	}

	if m.Player.State == "paused" {
		ct.increments = 0
		ct.ViewOffset = m.ViewOffset
		return nil
	}

	if shouldSendListen(ct) {
		log.Printf("Listen submission conditions have been met, sending listen...")
		curItem := metadataToMediaItem(ct.Metadata)
		curTrackMeta, err := beets.GetMetadataForItem(curItem)
		if err != nil {
			return fmt.Errorf("failed to get metadata from beets for item '%s': %v", curItem, err)
		}

		err = lb.SubmitListen(apiToken, curTrackMeta)
		if err != nil {
			return fmt.Errorf("listen submission for item '%s' failed: %v", curItem, err)
		}

		log.Printf("User %s has listened to '%s'", m.User.Title, curItem)
		pp.playingNow[m.User.ID].submitted = true
		return nil
	}

	if metadataEquals(m, ct.Metadata) || ct.submitted {
		log.Println("No change detected or track already submitted")

		// Start optimistically at the beginning
		if m.ViewOffset == 0 {
			pp.incrementViewOffset(m.User.ID)
			return nil
		}

		incrLimit := plexUpdateOffsetFreq.Milliseconds() / pp.polFreq.Milliseconds()
		if int64(ct.increments) >= incrLimit ||
			ct.ViewOffset+int(pp.polFreq.Milliseconds()) <= m.ViewOffset ||
			ct.ViewOffset >= m.ViewOffset+int(plexUpdateOffsetFreq.Milliseconds()) {
			ct.increments = 0
			ct.ViewOffset = m.ViewOffset
		}

		pp.incrementViewOffset(m.User.ID)
		return nil
	}

	newItem := metadataToMediaItem(m)
	newTrackMeta, err := beets.GetMetadataForItem(newItem)
	if err != nil {
		return fmt.Errorf("failed to process item '%s': %v", newItem, err)
	}

	pp.playingNow[m.User.ID] = &trackedMetadata{Metadata: m, submitted: false}
	return sendPlayingNow(apiToken, newTrackMeta, newItem, m.User.Title)
}

// shouldSendListen checks if the tracks has been playing for 240s (4 minutes)
// or more than a half of its duration. Returns true if either of these conditions is true.
func shouldSendListen(m *trackedMetadata) bool {
	return !m.submitted && (m.ViewOffset >= 240000 || m.ViewOffset >= (m.Duration/2))
}

func metadataEquals(m1 goplex.Metadata, m2 goplex.Metadata) bool {
	return m1.GrandparentTitle == m2.GrandparentTitle && m1.ParentTitle == m2.ParentTitle && m1.Title == m2.Title
}

// metadataToMediaItem converts metadata item from goplex to types.MediaItem
func metadataToMediaItem(m goplex.Metadata) *types.MediaItem {
	return &types.MediaItem{
		Artist: m.GrandparentTitle,
		Album:  m.ParentTitle,
		Track:  m.Title,
	}
}

func sendPlayingNow(apiToken string, tm *lb.TrackMetadata, item *types.MediaItem, username string) error {
	err := lb.PlayingNow(apiToken, tm)
	if err != nil {
		return fmt.Errorf("playing now request for item '%s' failed: %v", item.String(), err)
	}

	log.Printf("User %s is now listening to '%s'", username, item.String())
	return nil
}

func (pp PlexPoller) incrementViewOffset(key string) {
	pp.playingNow[key].ViewOffset += int(pp.polFreq.Milliseconds())
	pp.playingNow[key].increments += 1
}
