FROM debian:stretch-slim

RUN mkdir -p /ftp /web /web/dumps

COPY ddn-api /ddn-api
COPY web /web

VOLUME ["/ftp"]

EXPOSE 7010

ENTRYPOINT ["/ddn-api"]
