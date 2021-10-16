package types

import (
	"fmt"
)

type MediaItem struct {
	Artist string
	Album  string
	Track  string
}

func (item *MediaItem) String() string {
	return fmt.Sprintf("%s - %s (%s)", item.Artist, item.Track, item.Album)
}
