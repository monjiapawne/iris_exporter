# Version pinning
ARG ALPINE_VERSION=3.22
ARG GOLANG_VERSION=1.26

# Build
FROM golang:${GOLANG_VERSION} AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/iris_exporter ./cmd/iris_exporter

# Run
FROM alpine:${ALPINE_VERSION} AS runtime
RUN adduser -D -H -s /sbin/nologin promuser
RUN apk --no-cache add ca-certificates

COPY --from=build /out/iris_exporter /bin/iris_exporter

USER promuser
EXPOSE 10043
ENTRYPOINT [ "/bin/iris_exporter" ]
