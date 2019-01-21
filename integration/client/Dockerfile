FROM ubuntu:18.04

ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update && apt-get -y install \
   postgresql-client \
   postgresql-contrib \
   curl \
   dnsutils \
   wget \
   python-pip \
   python-dev \
   build-essential && \
   pip install pika


ENTRYPOINT [ "sleep","infinity" ]
