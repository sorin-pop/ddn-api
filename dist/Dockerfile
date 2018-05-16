FROM alpine

RUN mkdir -p /ftp /web /web/dumps

VOLUME ["/ftp"]

EXPOSE 7010

COPY web /web
COPY ddn-api /ddn-api

ENTRYPOINT ["/ddn-api"]
