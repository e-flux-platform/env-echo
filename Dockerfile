## Basic build
FROM golang:1.23-bookworm AS build
WORKDIR /app
COPY . ./
RUN go mod download -x
RUN CGO_ENABLED=0 go build -a -ldflags '-s -extldflags "-static"' -o /env-echo ./main.go

## Deploy
FROM gcr.io/distroless/base-debian12 AS prod
WORKDIR /
COPY --from=build /env-echo /env-echo

ENTRYPOINT ["./env-echo"]
