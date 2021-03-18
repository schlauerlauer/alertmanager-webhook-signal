FROM docker.io/library/golang:1.15.7 AS builder
WORKDIR /go/src/gitlab.com/schlauerlauer/alertmanager-webhook-signal/
RUN go get -d -v \
    gopkg.in/yaml.v2 \
    github.com/gin-gonic/gin
COPY main.go .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

FROM docker.io/library/alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /go/src/gitlab.com/schlauerlauer/alertmanager-webhook-signal/app .
COPY config.yaml .
EXPOSE 10000
CMD ["./app"]