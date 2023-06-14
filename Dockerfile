# syntax=docker/dockerfile:1

# Build the application from source
FROM golang:1.18 AS build-stage

WORKDIR /app

ENV GOPROXY=https://goproxy.cn
COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o subserver

# Run the tests in the container
FROM build-stage AS run-test-stage
RUN go test -v ./...

# Deploy the application binary into a lean image
FROM anjia0532/distroless.base-debian11 AS build-release-stage

WORKDIR /app

COPY --from=build-stage /app/subserver /app/subserver

EXPOSE 8008

USER nonroot:nonroot

ENTRYPOINT ["/app/subserver"]
