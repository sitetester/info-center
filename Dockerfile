# Dockerfile References: https://docs.docker.com/engine/reference/builder/
# https://docs.docker.com/language/golang/build-images/#create-a-dockerfile-for-the-application

# Start from golang:1.16-alpine base image
FROM golang:latest

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . ./

# Build the Go app
RUN go build -o info-center .

RUN chmod +x ./info-center

EXPOSE ${GIN_PORT}

# Run the executable
CMD [ "./info-center" ]