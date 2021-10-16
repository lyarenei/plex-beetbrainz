# plex-beetbrainz
Submit your listens from Plex to ListenBrainz. Integrates with [beets](https://github.com/beetbox/beets)
for that important metadata.

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

Before you run the app, make sure you have all environment variables set as described below in the [Configuration](#configuration) section. To configure the webhook itself, read the following section.

### Webhook configuration
Starting from version 1.2, you can choose between Plex or [Tautulli](https://tautulli.com) webhooks. Obviously, each webhook "type" has its own advantages and disadvantages.

Plex webhook is easier to set up, you just configure the URL and that's it. But, you also need Plexpass to use webhooks at all.

Tautulli webhooks are a bit more work, unlike Plex. First disadvantage is that you of course need the Tautulli itself. If you don't know about it, you can learn more [here](https://tautulli.com). Another disadvantage is - as already mentioned - that it's just a bit more work to configure it. However, Tautulli does not require Plex pass for most of the stuff it does and you also have a more control over the webhooks themselves. Another advantage is that the webhooks seem to be reliable unlike Plex ones, but that's something I did not extensively test.

Each webhook setup is described in the following sections.

#### Plex
To configure Plex webhook, go to your PMS webhook settings and create a new webhook pointing to the IP address or host where this app is running, together with the port (default 5000) and `/plex` path. For example: `http://localhost:5000/plex`.

Note that for the listen submission, a Plex's `media.scrobble` is used. This event does not conform to the ListenBrainz's specification for listen submission (4 minutes or half of a track).

#### Tautulli
If you want to use Tautulli instead of Plex for webhooks, you need to properly configure the webhook in Tautulli. You can create a webhook in Tautulli under Notification Agents in Settings. Following sections explain every tab of the webhook configuration in a detail.

##### Configuration tab
The webhook URL is the same as if you'd use Plex, however the path is `/tautulli`. For example: `http://localhost:5000/tautulli`. The webhook method should be set to `POST`. 

##### Triggers tab
Select `Playback Start` and `Watched`. Optionally, you can also select `Playback Resume` if you want your now playing status at Listebrainz to be a bit more precise. Watched percentage can be configured in the `General` settings. According to Listenbrainz guidelines, the Music Listened percentage should be set to 50%, but of course, that is everyone's own decision. Personally, I left it at 85%, which is the default value.

##### Conditions tab
You can skip this tab entirely, but if you want to minimize network traffic for some reason, then you can limit the webhook requests to be sent only if the `Media Type` equals to `track` and you can also allow only webhook events for users which are configured in the beetbrainz app. That said, these checks exists on the app side anyway, so this is completely optional as already mentioned.

##### Data tab
For each trigger selected in [Triggers tab](#triggers-tab), paste this JSON string into **JSON Data** field:
```json
{
  "action": "{action}",
  "user_name": "{username}",
  "artist_name": "{artist_name}",
  "album_name": "{album_name}",
  "track_name": "{track_name}",
  "track_artist": "{track_artist}",
  "media_type": "{media_type}"
}
```

And that's all. Don't forget to save the changes and you should be good to go.

### Beetbrainz configuration
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
Image build is two-staged to minimize the image size. At first, a Golang image is downloaded to build the app and then the binary is copied into a [distroless](https://github.com/GoogleContainerTools/distroless) image.

Provided you have docker installed, simply clone the repository and run `docker build . -t <your_image_tag>`.
