##
## Build
##

# pull official base image
FROM golang:bullseye

# copy project
COPY . /usr/src/app/

# Copy application data into image
COPY . /go/src/bartmika/mullberry-backend
WORKDIR /go/src/bartmika/mullberry-backend

COPY go.mod ./
COPY go.sum ./
RUN go mod download

# Copy only `.go` files, if you want all files to be copied then replace `with `COPY . .` for the code below.
COPY *.go .

# Build our application.
# RUN go build -o /go/src/bartmika/mullberry-backend/bin/mullberry-backend
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -gcflags "all=-N -l" -o /server

##
## Deploy
##
FROM alpine:latest
RUN mkdir /data

COPY --from=dev-env /server ./
CMD ["./server"]