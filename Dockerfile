##
## Build
##

# pull official base image
FROM golang:bullseye as build-env

# set work directory
WORKDIR /usr/src/app

# copy project from local
COPY . /usr/src/app/

# get modules
RUN go mod download

# build aruba_exporter binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -gcflags "all=-N -l" -o ./build/aruba_exporter



##
## Deploy
##

# pull official base image
FROM golang:bullseye

# set work directory
WORKDIR /usr/src/app

# copy binary from build-env container
COPY --from=build-env /usr/src/app/build/aruba_exporter ./

# run binary
CMD ["./aruba_exporter", "-config.file", "./config.yaml"]