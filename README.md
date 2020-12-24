# Alertmanager Webhook Signal

This project creates a little (dockerized) REST API Endpoint for an [Alertmanager webhook receiver](https://prometheus.io/docs/alerting/latest/configuration/#webhook_config)
and maps it to the [dockerized signal-cli](https://github.com/bbernhard/signal-cli-rest-api).

This is useful if you already have the [signal-cli from bbernhard](https://github.com/bbernhard/signal-cli-rest-api) running as a [Home-Assistant notifier](https://www.home-assistant.io/integrations/signal_messenger/) for example.

## Configuration

A `config.yaml` file is needed for configuration.

Example configuration:

```yaml
server:
  port: 10000 # required
signal:
  number: 23456 # required
  recipients: # required
  - 67890
  send: http://10.88.0.1:10001/v2/send # required
  ignoreLabels: # optional (default [])
  - alertname
  ignoreAnnotations: [] # optional (default [])
  generatorURL: true # optional (default false)
```

Entry | Example | Explanation | Required
-|-|-|-
server.port | 10000 | Port the script should listen on | yes
signal.number | +4912345678901 | Phone number of signal cli sender | yes
signal.recipients | ["+4923456789012"] | Phone number(s) of the recipients | yes
signal.send | http://10.88.0.1:10001/v2/send | http endpoint of the [signal cli](https://github.com/bbernhard/signal-cli-rest-api) | yes
signal.ignoreLabels | ["alertname"] | Name of label(s) not to include in the signal message | no
signal.ignoreAnnotations | ["message"] | Name of annotation(s) not to include in the signal message | no
signal.generatorURL | true | include prometheus generator link in signal message | no

Example run command:

```Makefile
REGISTRY=registry.gitlab.com
IMAGE=schlauerlauer/alertmanager-webhook-signal
RUNTIME=$(shell which docker or podman 2> /dev/null)
NAME=alertmanager-signal
PORT=10000

pull:
  $(RUNTIME) pull $(REGISTRY)/$(IMAGE)
run:
  $(RUNTIME) run -d --name $(NAME) \
    -v $(CURDIR)/config.yaml:/config.yaml:ro \
    -p $(PORT):10000 \
    $(REGISTRY)/$(IMAGE)
logs:
  $(RUNTIME) logs -f --tail 100 $(NAME)
stop:
  $(RUNTIME) stop $(NAME)
rm:
  $(RUNTIME) container rm $(NAME)
rmi:
  $(RUNTIME) rmi $(REGISTRY)/$(IMAGE)
rebuild: stop rm run
restart: stop start
start:
  $(RUNTIME) start $(NAME)
```

`make run logs`