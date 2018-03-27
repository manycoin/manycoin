FROM ubuntu:xenial

ENV PATH=/usr/lib/go-1.9/bin:$PATH

RUN \
  apt-get update && apt-get upgrade -q -y && \
  apt-get install -y --no-install-recommends golang-1.9 git make gcc libc-dev ca-certificates && \
  git clone --depth 1 --branch release/1.8 https://github.com/okcoin/go-okcoin && \
  (cd go-okcoin && make gokc) && \
  cp go-okcoin/build/bin/gokc /gokc && \
  apt-get remove -y golang-1.9 git make gcc libc-dev && apt autoremove -y && apt-get clean && \
  rm -rf /go-okcoin

EXPOSE 8545
EXPOSE 30303

ENTRYPOINT ["/gokc"]
