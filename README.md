## Blacklight

A pluggable secret scanner.


## Installation

```bash
curl -Lo blacklight https://github.com/adaptive-scale/blacklight/releases/download/v0.1.0/blacklight_$(uname -s | tr '[:upper:]' '[:lower:]')_$(arch)
chmod +x blacklight
sudo mv blacklight /usr/local/bin/
```

or 

```bash
curl https://raw.githubusercontent.com/adaptive-scale/blacklight/refs/heads/master/install.sh | sudo bash
```

```bash

blacklight init

```