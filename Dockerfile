FROM golang:1.22-alpine AS builder
WORKDIR /app
RUN apk add --no-cache ca-certificates git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server ./...

FROM gcr.io/distroless/static:nonroot
WORKDIR /app

COPY --from=builder /app/server /app/server
EXPOSE 13370
USER nonroot:nonroot
ENTRYPOINT ["/app/server"]
