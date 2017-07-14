FROM golang:1.9-alpine

RUN apk add --update git

WORKDIR /go/src/ecs-station-agent

COPY . .

RUN go-wrapper download

RUN go-wrapper install

ENTRYPOINT ["/go/bin/ecs-station-agent"]
