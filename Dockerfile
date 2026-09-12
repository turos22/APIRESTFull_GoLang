FROM golang:1.26-alpine as builder

WORKDIR /SRC

COPY go.mod go.sum ./
RUN go mod download

COPY . . 

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/api    ./cmd
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/worker ./cmd/worker

FROM golang:1.26-alpine AS ferramentas
RUN go install github.com/pressly/goose/v3/cmd/goose@v3.27.3

FROM gcr.io/distroless/static-debian12:nonroot AS api
COPY --from=builder /out/api /api
EXPOSE 8080
ENTRYPOINT ["/api"]

FROM gcr.io/distroless/static-debian12:nonroot AS worker
COPY --from=builder /out/worker /worker
ENTRYPOINT ["/worker"]

FROM gcr.io/distroless/static-debian12:nonroot AS migrate
COPY --from=ferramentas /go/bin/goose /goose
COPY internal/adapters/postgresql/migrations /migrations
ENTRYPOINT ["/goose", "-dir", "/migrations"]
CMD ["up"]