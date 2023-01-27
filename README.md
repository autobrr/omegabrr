# OMEGABRR

Gathers shows from the arrs and puts into filter in autobrr.

## Config

You can set multiple filters per arr. Find the filter ID by going into the webui and get the ID from the url like `http://localhost:7474/filters/10` where `10` is the filter ID.

Create a config like `config.yaml` somewhere like `~/.config/omegabrr`. `mkdir ~/.config/omegabrr && touch ~/.config/omegabrr/config.yaml`.

```yaml
server:
  host: 0.0.0.0
  port: 7441
  apiToken: GENERATED_TOKEN
schedule: "0 */6 * * *"
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
        - 15
      #matchRelease: false / true

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
        - 14
      #matchRelease: false / true

    - name: lidarr
      type: lidarr
      host: https://yourdomain.com/lidarr
      apikey: YOUR_API_KEY
      filters:
        - 13
      #matchRelease: false / true

    - name: readarr
      type: readarr
      host: https://yourdomain.com/readarr
      apikey: YOUR_API_KEY
      filters:
        - 12
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
        - 15
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
    - 14
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
    - 14
  tagsExclude:
    - myothertag
```

## Optionally use Match Releases field in your autobrr filter

By setting `matchRelease: true` in your config, it will use the `Match releases` field in your autobrr filter instead of fields like `Movies / Shows` and `Albums`.

Readarr will only use the `Match releases` field for now, so setting `matchRelease: false` for Readarr will be ignored.

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

### arr

Call with `omegabrr arr --config config.yaml`

Supports to run with `--dry-run` to only fetch shows and skip filter update.

### run

Run as a service and process on cron schedule. Defaults to every 6 hour `0 */6 * * *`.

## Service

When run as a service it exposes an HTTP server as well. Generate an **API Token** (see instructions above) and add to your config.

To refresh the filters you can make a **POST** or **GET** request to `http://localhost:7441/api/webhook/trigger`.

The API Token can be set as either an HTTP header like `X-API-Token`, or be passed in the url as a query param like `?apikey=MY_NEW_LONG_SECURE_TOKEN`.

## Docker compose

Check the `docker-compose.yml` example. 

1. Set `user: 1000:1000` with your user id you can get with the `id` command, or remove to run as **root**.
2. Set the `volume` so it matches your system. To run from the same path as the `docker-compose` first create a config dir like `mkdir config`, and place this `./config:/config` in the compose file. This will create a default config on the first run.

If you have custom networks then make sure to add those, so it can communicate with autobrr, sonarr and radarr.
