## Multi-stage Dockerfile for ascii-art-web
# Build stage
ARG GOVERSION=1.20
FROM golang:${GOVERSION}-alpine AS builder
LABEL org.opencontainers.image.title="ascii-art-web"
LABEL org.opencontainers.image.description="Small Go web server that converts text to ASCII art"
LABEL org.opencontainers.image.licenses="MIT"
LABEL org.opencontainers.image.source=""
WORKDIR /src

# Install git and build deps for private modules if necessary (none here) and ca-certificates
RUN apk add --no-cache ca-certificates

# Cache go mod downloads
COPY go.mod .
RUN go env -w GOPROXY=https://proxy.golang.org,direct
RUN go mod download || true

# Copy sources
COPY . .

# Build a statically linked binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags='-s -w' -o /app/ascii-art-web ./

# Final stage: small runtime image
FROM scratch AS runtime
LABEL org.opencontainers.image.title="ascii-art-web"
LABEL org.opencontainers.image.version="1.0.0"
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /app/ascii-art-web /ascii-art-web
# Copy templates and font files from the build context so they are available at runtime
COPY --from=builder /src/templates /templates
COPY --from=builder /src/*.txt /
EXPOSE 8080

# Metadata
LABEL maintainer="<your-name> <you@example.com>"
LABEL org.opencontainers.image.url=""

USER 65532:65532

ENV PORT=8080

ENTRYPOINT ["/ascii-art-web"]
