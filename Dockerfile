FROM golang:latest

MAINTAINER Chris Purta <cpurta@gmail.com>

RUN apt-get update && \
    mkdir -p /app

ADD . /app

WORKDIR /app

RUN bash /app/bootstrap.sh
