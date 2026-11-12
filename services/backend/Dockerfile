FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/
COPY cctv-visionfusion-backend-app .

RUN chmod +x ./cctv-visionfusion-backend-app

CMD ["./cctv-visionfusion-backend-app"]
