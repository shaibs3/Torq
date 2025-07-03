# Build stage
FROM golang:1.24.2-alpine AS builder

ARG PORT=8080
ARG VERSION=dev
ARG COMMIT=none
ARG DATE=unknown

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags "-X 'main.version=${VERSION}' -X 'main.commit=${COMMIT}' -X 'main.date=${DATE}'" -o main ./cmd/main.go

# Final stage
FROM alpine:latest

ARG PORT=8080
ENV PORT=${PORT}
ENV RPS_LIMIT=10
EXPOSE ${PORT}

RUN apk --no-cache add ca-certificates tzdata wget

RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

WORKDIR /app

COPY --from=builder /app/main .
COPY --from=builder /app/healthcheck.sh .
COPY --from=builder /app/TestFiles ./TestFiles

RUN chmod +x healthcheck.sh && \
    chown -R appuser:appgroup /app/main /app/healthcheck.sh /app/TestFiles

USER appuser

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD ./healthcheck.sh ${PORT}

ENTRYPOINT ["./main"]