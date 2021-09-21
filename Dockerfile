FROM golang:1.17.1 AS plugins
WORKDIR /build
RUN git clone https://github.com/Kong/go-pluginserver.git && cd go-pluginserver && go build
RUN cd /build
COPY go.mod .
COPY go.sum .
# Get dependancies - will also be cached if we won't change mod/sum
RUN go mod download
# Build
COPY . .
RUN go build -o /build/geofence geofence.go


FROM kong:2.5.0-ubuntu
USER root
RUN mkdir -p /plugins
COPY --from=plugins /build/geofence /plugins
COPY --from=plugins /build/go-pluginserver/go-pluginserver /usr/local/bin
USER kong
