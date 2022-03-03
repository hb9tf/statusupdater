FROM golang:1.17 as builder

WORKDIR /usr/src/app
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN CGO_ENABLED=0 go build -v -o /usr/local/bin/statusupdater .


FROM alpine

RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*

WORKDIR /
COPY --from=builder /usr/local/bin/statusupdater /usr/bin/

RUN adduser -S gouser

USER gouser

ENV TOKEN ""
ENV CALLSIGN ""
ENV SLACKCHANNEL ""
ENV FLTR ""

CMD statusupdater -slack_token "$TOKEN" -aprs_callsign "$CALLSIGN" -slack_channel "$SLACKCHANNEL" -aprs_filter "$FLTR"
