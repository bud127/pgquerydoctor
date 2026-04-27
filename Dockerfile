FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod ./
COPY . .
RUN go build -o /pgquerydoctor ./cmd/pgquerydoctor

FROM alpine:3.20
COPY --from=builder /pgquerydoctor /usr/local/bin/pgquerydoctor
WORKDIR /workspace
ENTRYPOINT ["pgquerydoctor"]
