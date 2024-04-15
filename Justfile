# setup containers for development

default:
	CONFIG_PATH=./custom-config.yaml air

pod:
	# 10001 = signal-cli
	# 10003 = prometheus
	# 10004 = alertmanager
	# 10005 = grafana
	podman pod create --name aws-pod --infra-name aws-infra \
		-p 127.0.0.1:10001:8080 \
		-p 127.0.0.1:10003:9090 \
		-p 127.0.0.1:10004:9093 \
		-p 127.0.0.1:10005:3000

grafana:
	podman run -d --name aws-grafana \
		--pod aws-pod \
		-v "aws-grafana:/var/lib/grafana:rw,Z" \
		docker.io/grafana/grafana:10.2.2-ubuntu

prometheus:
	podman run -d --name aws-prometheus \
		--pod aws-pod \
		-v "$(pwd)/prometheus-config.yml:/etc/prometheus/prometheus.yml:ro,Z" \
		-v "$(pwd)/prometheus-rules.yml:/etc/prometheus/rules/rules.yml:ro,Z" \
		quay.io/prometheus/prometheus:v2.48.0

alertmanager:
	podman run -d --name aws-alertmanager \
		--pod aws-pod \
		-v "$(pwd)/alertmanager-config.yml:/etc/alertmanager/alertmanager.yml:ro,Z" \
		quay.io/prometheus/alertmanager:v0.26.0 \
			--config.file=/etc/alertmanager/alertmanager.yml

signal:
	podman run -d --name aws-signal \
		--pod aws-pod \
		-v "aws-signal:/home/.local/share/signal-cli:rw,Z" \
		-e 'MODE=native' \
		docker.io/bbernhard/signal-cli-rest-api:0.81

signal_exec:
	podman exec -it --user signal-api bash

alertmanager-webhook-signal:
	podman run -it --rm --name aws-app \
		--pod aws-pod \
		-v "$(pwd)/custom-config.yml:/config.yaml:ro,Z" \
		docker.io/schlauerlauer/alertmanager-webhook-signal:1.0.1

## LINK DEVICE
# just signal_exec
# signal-cli link -n "aws-dev" # copy output
## in bash
# echo "OUTPUT" | xargs -L 1 qrencode -o /tmp/qrcode.png --level=H -v 10 & while [ ! -f /tmp/qrcode.png ]; do sleep 1; done; xdg-open /tmp/qrcode.png
