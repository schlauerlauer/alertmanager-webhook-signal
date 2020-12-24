REGISTRY=registry.gitlab.com
IMAGE=schlauerlauer/alertmanager-webhook-signal

certs:
	cp /etc/ssl/certs/ca-certificates.crt $(CURDIR)
main:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./src
signal:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o signal ./tests
run:
	podman run --rm -p 10000:10000 rest:test

login:
	podman login $(REGISTRY)
build: main
	podman build -t $(REGISTRY)/$(IMAEG) .
push:
	podman push $(REGISTRY)/$(IMAGE)