###############################################################################
FROM golang:1.15 as builder
RUN mkdir -p /data/src
COPY controller /data/src/
WORKDIR /data/src/
RUN go build -o kubernetes-notificator && chmod 770 kubernetes-notificator

###############################################################################
From ubuntu:focal-20211006

COPY --from=builder \
  /data/src/kubernetes-notificator \
  /usr/local/bin/kubernetes-notificator
COPY --from=builder \
  /etc/ssl/certs/ca-certificates.crt \
  /etc/ssl/certs/

ENTRYPOINT ["/usr/local/bin/kubernetes-notificator"]
