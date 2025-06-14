FROM golang:1.20.3

ENV GOPATH=/go

ARG GOLANGCI_LINT_VERSION

RUN \
    # fix for error with linter running "go list all"
    git config --global --add safe.directory /app && \
    # Golang ci lint tool
    wget -O- -nv https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s "${GOLANGCI_LINT_VERSION}" && \
    mv ./bin/* /usr/local/bin/ && \
    # Golang tools
    go install mvdan.cc/gofumpt@v0.4.0 && \
    go install github.com/daixiang0/gci@v0.8.0 && \
    go install github.com/kazhuravlev/options-gen/cmd/options-gen@latest && \
    go install entgo.io/ent/cmd/ent@v0.11.10 && \
    go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@master && \
    go install github.com/golang/mock/mockgen@latest && \
    go install github.com/onsi/ginkgo/v2/ginkgo@v2.9.1 && \
    go install github.com/moznion/gonstructor/cmd/gonstructor@v0.5.0 && \
    go install golang.org/x/tools/cmd/goimports@v0.7.0 && \
    mv /go/bin/* /bin

workdir /app
