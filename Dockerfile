FROM alpine

COPY psyche.linux /opt/service/psyche
WORKDIR /opt/service

EXPOSE 8080

CMD ["/opt/service/psyche"]
