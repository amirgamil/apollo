FROM golang:alpine3.14 as builder

WORKDIR /build
COPY . /build

ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

RUN apk add --no-cache git && \
    git apply disable_integrations.patch

RUN go build -a -tags netgo -ldflags '-w -extldflags "-static"' -o apollo cmd/apollo.go && \
    chmod +x ./apollo

FROM alpine:3.14

LABEL maintainer="Polonio Davide <poloniodavide@gmail.com>"
LABEL description="Docker image for Apollo personal search engine"

EXPOSE 8993

WORKDIR /opt/apollo
VOLUME /opt/apollo/data

RUN apk add --no-cache \
    youtube-dl \
    ffmpeg && \
    touch .env

COPY --from=builder /build/apollo /opt/apollo/apollo
COPY static/ /opt/apollo/static/

ENTRYPOINT ["/opt/apollo/apollo"]
