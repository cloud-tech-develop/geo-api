FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN go build -o geo-api ./cmd/api

# ─── final image ─────────────────────────────────────────────────────────────
FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/geo-api .
COPY data/ ./data/

ENV PORT=8082
ENV GEO_DATA_PATH=data/countries+states+cities.json

EXPOSE 8082
CMD ["./geo-api"]
