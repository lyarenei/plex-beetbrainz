import json
import os
import time
from typing import Dict, List, Optional
from urllib.parse import quote

from flask import Flask, request, Response
from pylistenbrainz import Listen, ListenBrainz
from pylistenbrainz.errors import InvalidAuthTokenException
from requests import get

tokens: Dict[str, str] = {}
for pair in os.environ.get('USER_TOKENS', '').split(','):
    user, token = pair.split(':', maxsplit=1)
    tokens[user] = token

app = Flask(__name__)
logger = app.logger
logger.setLevel(os.environ.get('LOGGING_LEVEL', 'INFO'))


@app.route("/plex", methods=['POST'])
def plex_request():
    request_data = request.values.to_dict().get('payload', '')
    request_data = json.loads(request_data)

    account = request_data.get('Account', {}).get('title', '').lower()
    if account not in tokens.keys():
        logger.info(f"User '{account}' is not enabled - ignoring request.")
        return Response(status=204)

    event = request_data.get('event')

    # Return if not applicable event
    if event not in ['media.play', 'media.scrobble']:
        logger.debug("Received an irrelevant event - not play or scrobble")
        return Response(status=204)

    metadata: Dict[str, str] = request_data.get('Metadata', {})
    item_type = metadata.get('type')

    # Return if not playing music
    if item_type != 'track':
        logger.debug("Not playing music")
        return Response(status=204)

    title = metadata.get('title')
    album = metadata.get('parentTitle')
    artist = metadata.get('grandparentTitle')

    beets_data = None
    if os.environ.get('BEETS_IP'):
        logger.info("Using beets to get metadata")
        beets_data = get_beets_data(title)

        # If found multiple results, try to match the item
        if beets_data and len(beets_data) > 1:
            logger.warning(f"Multiple results found for '{title}'")
            logger.debug(f"Received multiple results:\n{beets_data}")
            for item in beets_data:
                if title == item['title'] and (
                        artist == item['artist_credit']
                        or artist == item['artist']
                        or artist == item['albumartist']
                ):
                    beets_data = item
        elif beets_data:
            beets_data = beets_data[0]

    if beets_data:
        logger.info(f"Using beets metadata for listen submission of '{title}'")
        logger.debug(f"Got metadata from beets: {beets_data}")
        listen = Listen(
            track_name=beets_data['title'],
            release_name=beets_data['album'],
            artist_name=beets_data['artist'],
            recording_mbid=beets_data['mb_trackid'],
            release_mbid=beets_data['mb_albumid'],
            artist_mbids=[beets_data['mb_artistid']],
            release_group_mbid=beets_data['mb_releasegroupid'],
            work_mbids=[beets_data['mb_workid']],
        )
    else:
        logger.info(f"Using plex data for listen submission of '{title}'")
        listen = Listen(
            track_name=title,
            release_name=album,
            artist_name=artist,
        )

    listen.listening_from = 'Plex Media Server'
    lb = ListenBrainz()

    try:
        lb.set_auth_token(tokens[account])
    except InvalidAuthTokenException:
        logger.error(f"Invalid listenbrainz token for user '{account}'")
        return Response(status=400)

    if event == 'media.play':
        lb.submit_playing_now(listen)
    else:
        listen.listened_at = int(time.time())
        try:
            lb.submit_single_listen(listen)
        except Exception as e:
            logger.error(f"Listen submission failed:\n{e}")
            return Response(status=400)

    return Response(status=200)


def get_beets_data(title: str) -> Optional[List[Dict[str, str]]]:
    url = f"http://{os.environ['BEETS_IP']}:" \
          f"{os.environ.get('BEETS_PORT', 8337)}" \
          f"/item/query/title:{quote(title)}"

    try:
        logger.debug(f"Sending request: {url}")
        r = get(url)
        return r.json()['results']
    except Exception as e:
        logger.error(f"Failed to get data from beets:\n{e}")
