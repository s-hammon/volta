#!/bin/bash

# Install Goose
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

if [ "$ARCH" = "aarch64" ]; then
    ARCH="arm64"
fi

if [ $# -eq 0 ]; then
    GOOSE_URI="https://github.com/pressly/goose/releases/download/v3.24.3/goose_${OS}_${ARCH}"
else
    GOOSE_URI="https://github.com/pressly/goose/releases/download/${1}/goose_${OS}_${ARCH}"
fi

GOOSE_PATH="${GOOSE_INSTALL:-$HOME}"
BIN_DIR="${GOOSE_PATH}/bin"
GOOSE_EXE="${BIN_DIR}/goose"

if [ ! -d "${BIN_DIR}" ]; then
    mkdir -p "{BIN_DIR}"
fi

curl --silent --show-error --location --fail --location --output "${GOOSE_EXE}" "${GOOSE_URI}"
chmod +x "${GOOSE_EXE}"

echo "Goose installed successfully to ${GOOSE_EXE}"
if command -v goose >/dev/null; then
    echo "Run 'goose --help' to get started"
fi

# Install sqlc
go install github.com/pressly/goose/v3/cmd/goose@latest
# Install golangci-lint
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.1.6
echo "golangci-lint installed successfully"
# Install gosec
curl -sfL https://raw.githubusercontent.com/securego/gosec/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.22.4
