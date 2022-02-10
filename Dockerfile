FROM golang:alpine as builder

RUN apk update && apk add git && apk add ca-certificates
RUN adduser -D appuser
COPY . $GOPATH/src/
WORKDIR $GOPATH/src/
RUN go get -d -v
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o /go/bin/adn

FROM scratch

LABEL org.opencontainers.image.source https://github.com/jsawatzky/alertmanager-discord-notifier

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /go/bin/adn /go/bin/adn

EXPOSE 9094
USER appuser
ENTRYPOINT ["/go/bin/adn"]
