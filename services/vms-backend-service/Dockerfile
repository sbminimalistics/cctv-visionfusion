FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/
COPY vms-backend-app .
COPY config.yml .

RUN chmod +x ./vms-backend-app

CMD ["./vms-backend-app"]
