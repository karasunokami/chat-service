FROM golang:1.20.3

ENV GOPATH=/go

RUN \
    go install mvdan.cc/gofumpt@v0.4.0 && \
    go install github.com/daixiang0/gci@v0.8.0 && \
    mv /go/bin/* /usr/local/bin/

workdir /app
