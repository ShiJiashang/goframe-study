FROM golang:1.26-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY lessons/lesson22/app ./lessons/lesson22/app
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/lesson22 ./lessons/lesson22/app

FROM alpine:3.22

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S app \
    && adduser -S -G app app

WORKDIR /app
RUN mkdir -p /app/manifest/config
COPY --from=builder /out/lesson22 /app/lesson22

USER app
EXPOSE 8000
CMD ["/app/lesson22"]
