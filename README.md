# OMEGABRR

Omegabrr transforms items monitored by arrs or lists into autobrr filters. Useful for automating your filters for monitored media or racing criteria.

## Table of Contents

- [Config](#config)
  - [Tags](#tags)
  - [Lists](#lists)
- [Commands](#commands)
- [Service](#service)
  - [Docker Compose](#docker-compose)
  - [Distroless alternative](#distroless-docker-images)
  - [Systemd Setup](#systemd)

## Config

You can set multiple filters per arr. Find the filter ID by going into the webui and get the ID from the url like `http://localhost:7474/filters/10` where `10` is the filter ID.

Create a config like `config.yaml` somewhere like `~/.config/omegabrr`. `mkdir ~/.config/omegabrr && touch ~/.config/omegabrr/config.yaml`.

```yaml
server:
  host: 0.0.0.0
  port: 7441
  apiToken: GENERATED_TOKEN
schedule: "@every 6h"
clients:
  autobrr:
    host: http://localhost:7474
    apikey: YOUR_API_KEY
  arr:
    - name: radarr
      type: radarr
      host: https://yourdomain.com/radarr
      apikey: YOUR_API_KEY
      filters:
        - 15 # Change me
      #matchRelease: false / true
      #includeUnmonitored: false # Set to true to include unmonitored items

    - name: sonarr
      type: sonarr
      # host: http://localhost:PORT
      # host: http://sonarr
      host: https://yourdomain.com/sonarr
      apikey: YOUR_API_KEY
      basicAuth:
        user: username
        pass: password
      filters:
        - 14 # Change me
      #matchRelease: false / true
      #excludeAlternateTitles: false/ true # only works for Sonarr and defaults to false
      #includeUnmonitored: false # Set to true to include unmonitored items

    - name: lidarr
      type: lidarr
      host: https://yourdomain.com/lidarr
      apikey: YOUR_API_KEY
      filters:
        - 13 # Change me
      #matchRelease: false / true

    - name: readarr
      type: readarr
      host: https://yourdomain.com/readarr
      apikey: YOUR_API_KEY
      filters:
        - 12 # Change me

    - name: whisparr
      type: whisparr
      host: https://yourdomain.com/whisparr
      apikey: YOUR_API_KEY
      filters:
        - 69 # Change me
      matchRelease: true # needed as we grab site names

lists:
  - name: Latest TV Shows
    type: mdblist
    url: https://mdblist.com/lists/garycrawfordgc/latest-tv-shows/json
    filters:
      - 1 # Change me

  - name: Anticipated TV
    type: trakt
    url: https://api.autobrr.com/lists/trakt/anticipated-tv
    filters:
      - 22 # Change me

  - name: Upcoming Movies
    type: trakt
    url: https://api.autobrr.com/lists/trakt/upcoming-movies
    filters:
      - 21 # Change me

  - name: Upcoming Bluray
    type: trakt
    url: https://api.autobrr.com/lists/trakt/upcoming-bluray
    filters:
      - 24 # Change me

  - name: Popular TV
    type: trakt
    url: https://api.autobrr.com/lists/trakt/popular-tv
    filters:
      - 25 # Change me

  - name: StevenLu
    type: trakt
    url: https://api.autobrr.com/lists/stevenlu
    filters:
      - 23 # Change me

  - name: New Albums
    type: metacritic
    url: https://api.autobrr.com/lists/metacritic/new-albums
    filters:
      - 9 # Change me

  - name: Upcoming Albums
    type: metacritic
    url: https://api.autobrr.com/lists/metacritic/upcoming-albums
    filters:
      - 20 # Change me

  - name: Steam Wishlist
    type: steam
    url: https://store.steampowered.com/wishlist/id/USERNAME/wishlistdata
    filters:
      - 20 # Change me
```

If you're trying to reach radarr or sonarr hosted on swizzin from some other location, you need to do it like this with basic auth:

```yaml
arr:
  - name: radarr
    type: radarr
    host: https://domain.com/radarr
    apikey: YOUR_API_KEY
    basicAuth:
      user: username
      pass: password
    filters:
      - 15 # Change me
```

Same goes for autobrr if it's behind basic auth.

```yaml
autobrr:
  host: http://localhost:7474
  apikey: YOUR_API_KEY
  basicAuth:
    user: username
    pass: password
```

### Tags

This works for both sonarr and radarr.

If you want to match only certain tags you can use the `tagsInclude`.

```yaml
- name: sonarr
  type: sonarr
  host: http://localhost:8989
  apikey: API_KEY
  filters:
    - 14 # Change me
  tagsInclude:
    - mytag
```

If you want to exclude certain tags, you can use the `tagsExclude`.

```yaml
- name: sonarr
  type: sonarr
  host: http://localhost:8989
  apikey: API_KEY
  filters:
    - 14 # Change me
  tagsExclude:
    - myothertag
```

### Lists

Formerly known as regbrr and maintained by community members is now integrated into omegabrr! We now maintain the lists of media.

**Trakt**

If you are using the Trakt api directly you need to have an **API key** which you can set via the headers object along with any other header needed for the request.

```yaml
lists:
  - name: Some custom Trakt endpoint
    type: trakt
    url: https://api.trakt.tv/calendars/all
    headers:
      trakt-api-key: your_key_goes_here
    filters:
      - 22 # Change me
```

## Optionally use Match Releases field in your autobrr filter

By setting `matchRelease: true` in your config, it will use the `Match releases` field in your autobrr filter instead of fields like `Movies / Shows` and `Albums`.

Readarr will only use the `Match releases` field for now, so setting `matchRelease: false` for Readarr will be ignored.

## Exclude alternative titles from Sonarr

You can drop alternate show titles from being added by setting `excludeAlternateTitles: true` for Sonarr in your config.

### Include Unmonitored Items

By default, omegabrr only processes monitored items. You can include unmonitored items by setting `includeUnmonitored: true` in your arr configuration. This is particularly useful in cross-seed scenarios where you want to match against all items.

## Plaintext lists specific options

Plaintext lists can be anything, therefore you can optionally set `matchRelease: true` or `album: true` to use these fields in your autobrr filter. If not set, it will use the `Movies / Shows` field.

```yaml
lists:
  - name: Personal list
    type: plaintext
    url: https://gist.githubusercontent.com/autobrr/somegist/raw
    filters:
      - 27 # change me
```

## Commands

Available commands.

### generate-token

Generate an API Token to use when triggering via webhook. Copy the output and put in your config like

```yaml
server:
  host: 0.0.0.0
  port: 7441
  apiToken: MY_NEW_LONG_SECURE_TOKEN
```

Call with `omegabrr generate-token`
If you are using docker `docker exec omegabrr omegabrr generate-token`
Optionally call with `--length <number>`for a custom length.

### arr

Call with `omegabrr arr --config config.yaml`

Supports to run with `--dry-run` to only fetch shows and skip filter update.

### lists

Call with `omegabrr lists --config config.yaml`

Supports to run with `--dry-run` to only fetch shows and skip filter update.

### run

Run as a service and process on cron schedule. Defaults to every 6 hour `0 */6 * * *`.

## Service

When run as a service it exposes an HTTP server as well. Generate an **API Token** (see instructions above) and add to your config.

To refresh the filters you can make a **POST** or **GET** request to the following:

- `http://localhost:7441/api/webhook/trigger/arr?apikey=MY_NEW_LONG_SECURE_TOKEN` - This will trigger all arr filters. (Use this in you arr instances)
- `http://localhost:7441/api/webhook/trigger/lists?apikey=MY_NEW_LONG_SECURE_TOKEN` - This will trigger all lists filters.
- `http://localhost:7441/api/webhook/trigger?apikey=MY_NEW_LONG_SECURE_TOKEN` - This will trigger all filters.

The API Token can be set as either an HTTP header like `X-API-Token`, or be passed in the url as a query param like `?apikey=MY_NEW_LONG_SECURE_TOKEN`.

### Docker compose

Check the `docker-compose.yml` example.

1. Set `user: 1000:1000` with your user id you can get with the `id` command, or remove to run as **root**.
2. Set the `volume` so it matches your system. To run from the same path as the `docker-compose` first create a config dir like `mkdir config`, and place this `./config:/config` in the compose file. This will create a default config on the first run.

If you have custom networks then make sure to add those, so it can communicate with autobrr, sonarr and radarr.

### Distroless Docker Images

For users who prioritize container security, we offer alternative Docker images built on [Distroless](https://github.com/GoogleContainerTools/distroless). Specifically the `distroless/static-debian12:nonroot` base image.

Distroless images do not contain a package manager or shell, thereby reducing the potential attack surface and making them a more secure option. These stripped-back images contain only the application and its runtime dependencies.

### Systemd

On Linux-based systems it is recommended to run omegabrr as a systemd service.

Download the [latest binary](https://github.com/autobrr/omegabrr/releases/latest) for your system and place it in `/usr/local/bin`.

Example: Download binary

    wget https://github.com/autobrr/omegabrr/releases/download/$VERSION/omegabrr_$VERSION_linux_x86_64.tar.gz

Extract

    tar -xvf omegabrr_$VERSION_linux_x86_64.tar.gz ~/

Move to somewhere in `$PATH`. Needs to be edited in the systemd service file if using other location.

    sudo mv ~/omegabrr /usr/local/bin/

After that create the config directory for your user:

    mkdir -p ~/.config/omegabrr

You will then need to create a service file in `/etc/systemd/system/` called `omegabrr@.service`.

```shell
touch /etc/systemd/system/omegabrr@.service
```

Then place the following content inside the file (e.g. via nano/vim/ed) or [copy the file ](./distrib/systemd/omegabrr@.service).

```ini
[Unit]
Description=omegabrr service for %i
After=syslog.target network-online.target

[Service]
Type=simple
User=%i
Group=%i
ExecStart=/usr/local/bin/omegabrr run --config=/home/%i/.config/omegabrr/config.yaml

[Install]
WantedBy=multi-user.target
```

Start the service. Enable will make it startup on reboot.

    sudo systemctl enable -q --now omegabrr@$USERNAME

Make sure it's running and **active**

    sudo systemctl status omegabrr@$USERNAME.service

By default, the config is set to listen on only `127.0.0.1`. It's highly advised to put it behind a reverse-proxy like nginx or traefik etc.

If you are not running a reverse proxy change host in the `config.toml` to `0.0.0.0`.

On first run it will create a default config, `~/.config/omegabrr/config.yaml` that you will need to edit.

After the config is edited you need to restart the service `systemctl restart omegabrr@$USERNAME.service`.
