FROM golang:latest as build

WORKDIR /go/src/app
COPY . /go/src/app

RUN go get -d -v ./...
RUN go build -o /go/bin/plex-beetbrainz

FROM gcr.io/distroless/base
COPY --from=build /go/bin/plex-beetbrainz /

ENTRYPOINT ["plex-beetbrainz"]
