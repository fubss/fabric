# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

ARG GO_VER
ARG ALPINE_VER
#FROM golang:${GO_VER}-alpine${ALPINE_VER} as golang
FROM fubss/rocksdb:6.27.3-golang${GO_VER}-alpine${ALPINE_VER} as golang-rocksdb

RUN apk add --no-cache \
	bash \
	binutils-gold \
	gcc \
	git \
	make \
	musl-dev \
	gflags;

ADD . $GOPATH/src/github.com/hyperledger/fabric
WORKDIR $GOPATH/src/github.com/hyperledger/fabric

FROM golang-rocksdb as tools
ARG GO_TAGS
RUN CGO_ENABLED=1 make tools GO_TAGS=${GO_TAGS}

FROM golang:${GO_VER}-alpine${ALPINE_VER}
#RUN apk add alpine-keys=2.4.0-r0
RUN apk upgrade
RUN apk add -X https://dl-cdn.alpinelinux.org/alpine/v3.16/main -u alpine-keys
# git is required to support `go list -m`
RUN apk add --no-cache \
	bash \
	git \
	jq \
	tzdata;
ENV FABRIC_CFG_PATH /etc/hyperledger/fabric
VOLUME /etc/hyperledger/fabric
COPY --from=tools /go/src/github.com/hyperledger/fabric/build/bin /usr/local/bin
COPY --from=tools /go/src/github.com/hyperledger/fabric/sampleconfig ${FABRIC_CFG_PATH}
##Copying dynamic c libraries (*.so)
COPY --from=tools /usr/local/rocksdb/lib /usr/local/rocksdb/lib
COPY --from=tools /usr/lib /usr/lib
COPY --from=tools /lib /lib
COPY --from=tools /usr/local/rocksdb /usr/local/rocksdb
COPY --from=tools /usr/include /usr/include
