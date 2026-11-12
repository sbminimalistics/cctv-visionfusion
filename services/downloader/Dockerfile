FROM golang:1.26.11-alpine
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app/

COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY proto/ ./proto/
COPY go.mod ./go.mod
COPY go.sum ./go.sum

RUN go mod tidy

RUN set GOOS=linux&& set GOARCH=amd64&& set CGO_ENABLED=0&& go build -o cctv-visionfusion-downloader-app ./cmd/app/main.go

RUN chmod +x ./cctv-visionfusion-downloader-app

CMD ["./cctv-visionfusion-downloader-app"]
