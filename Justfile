# setup containers for development

grafana:
	podman run -d --name grafana \
		-p 10005:3000 \
		docker.io/grafana/grafana:10.2.2-ubuntu

prometheus:
	podman run -d --name prometheus \
		-p 10003:9090 \
		-v "$(pwd)/prometheus-config.yml:/etc/prometheus/prometheus.yml:ro,Z" \
		-v "$(pwd)/prometheus-rules.yml:/etc/prometheus/rules/rules.yml:ro,Z" \
		quay.io/prometheus/prometheus:v2.48.0

alertmanager:
	podman run -d --name alertmanager \
		-p 10004:9093 \
		-v "$(pwd)/alertmanager-config.yml:/etc/alertmanager/alertmanager.yml:ro,Z" \
		quay.io/prometheus/alertmanager:v0.26.0 \
			--config.file=/etc/alertmanager/alertmanager.yml


pod:
	podman pod create --name aws-pod --infra-name aws-infra \
		-p 10000:10000

signal:
	podman run -d --name signal-cli \
		--pod aws-pod \
		-v signal-volume:/home/.local/share/signal-cli \
		-e 'MODE=native' \
		docker.io/bbernhard/signal-cli-rest-api:0.70

alertmanager-webhook-signal:
	podman run -it --rm --name alertmanager-webhook-signal \
		--pod aws-pod \
		-v "$(pwd)/custom-config.yml:/config.yaml:ro,Z" \
		docker.io/schlauerlauer/alertmanager-webhook-signal:1.0.1
