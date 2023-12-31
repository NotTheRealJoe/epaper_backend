# build container
FROM golang:1.20 AS builder
COPY src/go.mod /root/app/
COPY src/go.sum /root/app/
WORKDIR /root/app
RUN go mod download
COPY src/ /root/app
WORKDIR /root/app/cmd
RUN go build .
RUN mv cmd epaper-backend

# runtime container
FROM debian:stable-slim
RUN apt-get -y update && apt-get -y upgrade ca-certificates
COPY --from=builder /root/app/cmd/epaper-backend /usr/local/bin/
COPY src/config.json .
COPY src/static /var/www/static
COPY src/templates /var/www/templates
ENTRYPOINT epaper-backend
