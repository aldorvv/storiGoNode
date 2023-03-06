FROM golang:latest AS builder

WORKDIR /build
COPY . .
RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server ./cmd/server

# Run stage
FROM alpine:latest

WORKDIR /server
COPY templates templates
COPY --from=builder /build/server .

CMD ["/bin/sh", "-c", "sleep 60; ./server"]