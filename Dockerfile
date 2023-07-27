# First attempt was usign Chat GPT 3.5-turbo and
# it didn't quite work (https://sharegpt.com/c/KaPDmHU)

# Second attempt was from Docker's official repo:
# https://docs.docker.com/language/golang/build-images/

# syntax=docker/dockerfile:1

# Build the application from source
FROM golang:1.20 AS build-stage

RUN apt update && apt install -y git

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY *.go ./
COPY .git/ ./.git

RUN CGO_ENABLED=0 go build -x -o /dagger_git_test

# Deploy the application binary into a lean image
FROM scratch

WORKDIR /

COPY --from=build-stage /dagger_git_test /dagger_git_test

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/dagger_git_test"]
