FROM debian:jessie
MAINTAINER Alex Wilson a.wilson@alumni.warwick.ac.uk

RUN apt-get update && \
  apt-get install -qy golang-go git make

RUN mkdir -p /usr/local/go/bin
ENV GOPATH /usr/local/go
ENV GOBIN /usr/local/go/bin
ENV PATH $PATH:$GOBIN

RUN go get github.com/mrwilson/helixdns && \
  cd /usr/local/go/src/github.com/mrwilson/helixdns && \
  make install

EXPOSE 53
