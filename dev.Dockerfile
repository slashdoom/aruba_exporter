# pull official base image
FROM golang:bullseye

# set work directory
WORKDIR /usr/src/app

# check go version
RUN go version

# install application for hot-reloading capability.
RUN go install github.com/githubnemo/CompileDaemon@latest

# copy project
COPY . /usr/src/app/

ENTRYPOINT /go/bin/CompileDaemon -polling -build="go build ./build/aruba_exporter" -command="./build/aruba_exporter -config.file ./config.yaml" -directory="./"