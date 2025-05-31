# Blacklight
<img width="438" alt="Screenshot 2025-05-31 at 02 16 47" src="https://github.com/user-attachments/assets/6e8e0841-b0d4-45d7-a704-93b0d333b7a2" />

Blacklight is a powerful secret, keys and sensitive data scanning tool that helps you detect and prevent sensitive information leaks in your codebase, databases, cloud storage, and communication platforms.

## Features

- **Multi-Source Scanning**
  - Local files and directories
  - Databases (PostgreSQL, MySQL)
  - AWS S3 buckets
  - Slack workspace messages and files
  - Cloud Storage (Google Drive, Dropbox)
  - Git repositories

- **Advanced Detection**
  - Pattern-based secret detection
  - Context-aware scanning
  - Multi-language support
  - Configurable severity levels
  - Rule categorization
  - Smart file format detection

- **Supported File Formats**
  - Plain text files
  - JSON files (with nested object support)
  - YAML files (with nested object support)
  - XML files (with attribute scanning)
  - INI/Config files
  - Environment files (.env)
  - Configuration files

- **User Experience**
  - Cross-platform compatibility (Windows, Linux, macOS)
  - Beautiful table output with go-pretty formatting
  - Color-coded severity indicators
  - Detailed violation reporting
  - Rich context for findings

## Installation

### Quick Install (Linux/macOS)

```bash
# macOS (Apple Silicon)
curl -Lo blacklight https://github.com/adaptive-scale/blacklight/releases/latest/download/blacklight-darwin-arm64 && \
chmod +x blacklight && \
sudo mv blacklight /usr/local/bin/

# macOS (Intel)
curl -Lo blacklight https://github.com/adaptive-scale/blacklight/releases/latest/download/blacklight-darwin-amd64 && \
chmod +x blacklight && \
sudo mv blacklight /usr/local/bin/

# Linux (x86_64)
curl -Lo blacklight https://github.com/adaptive-scale/blacklight/releases/latest/download/blacklight-linux-amd64 && \
chmod +x blacklight && \
sudo mv blacklight /usr/local/bin/

# Linux (ARM64)
curl -Lo blacklight https://github.com/adaptive-scale/blacklight/releases/latest/download/blacklight-linux-arm64 && \
chmod +x blacklight && \
sudo mv blacklight /usr/local/bin/

# Verify installation
blacklight version
```

### Using Go

```bash
# Using go install
go install github.com/adaptive-scale/blacklight@latest

# Or clone and build
git clone https://github.com/adaptive-scale/blacklight.git
cd blacklight
make build
```

### Docker

```bash
# Pull the latest image
docker pull adaptivescale/blacklight:latest

# Run a scan
docker run -v $(pwd):/workspace adaptivescale/blacklight:latest scan /workspace
```

### Windows

Download the appropriate `.exe` file from our [releases page](https://github.com/adaptive-scale/blacklight/releases) and add it to your PATH.

### Verifying the Installation

After installation, verify that Blacklight is working correctly:

```bash
# Check version
blacklight version

# View help
blacklight --help

# Run a test scan
blacklight scan --help
```

## Usage

### Basic Usage

```bash
# Scan a directory
blacklight scan /path/to/directory

# Scan with verbose output
blacklight scan /path/to/directory --verbose

# Scan a database
blacklight scan --db "postgresql://user:pass@localhost:5432/dbname"

# Scan an S3 bucket
blacklight scan --s3 "s3://bucket-name"

# Scan cloud storage
blacklight scan --drive "gdrive://folder-id"
```

### Rule Management

```bash
# List all rules
blacklight rules list

# List rules by type
blacklight rules list --type cloud

# List rules by severity
blacklight rules list --severity 3

# Add a new rule
blacklight rules add --name "Custom API Key" \
                     --regex "api_key_[a-zA-Z0-9]{32}" \
                     --severity 2 \
                     --type "secret"
```

### Slack Workspace Scanning

Blacklight includes a powerful Slack scanner that can detect secrets and sensitive information in:
- Channel messages (public and private)
- Message threads
- Direct messages (DMs)
- Group messages
- Shared files
- File comments

#### Setup

1. Create a Slack App at https://api.slack.com/apps
2. Add the following OAuth scopes:
   ```
   channels:history   - View messages and other content in public channels
   channels:read     - View basic information about public channels
   files:read       - View files shared in channels and conversations
   groups:history   - View messages and other content in private channels
   groups:read      - View basic information about private channels
   im:history      - View messages and other content in direct messages
   im:read        - View basic information about direct messages
   mpim:history   - View messages and other content in group direct messages
   mpim:read     - View basic information about group direct messages
   ```
3. Install the app to your workspace
4. Copy the Bot User OAuth Token (starts with `xoxb-`)

#### Usage

```bash
# Basic scan of all accessible channels
blacklight slack --token xoxb-your-token

# Scan specific channels
blacklight slack --token xoxb-your-token --channels C01234567,C89012345

# Scan recent messages
blacklight slack --token xoxb-your-token --days 7

# Full scan including threads and files
blacklight slack --token xoxb-your-token --include-threads --include-files
```

#### Configuration Options

| Option | Description | Default |
|--------|-------------|---------|
| `--token` | Slack Bot User OAuth Token (required) | - |
| `--channels` | Comma-separated list of channel IDs | All accessible |
| `--days` | Number of days of history to scan | 30 |
| `--include-threads` | Scan message threads | false |
| `--include-files` | Scan file contents | false |
| `--exclude-archived` | Skip archived channels | true |

#### Performance Considerations

- File scanning is disabled by default to improve performance
- Files larger than 10MB are skipped
- Use the `--days` flag to limit the scan window
- Specify channels to scan for faster results

## Cloud Storage Scanning

Blacklight can scan files in various cloud storage services for secrets and sensitive information:

### Implemented Providers
- **Google Drive** (`gdrive://`)
  - Scans files in specified folders
  - Supports file content analysis
  - Respects file size limits
  - OAuth2 authentication
  - Automatic file format detection
  - Recursive folder scanning

- **Dropbox** (`dropbox://`)
  - Full folder scanning
  - File content analysis
  - Path-based access
  - Access token authentication
  - Smart file format handling
  - Size-based file filtering

### Coming Soon
- **OneDrive** (`onedrive://`) - In development
- **Box** (`box://`) - Planned

### Authentication

Each provider requires appropriate authentication:

```bash
# Google Drive - OAuth2 client configuration
export CLOUD_TOKEN='{"client_id":"...","client_secret":"...","redirect_uris":["..."]}'

# Dropbox - Access Token
export CLOUD_TOKEN="your-dropbox-access-token"
```

### Usage Examples

```bash
# Scan Google Drive folder
blacklight scan --drive "gdrive://folder-id"

# Scan Dropbox folder
blacklight scan --drive "dropbox://path/to/folder"

# Include shared files (Google Drive)
blacklight scan --drive "gdrive://folder-id" --include-shared

# Limit scan history
blacklight scan --drive "dropbox://folder" --days 7

# Adjust file size limit
blacklight scan --drive "gdrive://folder-id" --max-size 5242880  # 5MB
```

### Configuration Options

| Option | Description | Default |
|--------|-------------|---------|
| `--drive, -r` | Cloud storage URL to scan | - |
| `--include-shared` | Include shared files | false |
| `--days` | Days of history to scan | 30 |
| `--max-size` | Maximum file size (bytes) | 10MB |

### File Format Support

The cloud storage scanner automatically detects and processes various file formats:

| Format | Extensions | Detection |
|--------|------------|-----------|
| JSON | .json | Extension + Content |
| YAML | .yaml, .yml | Extension + Content |
| XML | .xml | Extension |
| INI | .ini, .conf, .config | Extension |
| ENV | .env | Extension |
| Text | others | Default |

### Performance Considerations

- Files larger than the max-size limit are skipped
- Use `--days` to limit scan scope
- Specify precise folder paths for faster scans
- Token expiration is handled automatically
- File format detection optimizes scanning

### Security Notes

- Tokens should be kept secure and not shared
- Use read-only access tokens when possible
- Consider using environment variables for token storage
- Regularly rotate access tokens
- Ensure proper access permissions

## Rule Types

Blacklight organizes its scanning rules into the following categories:

### Authentication & Authorization
- `auth`: Authentication tokens, passwords, OAuth credentials
- `key`: Cryptographic keys (RSA, DSA, PGP, SSH)

### Cloud & Infrastructure
- `cloud`: Cloud provider credentials (AWS, Azure, GCP)
- `container`: Container platform secrets (Docker, Kubernetes)
- `iac`: Infrastructure as Code secrets (Terraform)
- `cdn`: Content Delivery Network tokens

### APIs & Services
- `api`: Generic and service-specific API keys
- `monitoring`: Monitoring service tokens (NewRelic, Rollbar)
- `ci`: CI/CD platform credentials
- `vcs`: Version Control System tokens (GitHub, GitLab)

### Payment & Financial
- `payment`: Payment gateway credentials
- `pci`: Payment Card Industry data
- `ecommerce`: E-commerce platform tokens

### Data & Storage
- `database`: Database credentials and endpoints
- `messaging`: Message queue credentials
- `package`: Package registry tokens

### Other
- `secret`: Generic secrets and environment variables
- `social`: Social media platform tokens
- `security`: Security-related credentials
- `config`: Configuration file secrets
- `ai`: AI service credentials

## Rule Configuration

Rules are stored in `~/.blacklight/rules.yaml`. Each rule has the following properties:

| Property | Description | Required |
|----------|-------------|----------|
| `id` | Unique identifier | Yes |
| `name` | Human-readable name | Yes |
| `description` | What the rule detects | No |
| `regex` | Detection pattern | Yes |
| `severity` | 1 (low) to 3 (high) | Yes |
| `type` | Category from above | Yes |
| `disabled` | Skip this rule | No |

### Example Rule File

```yaml
- id: "aws_access_key"
  name: "AWS Access Key"
  description: "Amazon Web Services access key ID"
  regex: "AKIA[0-9A-Z]{16}"
  severity: 3
  type: "cloud"
  disabled: false

- id: "stripe_key"
  name: "Stripe API Key"
  description: "Stripe secret API key"
  regex: "sk_live_[0-9a-zA-Z]{24}"
  severity: 3
  type: "payment"
  disabled: false
```

## Output Format

Blacklight provides rich, color-coded output:

```
[Severity 3]: AWS Access Key Found
Location: slack://channel/C0123456/message/1234567890.123
Context: ...config = { accessKeyId: "AKIAXXXXXXXXXXXXXXXX", region: "us-east-1" }...
Match: AKIAXXXXXXXXXXXXXXXX
--------------------------------------------------------------------------------
```

The table output uses go-pretty for enhanced readability:

```
╭──────────────────────────┬──────────┬──────────┬─────────┬───────────────────────────────────╮
│ NAME                     │ TYPE     │ SEVERITY │ STATUS  │ PATTERN                           │
├──────────────────────────┼──────────┼──────────┼─────────┼───────────────────────────────────┤
│ AWS Access Key           │ cloud    │    3     │ Enabled │ AKIA[0-9A-Z]{16}                 │
│ Stripe API Key          │ payment  │    3     │ Enabled │ sk_live_[0-9a-zA-Z]{24}          │
╰──────────────────────────┴──────────┴──────────┴─────────┴───────────────────────────────────╯
```

## Default Rules

Blacklight comes with a comprehensive set of pre-configured rules for detecting various types of secrets and sensitive information:

### Authentication & Authorization
| Rule Name | Description | Severity |
|-----------|-------------|----------|
| AWS Access Key | Amazon Web Services access key ID | High |
| AWS Secret Key | Amazon Web Services secret access key | High |
| Generic API Key | Generic API key patterns | Medium |
| JWT Token | JSON Web Token | Medium |
| Basic Auth | Basic Authentication credentials | High |
| OAuth Client Secret | OAuth 2.0 client secret | High |
| OAuth Access Token | OAuth 2.0 access token | High |
| SSH Private Key | SSH private key content | High |

### Cloud Services
| Rule Name | Description | Severity |
|-----------|-------------|----------|
| Azure Storage Account Key | Azure storage account access key | High |
| GCP Service Account | Google Cloud Platform service account key | High |
| Firebase Database URL | Firebase realtime database URL | Medium |
| Heroku API Key | Heroku platform API key | High |
| Digital Ocean Token | DigitalOcean API token | High |
| Cloudflare API Key | Cloudflare API key and token | High |
| AWS Session Token | AWS temporary session token | High |

### Payment & Financial
| Rule Name | Description | Severity |
|-----------|-------------|----------|
| Stripe API Key | Stripe secret API key | High |
| Stripe Restricted Key | Stripe restricted API key | High |
| PayPal Access Token | PayPal OAuth2 access token | High |
| Square Access Token | Square OAuth2 access token | High |
| Credit Card Number | Credit card number patterns | High |

### Database & Storage
| Rule Name | Description | Severity |
|-----------|-------------|----------|
| PostgreSQL Connection | PostgreSQL connection string | High |
| MySQL Connection | MySQL connection string | High |
| MongoDB Connection | MongoDB connection URI | High |
| Redis Connection | Redis connection string | High |
| Elasticsearch Connection | Elasticsearch connection string | Medium |

### Communication & Messaging
| Rule Name | Description | Severity |
|-----------|-------------|----------|
| Slack Token | Slack API token and webhook URL | High |
| Slack Webhook | Slack incoming webhook URL | Medium |
| Discord Token | Discord bot token | High |
| Discord Webhook | Discord webhook URL | Medium |
| Twilio API Key | Twilio API key | High |
| SendGrid API Key | SendGrid API key | High |

### Development & CI/CD
| Rule Name | Description | Severity |
|-----------|-------------|----------|
| GitHub Token | GitHub personal access token | High |
| GitLab Token | GitLab personal access token | High |
| NPM Token | NPM authentication token | Medium |
| Docker Registry | Docker registry credentials | Medium |
| CircleCI Token | CircleCI API token | High |
| Jenkins Token | Jenkins API token | High |

### AI & Machine Learning
| Rule Name | Description | Severity |
|-----------|-------------|----------|
| OpenAI API Key | OpenAI API key | High |
| Hugging Face Token | Hugging Face API token | High |
| Anthropic API Key | Anthropic API key | High |
| Cohere API Key | Cohere API key | High |
| Replicate API Token | Replicate API token | High |

### Analytics & Monitoring
| Rule Name | Description | Severity |
|-----------|-------------|----------|
| Google Analytics | Google Analytics API key | Medium |
| New Relic Key | New Relic license key | High |
| Datadog API Key | Datadog API key | High |
| Sentry DSN | Sentry client key | Medium |
| Mixpanel Token | Mixpanel project token | Medium |

### Generic Patterns
| Rule Name | Description | Severity |
|-----------|-------------|----------|
| Private Key | Generic private key content | High |
| Password in URL | Password in URL parameters | High |
| Environment Variable | Hardcoded environment variables | Medium |
| IP Address | Private IP addresses | Low |
| Internal Path | Internal system paths | Low |

### Custom Rules

You can add your own custom rules by creating a file at `~/.blacklight/rules.yaml`:

```yaml
- id: "custom_api_key"
  name: "Custom API Key"
  description: "Detects custom API key pattern"
  regex: "myapi_[a-zA-Z0-9]{32}"
  severity: 2
  type: "api"
  disabled: false

- id: "internal_token"
  name: "Internal Service Token"
  description: "Internal service authentication token"
  regex: "int_[a-zA-Z0-9]{24}"
  severity: 3
  type: "auth"
  disabled: false
```

### Rule Properties

| Property | Description | Required | Example |
|----------|-------------|----------|---------|
| `id` | Unique identifier | Yes | `aws_access_key` |
| `name` | Human-readable name | Yes | "AWS Access Key" |
| `description` | What the rule detects | No | "Amazon Web Services access key ID" |
| `regex` | Detection pattern | Yes | `AKIA[0-9A-Z]{16}` |
| `severity` | 1 (low) to 3 (high) | Yes | 3 |
| `type` | Category from types list | Yes | "cloud" |
| `disabled` | Skip this rule | No | false |

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

Copyright © 2025 Debarshi Basak 

Licensed under the Apache License, Version 2.0
