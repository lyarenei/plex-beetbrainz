# plex-beetbrainz
Submit your listens from Plex to ListenBrainz. Integrates with [Beets](https://github.com/beetbox/beets)
for that important metadata.

Note: This app uses Plex's `media.scrobble` event for listen submission.
This event is triggered when a 90% of an item is played (at least according to their documentation),
which does not match ListenBrainz's specification.

## Why
I wanted to track my music activity to ListenBrainz. As I use Plex for the playback,
I was kind of stuck with no options. I dabbled into Jellyfin and actually adapted the Last.fm plugin for ListenBrainz,
but Jellyfin still does not offer the same UX as Plex does (especially on mobile), so this was not the way for me.

Then I found [eavesdrop.fm](https://github.com/simonxciv/eavesdrop.fm), but there was something missing.
Eavesdrop doesn't submit the track metadata as Plex doesn't provide them, which means the submissions are not linked
to MusicBrainz database.

As I want to future proof my experience with ListenBrainz, I want to submit as much data as possible,
so when new features are introduced, the metadata can be immediately used, if needed.

I could add this feature directly into Eavesdrop, but in the end 
I went with python implementation for my own convenience.

## Beets integration
This app/service integrates with Beets for one and only reason.
To pull up the metadata for the currently listened track, so I could submit them to ListenBrainz.
However, this is not a universal solution and only works if you use beets for managing your music library.
Beets stores all metadata about each track, and these data are easily accessible through a simple JSON API.

### How to run
You can run this server without Beets integration, but then you might be better with actually using eavesdrop.fm
mentioned in the beginning, as it's more polished and probably more secure.

Before you run the app, make sure you have all environment variables set as described below.
then run the app with `flask run` command.
To make Flask to bind to all interfaces, you can specify tho host parameter: `flask run -h 0.0.0.0`

Then you can go to your PMS webhook settings and create a new webhook pointing to the IP address or host where this
app is running, together with the port (default 5000) and `/plex` path. For example: `http://localhost:5000/plex`.

#### Configuration
Everything is configured via environment variables, so it's all as simple as possible.
These variables are:
- USER_TOKENS
  - Comma separated list of `<user>:<listenbrainz token>` pairs for configuration and submission.
  - The `user` key must correspond to **Plex** user, not ListenBrainz user.
- LOGGING_LEVEL
  - Logging level. Set to info by default. Can be set to whatever level specified in standard python logging levels.
  - Set to DEBUG to investigate issues.
- BEETS_IP
  - IP address of the Beets web application. Optional.
- BEETS_PORT
  - Port of the Beets web application. Defaults to 8337.

Additionally, a Flask environment variable needs to be set:
- FLASK_APP=app
