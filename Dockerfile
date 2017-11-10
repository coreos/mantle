# Image for cross compiling mantle.

FROM ubuntu:17.04

ARG GO_VERSION=1.9.1

RUN apt-get update && apt-get install -y \
	apt-utils \
	gcc \
	gcc-aarch64-linux-gnu \
	git \
	wget \
	&& rm -rf /var/lib/apt/lists/*

RUN wget --no-verbose https://storage.googleapis.com/golang/go${GO_VERSION}.linux-amd64.tar.gz && \
	tar -C /usr/local -xf go${GO_VERSION}.linux-amd64.tar.gz && \
	rm go${GO_VERSION}.linux-amd64.tar.gz

ENV PATH /usr/local/go/bin:${PATH}

CMD /bin/bash
