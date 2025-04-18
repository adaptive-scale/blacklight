## Blacklight

A pluggable secret scanner that can produce results in SARIF format.

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


## Initialization
```bash
blacklight init
```

This creates a directory `~/.blacklight` and writes initial configuration file with pre-defined regexes.

## Usage

```bash
blacklight scan <path>
```

To ignore certain directories:
```bash
blacklight scan <path> --ignore=<dir1>,<dir2>
```
