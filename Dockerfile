FROM golang:1.8 as builder

WORKDIR /go/src/updater
COPY . .

RUN go get -d -v ./...
RUN CGO_ENABLED=0 go install -v ./...

FROM alpine

RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*

WORKDIR /
COPY --from=builder /go/bin/updater /usr/bin/

RUN adduser -S gouser

USER gouser

ENV TOKEN ""
ENV CALLSIGN ""
ENV SLACKCHANNEL ""

CMD updater -slack_token $TOKEN -aprs_callsign $CALLSIGN -slack_channel $SLACKCHANNEL
