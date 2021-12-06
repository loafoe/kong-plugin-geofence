FROM golang:1.17.4-alpine3.14 AS builder
WORKDIR /build
COPY go.mod .
COPY go.sum .
# Get dependancies - will also be cached if we won't change mod/sum
RUN go mod download
# Build
COPY . .
RUN go build -o /build/geofence geofence.go


FROM kong:2.6.0-alpine
USER root
COPY --from=builder /build/geofence /usr/local/bin
USER kong
