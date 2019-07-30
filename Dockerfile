FROM golang:1.8 as builder

WORKDIR /go/src/updater
COPY . .

RUN go get -d -v ./...
RUN CGO_ENABLED=0 go install -v ./...

FROM alpine

WORKDIR /
COPY --from=builder /go/bin/updater /usr/bin/

RUN adduser -S gouser

USER gouser

ENV TOKEN ""
ENV CALLSIGN ""

CMD updater -slack_token $TOKEN -aprs_callsign $CALLSIGN
