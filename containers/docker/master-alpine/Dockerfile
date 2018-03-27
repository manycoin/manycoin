FROM alpine:3.7

RUN \
  apk add --update go git make gcc musl-dev linux-headers ca-certificates && \
  git clone --depth 1 --branch release/1.8 https://github.com/okcoin/go-okcoin && \
  (cd go-okcoin && make gokc) && \
  cp go-okcoin/build/bin/gokc /gokc && \
  apk del go git make gcc musl-dev linux-headers && \
  rm -rf /go-okcoin && rm -rf /var/cache/apk/*

EXPOSE 8545
EXPOSE 30303

ENTRYPOINT ["/gokc"]
