# bougou/goceph:latest

FROM ubuntu:24.04 AS base
ENV DEBIAN_FRONTEND=noninteractive

RUN true && \
  apt-get update && \
  apt-get -y upgrade && \
  apt-get install -y gpg locales unzip && \
  locale-gen zh_CN.UTF-8 && \
  apt-get install -y python-is-python3 rsync lsyncd python3 python3-requests python3-arrow runit nginx python3-tz wget vim telnet xattr attr && \
  apt-get install -y gcc ceph-common librados-dev librbd-dev libcephfs-dev && \
  apt-get clean && rm -rf /var/lib/apt/lists/* && \
  true

ENV TERM=xterm
ENV LANG=zh_CN.UTF-8
ENV LANGUAGE=zh_CN:en
ENV LC_ALL=zh_CN.UTF-8

RUN wget https://go.dev/dl/go1.24.12.linux-amd64.tar.gz && \
    rm -rf /usr/local/go && \
    tar -C /usr/local -xzf go1.24.12.linux-amd64.tar.gz && \
    rm go1.24.12.linux-amd64.tar.gz

ENV PATH="/usr/local/go/bin:${PATH}"
