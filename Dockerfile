# pull official base image
FROM golang:bullseye

# set work directory
WORKDIR /usr/src/app

# check go version
RUN go version
RUN go install github.com/githubnemo/CompileDaemon@latest

# copy project
COPY . /usr/src/app/
