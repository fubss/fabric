# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

ARG GO_VER
ARG ALPINE_VER

FROM alpine:${ALPINE_VER} as peer-base
RUN apk add --no-cache tzdata
# set up nsswitch.conf for Go's "netgo" implementation
# - https://github.com/golang/go/blob/go1.9.1/src/net/conf.go#L194-L275
# - docker run --rm debian:stretch grep '^hosts:' /etc/nsswitch.conf
RUN echo 'hosts: files dns' > /etc/nsswitch.conf

#FROM golang:${GO_VER}-alpine${ALPINE_VER} as golang
FROM fubss/rocksdb:6.27.3-golang${GO_VER}-alpine${ALPINE_VER} as golang-rocksdb
RUN apk add alpine-keys=2.4.0-r0
RUN apk upgrade
RUN apk add -X https://dl-cdn.alpinelinux.org/alpine/v3.16/main -u alpine-keys
RUN apk add --no-cache \
	bash \
	binutils-gold \
	gcc \
	git \
	make \
    gflags
ADD . $GOPATH/src/github.com/hyperledger/fabric
WORKDIR $GOPATH/src/github.com/hyperledger/fabric

FROM golang-rocksdb as peer
ARG GO_TAGS
RUN make peer GO_TAGS=${GO_TAGS}

FROM peer-base
ENV FABRIC_CFG_PATH /etc/hyperledger/fabric
VOLUME /etc/hyperledger/fabric
VOLUME /var/hyperledger
COPY --from=peer /go/src/github.com/hyperledger/fabric/build/bin /usr/local/bin
COPY --from=peer /go/src/github.com/hyperledger/fabric/sampleconfig/msp ${FABRIC_CFG_PATH}/msp
COPY --from=peer /go/src/github.com/hyperledger/fabric/sampleconfig/core.yaml ${FABRIC_CFG_PATH}
##Copying dynamic c libraries (*.so)
COPY --from=peer /usr/local/rocksdb/lib /usr/local/rocksdb/lib
COPY --from=peer /usr/lib /usr/lib
COPY --from=peer /usr/local/rocksdb /usr/local/rocksdb
COPY --from=peer /usr/include /usr/include
#==================================================================
#=======================T__E__S__T__I__N__G========================
#==================================================================
#COPY --from=golang-rocksdb /usr/local/go/ /usr/local/go/
#ENV PATH="/usr/local/go/bin:${PATH}"
#RUN apk add --no-cache \
#	gcc
#RUN echo "@testing http://nl.alpinelinux.org/alpine/edge/testing" >>/etc/apk/repositories
#RUN apk add --update --no-cache zlib zlib-dev bzip2 bzip2-dev snappy snappy-dev lz4 lz4-dev zstd@testing zstd-dev@testing libtbb-dev@testing libtbb@testing
#==================================================================
EXPOSE 7051
CMD ["peer","node","start"]
