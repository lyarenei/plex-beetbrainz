# Stage 1: build the app
FROM golang:latest as build

WORKDIR /go/src/app
COPY . /go/src/app

RUN go get -d -v ./...
RUN go build -o /go/bin/plex-beetbrainz

# Stage 2: create final image
FROM gcr.io/distroless/base
COPY --from=build /go/bin/plex-beetbrainz /

ENTRYPOINT ["./plex-beetbrainz"]
