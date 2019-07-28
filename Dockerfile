FROM golang:1.8 as builder

WORKDIR /go/src/updater
COPY . .

RUN apt-get update && apt-get install -y patch
RUN go get -d -v ./...
RUN patch /go/src/github.com/nlopes/slack/users.go slack.users.go.diff
RUN CGO_ENABLED=0 go install -v ./...

FROM alpine

WORKDIR /
COPY --from=builder /go/bin/updater /usr/bin/

RUN adduser -S gouser

USER gouser

ENV TOKEN ""
ENV CALLSIGN ""

CMD updater -token $TOKEN -callsign $CALLSIGN
