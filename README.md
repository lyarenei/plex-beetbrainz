# plex-beetbrainz
Submit your listens from Plex to ListenBrainz. Integrates with [beets](https://github.com/beetbox/beets)
for that important metadata.

Note: This app uses Plex's `media.scrobble` webhook event for listen submission, 
which does not conform to the ListenBrainz's specification for listen submission.

## Why
I want to track my music activity to ListenBrainz. As I use Plex for music playback,
I was kind of stuck with no options. I dabbled into Jellyfin and actually 
[adapted](https://github.com/lyarenei/jellyfin-plugin-listenbrainz) the Last.fm plugin for ListenBrainz.
However, Jellyfin still does not offer the same user experience level as Plex does (especially on mobile),
so I still don't use it primarily.

There is [eavesdrop.fm](https://github.com/simonxciv/eavesdrop.fm), 
but it doesn't submit the track metadata (as Plex doesn't provide them), 
which means that the submitted Listens are not linked to MusicBrainz database entries.

As I want to future proof my experience with ListenBrainz, 
I want to submit as much data as possible - so when new features are introduced, the metadata can be used, if applicable.

## How to run
Naturally, the beets integration only works if you have a beets library somewhere available 
and you also use beets to manage your music library (to avoid no matches).
If you decide to edit artist/album/track names, you need to do so in both Plex and beets, so the metadata match will be possible.
If you want, you can also run this app without beets integration. In that case, only the data provided by Plex will be submitted.

---

There are two ways to run this app.
The first one is running the binary as-is. The second option is to use docker image.
There are 64-bit binaries provided for macOS, Linux and Windows.
The binaries are available [here](https://github.com/lyarenei/plex-beetbrainz/releases).

The docker image is only available for 64-bit Linux platform and is available [here](https://github.com/lyarenei/plex-beetbrainz/pkgs/container/plex-beetbrainz)

If you need another platform, you can easily compile the app or build the docker image yourself.
Refer to [Building](#building) section for more information.

Before you run the app, make sure you have all environment variables set as described below in the [Configuration](#configuration) section.
Don't forget to go to your PMS webhook settings and create a new webhook pointing to the IP address or host where this
app is running, together with the port (default 5000) and `/plex` path. For example: `http://localhost:5000/plex`.

### Configuration
There are few configuration options, all set via environment variables.
These variables are:
- USER_TOKENS
  - Comma separated list of `<user>:<listenbrainz token>` pairs for configuration and submission.
  - The `user` key must correspond to **Plex** user, not ListenBrainz user.
  - The `user` matching is performed case-insensitively.
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
If you need the binary or docker image on your specific platform, or just simply want to compile the code (or build the docker image) yourself...

### Compilation
This app is written in Go language, so you need to set up Golang development environment first. Refer to [this guide](https://golang.org/doc/install) for more information.
After you have everything set up, simply clone this repository and run `go build`. This should produce a binary named `plex-beetbrainz` in the current directory.

### Building the docker image
Image build is two-staged to minimize the image size. At first, a Golang image is downloaded to build the app and then the binary is copied into [distroless](https://github.com/GoogleContainerTools/distroless) image.

Provided you have docker installed, simply clone the repository and run `docker build . -t <your_image_tag>`.
