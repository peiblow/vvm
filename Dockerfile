# ─────────────────────────────────────────────
#  synx-core / vvm
#  Go multi-stage build
# ─────────────────────────────────────────────
 
# Stage 1 — Build
FROM golang:1.25-alpine AS builder
 
WORKDIR /app
 
RUN apk add --no-cache git ca-certificates tzdata
RUN apk add --no-cache netcat-openbsd
 
COPY go.mod go.sum ./
RUN go mod download
 
COPY . .
 
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" -o /bin/vvm .
 
# Stage 2 — Runtime (minimal image)
FROM scratch
 
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /bin/vvm /bin/vvm
 
# SWP (Synx Wire Protocol) TCP port
EXPOSE 4000
 
ENTRYPOINT ["/bin/vvm"]