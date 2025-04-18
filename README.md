## Blacklight

A pluggable secret scanner that can produce results in SARIF format.

Goal is to build a pluggable secret and sensitive data scanner that can be used in developer flows, CI/CD pipelines, data protection and other security tools.

## Installation

```bash
curl -Lo blacklight https://github.com/adaptive-scale/blacklight/releases/download/v0.1.1/blacklight_$(uname -s | tr '[:upper:]' '[:lower:]')_$(arch)
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
blacklight scan .
```

To ignore certain directories:
```bash
blacklight scan . --ignore=<dir1>,<dir2>
```

To ignore scan and generate SARIF report:
```bash
blacklight scan . --ignore=<dir1>,<dir2> --sarif
```

## Adding custom regexes:
You can add the custom regexes as json file inside `~/.blacklight/`. The binary automatically picks up the configuration and executes them.