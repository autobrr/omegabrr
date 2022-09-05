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
  radarr:
    - name: radarr
      host: https://yourdomain.com/radarr
      apikey: YOUR_API_KEY
      filters:
        - 15
  sonarr:
    - name: sonarr
      # host: http://localhost:PORT
      # host: http://sonarr
      host: https://yourdomain.com/sonarr
      apikey: YOUR_API_KEY
      basicAuth:
        user: username
        pass: password
      filters:
        - 14
```

If you're trying to reach radarr or sonarr hosted on swizzin from some other location, you need to do it like this with basic auth:

```yaml
  radarr:
    - name: radarr
      host: https://domain.com/radarr
      apikey: YOUR_API_KEY
      basicAuth:
        user: username
        pass: password
      filters:
        - 15
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

### arr

Call with `omegabrr arr --config config.yaml`

Supports to run with `--dry-run` to only fetch shows and skip filter update.

### run

Run as a service and process on cron schedule. Defaults to every 6 hour `0 */6 * * *`.

## Service

When run as a service it exposes an HTTP server as well. Generate an API Token (see instructions above) and add to your config.

To refresh the filters you can make a POST or GET request to `http://localhost:7441/api/webhook/trigger`.

The API Token can be set as either an HTTP header like `X-API-Token`, or be passed in the url as a query param like `?apikey=MY_NEW_LONG_SECURE_TOKEN`.