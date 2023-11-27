# Alertmanager Webhook Signal [![pipeline status](https://gitlab.com/schlauerlauer/alertmanager-webhook-signal/badges/main/pipeline.svg)](https://gitlab.com/schlauerlauer/alertmanager-webhook-signal/-/commits/main)

This project creates a little (dockerized) REST API Endpoint for an [Alertmanager webhook receiver](https://prometheus.io/docs/alerting/latest/configuration/#webhook_config)
and maps it to the [dockerized signal-cli](https://github.com/bbernhard/signal-cli-rest-api).

This is useful if you already have the [signal-cli from bbernhard](https://github.com/bbernhard/signal-cli-rest-api) running as a [Home-Assistant notifier](https://www.home-assistant.io/integrations/signal_messenger/) for example.

Use a prometheus label `recipients` to specify where you want the signal message to be send

It now supports alert webhooks from Grafana aswell, including a preview graph image!

![grafana](media/grafana.png)

![alertmanager](media/alertmanager.jpg)

## Run container

> This image get's pushed automatically to `docker.io/schlauerlauer/alertmanager-webhook-signal` and `registry.gitlab.com/schlauerlauer/alertmanager-webhook-signal`

Default config

```bash
docker run -d --rm --name alertmanager-signal \
  -p 10000:10000 \
  docker.io/schlauerlauer/alertmanager-webhook-signal:latest
```

Custom config

```bash
docker run -d --rm --name alertmanager-signal \
  -p 10000:10000 \
  -v $(pwd)/config.yaml:/root/config.yaml \
  docker.io/schlauerlauer/alertmanager-webhook-signal:latest
```

### Test webhook

```bash
curl -X POST "localhost:10000/api/v3/alertmanager" -d @tests/alert.json
curl -X POST "localhost:10000/api/v3/grafana" -d @tests/grafana.json
```

## Configuration

A `config.yaml` file is needed for configuration.

Example configuration:

```yaml
# Alertmanager webhook url: /api/v2/alertmanager
# Signal webhook url: /api/v2/signal
# Reload this config with a GET request on: /-/reload
server:
  port: 10000 # required
  debug: false
signal:
  number: 23456 # required
  recipients: # required (default recipient, if annotations - recipients is not set in the alert)
  - 123123123
  send: http://10.88.0.1:10001/v2/send # required
alertmanager:
  ignoreLabels: # optional (default is [])
  - alertname
  ignoreAnnotations: [] # optional
  generatorURL: true # optional (default false)
  matchLabel: recipients
recipients: # optional list of recipient names and numbers for label matching
  name1: "123123123"
  name2: "234234234"
```

Example PrometheusRule:

```yaml
groups:
- name: test.rules
  rules:
  - alert: Watchdog
    annotations:
      message: 'Testalert'
    labels:
      recipients: name1
    expr: 'vector(1)'
    for: 1m
```

Example Alertmanager config.yml:

```yaml
global:
route:
  receiver: signal
  group_by: ["alertname"]
  group_wait: 5s
  group_interval: 5m
  repeat_interval: 6d
receivers:
- name: signal
  webhook_configs:
    - url: 'http://10.88.0.1:10000/api/v2/alertmanager'
      send_resolved: true
```

Entry | Example | Explanation | Required
-|-|-|-
server.port | 10000 | Port the script should listen on | yes
signal.number | "+4912345678901" | Phone number of signal cli sender | yes
signal.recipients | ["+4923456789012"] | Phone number(s) of the recipients | yes
signal.send | "http://10.88.0.1:10001/v2/send" | http endpoint of the [signal cli](https://github.com/bbernhard/signal-cli-rest-api) | yes
alertmanager.ignoreLabels | ["alertname"] | Name of label(s) not to include in the signal message | no
alertmanager.ignoreAnnotations | ["message"] | Name of annotation(s) not to include in the signal message | no
alertmanager.generatorURL | true | include prometheus generator link in signal message | no

> Note: if there are errors in the config.yaml the app won't start.
