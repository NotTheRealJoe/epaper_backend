# build container
FROM golang:1.20 AS builder
COPY src/ /root/app
WORKDIR /root/app
RUN go mod download
WORKDIR /root/app/cmd
RUN go build .
RUN mv cmd epaper-backend

# runtime container
FROM debian:stable-slim
RUN apt-get -y update && apt-get -y upgrade ca-certificates
COPY --from=builder /root/app/cmd/epaper-backend /usr/local/bin/
COPY src/config.json .
ENTRYPOINT epaper-backend