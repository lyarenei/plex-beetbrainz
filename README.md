# plex-beetbrainz
Submit your listens from Plex in ListenBrainz. Integrates with [Beets](https://github.com/beetbox/beets)
for that important metadata.

Note: This app uses Plex's `media.scrobble` event for listen submission, 
which does not conform to the ListenBrainz's specification for listen submission.

## Why
I wanted to track my music activity to ListenBrainz. As I use Plex for the playback,
I was kind of stuck with no options. I dabbled into Jellyfin and actually 
[adapted](https://github.com/lyarenei/jellyfin-plugin-listenbrainz) the Last.fm plugin for ListenBrainz.
However, Jellyfin still does not offer the same user experience level as Plex does (especially on mobile),
so I still don't use it primarily.

Then I found [eavesdrop.fm](https://github.com/simonxciv/eavesdrop.fm), but there was something missing.
Eavesdrop doesn't submit the track metadata (as Plex doesn't provide them), which means that the submitted Listens are not linked
to MusicBrainz database entries.

As I want to future proof my experience with ListenBrainz, 
I want to submit as much data as possible - so when new features are introduced, the metadata can be immediately used, if needed.

## Beets integration
This app/service integrates with beets to get the metadata of the currently listened track, so I could submit them to ListenBrainz.
However, this is not a universal solution and only works if you use beets for managing your music library.
Beets stores MusicBrainz metadata of each track, and can be easily accessed through a simple JSON API.

## How to run
You can run this server without Beets integration, but then you might be better with actually using `eavesdrop.fm`
mentioned in the beginning, as it's more polished and probably more secure.

You can either build (or get) docker image or run the app directly.
Before you run the app, make sure you have all environment variables set as described below in the Configuration section.

Don't forget to go to your PMS webhook settings and create a new webhook pointing to the IP address or host where this
app is running, together with the port (default 5000) and `/plex` path. For example: `http://localhost:5000/plex`.

### Configuration
There are few configuration options, all set via environment variables.
These variables are:
- USER_TOKENS
  - Comma separated list of `<user>:<listenbrainz token>` pairs for configuration and submission.
  - The `user` key must correspond to **Plex** user, not ListenBrainz user.
- BEETS_IP
  - IP address of the Beets web application. If not set, the beets integration is simply disabled.
- BEETS_PORT
  - Port of the Beets web application. Defaults to 8337. Has no effect if `BEETS_IP` is not set.
- BEETBRAINZ_IP
  - Bind to specified IP of an interface. If not set, all interfaces will be used (0.0.0.0).
  - Applies only to IPv4. IPv6 is not supported.
- BEETBRAINZ_PORT
  - Listen on the specified port. Defaults to 5000.


## Building
This app is written in Go language, so you need to set up Golang development environment first. 
After you have everything set up, simply clone this repository and run `go build`. This should produce a binary named `plex-beetbrainz` in the current directory.
