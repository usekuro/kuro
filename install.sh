#!/bin/bash
set -e

echo '🚀 Installing UseKuro...'

OS=$(uname -s)
ARCH=$(uname -m)

case $ARCH in
    x86_64) ARCH='amd64';;
    aarch64|arm64) ARCH='arm64';;
esac

case $OS in
    Linux*) BINARY='usekuro-linux-${ARCH}'; INSTALL_DIR='/usr/local/bin';;
    Darwin*) BINARY='usekuro-darwin-${ARCH}'; INSTALL_DIR='/usr/local/bin';;
    *) echo 'Unsupported OS: $OS'; exit 1;;
esac

REPO='usekuro/kuro'
URL='https://github.com/${REPO}/releases/latest/download/${BINARY}'

echo 'Downloading ${BINARY}...'
mkdir -p $INSTALL_DIR

if command -v curl >/dev/null 2>&1; then
    curl -fsSL $URL -o $INSTALL_DIR/usekuro || {
        echo 'Download failed, building from source...'
        git clone https://github.com/${REPO}.git /tmp/usekuro-build
        cd /tmp/usekuro-build
        go build -o $INSTALL_DIR/usekuro cmd/usekuro/main.go
        cd - && rm -rf /tmp/usekuro-build
    }
else
    echo 'Building from source...'
    git clone https://github.com/${REPO}.git /tmp/usekuro-build
    cd /tmp/usekuro-build  
    go build -o $INSTALL_DIR/usekuro cmd/usekuro/main.go
    cd - && rm -rf /tmp/usekuro-build
fi

chmod +x $INSTALL_DIR/usekuro
echo '✅ UseKuro installed successfully!'
echo 'Run: usekuro web'
