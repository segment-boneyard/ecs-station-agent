FROM library/golang:1.9-alpine

RUN apk add --update git

WORKDIR /go/src/ecs-station-agent

COPY . .

RUN go-wrapper download

RUN go-wrapper install

ENTRYPOINT ["/ecs-station-agent"]
