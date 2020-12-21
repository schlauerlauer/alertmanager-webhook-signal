certs:
	cp /etc/ssl/certs/ca-certificates.crt $(CURDIR)
go:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .
build:
	podman build . -t rest:test
run:
	podman run --rm -p 10000:10000 rest:test
