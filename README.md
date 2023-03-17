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
      #excludeAlternateTitles: false/ true # only works for Sonarr and defaults to false

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

    - name: whisparr
      type: whisparr
      host: https://yourdomain.com/whisparr
      apikey: YOUR_API_KEY
      filters:
        - 69
      #matchRelease: false / true
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

## Exclude alternative titles from Sonarr

You can drop alternate show titles from being added by setting `excludeAlternateTitles: true` for Sonarr in your config.

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

### run

Run as a service and process on cron schedule. Defaults to every 6 hour `0 */6 * * *`.

## Service

When run as a service it exposes an HTTP server as well. Generate an **API Token** (see instructions above) and add to your config.

To refresh the filters you can make a **POST** or **GET** request to `http://localhost:7441/api/webhook/trigger`.

The API Token can be set as either an HTTP header like `X-API-Token`, or be passed in the url as a query param like `?apikey=MY_NEW_LONG_SECURE_TOKEN`.

### Docker compose

Check the `docker-compose.yml` example. 

1. Set `user: 1000:1000` with your user id you can get with the `id` command, or remove to run as **root**.
2. Set the `volume` so it matches your system. To run from the same path as the `docker-compose` first create a config dir like `mkdir config`, and place this `./config:/config` in the compose file. This will create a default config on the first run.

If you have custom networks then make sure to add those, so it can communicate with autobrr, sonarr and radarr.

### Systemd

On Linux-based systems it is recommended to run omegabrr as a systemd service.

Download the [latest binary](https://github.com/autobrr/omegabrr/releases/latest) for your system and place it in `/usr/bin`. 

Example: Download binary

    wget https://github.com/autobrr/omegabrr/releases/download/$VERSION/omegabrr_$VERSION_linux_x86_64.tar.gz

Extract

    tar -xvf omegabrr_$VERSION_linux_x86_64.tar.gz ~/

Move to somewhere in `$PATH`. Needs to be edited in the systemd service file if using other location.

    sudo mv ~/omegabrr /usr/bin/

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
ExecStart=/usr/bin/omegabrr --config=/home/%i/.config/omegabrr/config.yaml

[Install]
WantedBy=multi-user.target
```

Start the service. Enable will make it startup on reboot.

    sudo systemctl enable -q --now omegabrr@USERNAME

Make sure it's running and **active**

    sudo systemctl status omegabrr@USERNAME.service

By default, the config is set to listen on only `127.0.0.1`. It's highly advised to put it behind a reverse-proxy like nginx or traefik etc.

If you are not running a reverse proxy change host in the `config.toml` to `0.0.0.0`.

On first run it will create a default config, `~/.config/omegabrr/config.yaml` that you will need to edit.

After the config is edited you need to restart the service `systemctl restart omegabrr@USERNAME.service`.