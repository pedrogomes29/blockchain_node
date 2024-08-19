FROM golang:1.22.5-bullseye as build
    WORKDIR /blockchain/node
    # Copy the Go module files and download dependencies
    COPY go.mod go.sum ./
    RUN go mod download
    # Copy the application source code
    COPY . ./

    # Build the application binary
    RUN go build -o ./node ./main.go
