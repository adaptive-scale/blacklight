#!/bin/bash

os=$(uname -s | tr '[:upper:]' '[:lower:]')
arch=$(uname -m)
echo "os=$os"
echo "arch=$arch"
echo "Detected OS: $os"
echo "Detected ARCH: $arch"

curl -o blacklight https://github.com/adaptive-scale/blacklight/releases/download/v0.1.0/blacklight_$(os)_$(arch)
chmod +x blacklight
sudo mv blacklight /usr/local/bin/

echo "Blacklight installed successfully."
echo "You can run it using the command: blacklight init"