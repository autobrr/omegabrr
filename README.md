# OMEGABRR

Gathers shows from the arrs and puts into filter in autobrr.

## Config

You can set multiple filters per arr. Find the filter ID by going into the webui and get the ID from the url like `http://localhost:7474/filters/10` where `10` is the filter ID.

Create a config like `config.yaml` somewhere like `~/.config/omegabrr`. `mkdir ~/.config/omegabrr && touch ~/.config/omegabrr/config.yaml`.

```yaml
clients:
  autobrr:
    host: http://localhost:7474
    apikey: YOUR_API_KEY
  radarr:
    host: https://yourdomain.com/radarr
    apikey: YOUR_API_KEY
    filters:
      - 15
  sonarr:
    # host: http://localhost:PORT
    # host: http://sonarr
    host: https://yourdomain.com/sonarr
    apikey: YOUR_API_KEY
    basicauth:
      user: username
      pass: password
    filters:
      - 14
```

If you're trying to reach radarr or sonarr hosted on swizzin from some other location, you need to do it like this with basic auth:

```yaml
  radarr:
    host: https://domain.com/radarr
    apikey: YOUR_API_KEY
    basicauth:
      user: username
      pass: password
    filters:
      - 15
```

## Commands

Available commands.

### arr

Call with `omegabrr arr --config config.yaml`

Supports to run with `--dry-run` to only fetch shows and skip filter update.