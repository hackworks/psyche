FROM alpine

RUN apk --update upgrade && \
    apk add curl ca-certificates && \
    update-ca-certificates && \
    rm -rf /var/cache/apk/*

COPY psyche.linux /opt/service/psyche
WORKDIR /opt/service

EXPOSE 8080

CMD ["/opt/service/psyche"]
