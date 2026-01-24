# build
FROM golang:1.25-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -o qashare

# run
FROM gcr.io/distroless/base-debian12

WORKDIR /app

COPY --from=builder /build/qashare /app/qashare

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/app/qashare"]
