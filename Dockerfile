FROM golang:alpine AS builder
WORKDIR /build
COPY . /build
RUN CGO_ENABLED=0 GOOS=linux go build -o /build/traefik-log-separator cmd/traefik-log-separator/main.go
RUN apk add -U --no-cache ca-certificates

FROM scratch
EXPOSE 8080
COPY --from=builder /build/traefik-log-separator /app
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT ["/app","-i", "/log/access.log", "-o", "/log/output/"]
