# Run the build
FROM mojlighetsministeriet/go-polymer-faster-build
ENV WORKDIR /go/src/github.com/mojlighetsministeriet/letsencrypt
COPY . $WORKDIR
WORKDIR $WORKDIR
RUN go get -t -v ./...
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build

# Create the final docker image
FROM alpine:3.6
RUN apk upgrade --update --no-cache
RUN apk add --no-cache ca-certificates certbot
RUN mkdir -p /var/www/.well-known
COPY certbot-renew /etc/periodic/daily/
COPY --from=0 /go/src/github.com/mojlighetsministeriet/letsencrypt/letsencrypt /
ENTRYPOINT ["/letsencrypt"]
