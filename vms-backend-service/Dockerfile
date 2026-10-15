FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/
COPY vms-app .
COPY config.yml .

RUN chmod +x ./vms-app

CMD ["./vms-app"]
