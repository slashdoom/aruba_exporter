##
## Development
##

# pull official base image
FROM golang:bullseye

# set work directory
WORKDIR /go/aruba_exporter

# install application for hot-reloading capability.
RUN go install github.com/githubnemo/CompileDaemon@latest

# run app in CompileDaemon
ENTRYPOINT /go/bin/CompileDaemon -directory="./" -build="go build -o ./build/" -command="./build/aruba_exporter -config.file ./config.yaml"