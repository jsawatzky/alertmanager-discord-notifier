# alertmanager-discord-notifier
Send AlertManager notifications to a discord webhook

This program accepts webhook notifications from [AlertManager](https://github.com/prometheus/alertmanager).
It is not a replacement to AlertManager. This application does not accept raw alert notification from Prometheus.

## Usage
The program can be configured with both command line flags and environment variables. The available
options are described below:

| Cmd Flag | Env Var | Description | Required? | Default |
| :---: | :---: | --- | :---: | --- |
| --webhook, -w | ADN_WEBHOOK | The Discord webhook URL to send alerts to | Yes | none |
| --listen, -l | ADN_LISTEN | The address to listen for AlertManager notifications | No | `0.0.0.0:9094` |
|  --debug, -d | ADN_DEBUG | Enable debug logging | No | false |

## Deployment
The easiest way to deploy this program is with Docker. It can also be build from source and deployed manually
### Docker Run
`docker run --rm -d -p 9094:9094 ghcr.io/jsawatzky/alertmanager-discord:latest -w <discord_webhook>`

### Docker Compose
    version: '3.3'
    services:
      discord-notifier:
        image: ghcr.io/jsawatzky/alertmanager-discord:latest
        restart: unless-stopped
        environment:
          - ADN_WEBHOOK=${DISCORD_WEBHOOK}

### Build From Source
To build from source ensure you have Go and Git installed and run the following commands

    git clone https://github.com/jsawatzky/alertmanager-discord-notifier.git
    cd alertmanager-discord-notifier
    go get
    go build -o adn

The program can then be run using `./adn -w <discord_webhook>`

## AlertManager Configuration
An example AlertManager configuration to leverage this program would be:

    ...
    receivers:
      - name: 'example'
        webhook_configs:
          - url: 'http://<program_url>:9094'
    ...

Replacing `<program_url>` with the URL pointing to your deployment of the application.

## Future Plans
 - In future revisions, the message sent to discord will be fully configurable using the same templating
method as AlertManager allowing complete control over the appearance and content of your alerts
 - Allow sending alerts to different channels based on the path of the AlertManager webhook request