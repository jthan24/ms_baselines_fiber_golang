FROM golang:1.18 as build

WORKDIR /go/src/app
RUN sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b /usr/local/bin

COPY . .
RUN task go:mod:download && task go:build "EXE=/go/bin/app"

FROM gcr.io/distroless/static-debian11
COPY --from=build /go/bin/app /
CMD ["/app"]
