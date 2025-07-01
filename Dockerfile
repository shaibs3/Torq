# Build stage
FROM golang:1.21-alpine AS builder

ARG PORT=8080

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Final stage
FROM alpine:latest

ARG PORT=8080
ENV PORT=${PORT}
EXPOSE ${PORT}

RUN apk --no-cache add ca-certificates tzdata wget

RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

WORKDIR /root/

COPY --from=builder /app/main .
COPY --from=builder /app/healthcheck.sh .

RUN chmod +x healthcheck.sh && \
    chown appuser:appgroup /root/main /root/healthcheck.sh

USER appuser

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD ./healthcheck.sh ${PORT}

ENTRYPOINT ["./main"]