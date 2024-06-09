############################
# STEP 1 Build
############################
FROM golang:1.22-alpine3.20 AS builder
WORKDIR /build
RUN apk add --no-cache gcc musl-dev
COPY ["go.mod", "go.sum", "./"]
RUN go mod download -x
COPY . .
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o /build/bin/exporter main.go

############################
# STEP 2 Finalize image
############################
FROM alpine:3.20
WORKDIR /
COPY config.example.yaml .
COPY --from=builder /build/bin/exporter /usr/bin/exporter
ENTRYPOINT [ "exporter" ]
CMD [ "metrics" ]