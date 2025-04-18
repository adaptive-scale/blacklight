/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"blacklight/internal/model"
	"encoding/json"
	"fmt"
	"os"
	"path"

	"github.com/spf13/cobra"
)

var preConfig = []model.Configuration{
	{Name: "AWS Access Key", Regex: `AKIA[0-9A-Z]{16}`, Severity: 3},
	{Name: "AWS Secret Key", Regex: `(?i)aws(.{0,20})?['"][0-9a-zA-Z/+]{40}['"]`, Severity: 3},
	{Name: "Google API Key", Regex: `AIza[0-9A-Za-z\\-_]{35}`, Severity: 3},
	{Name: "Slack Token", Regex: `xox[baprs]-[0-9a-zA-Z]{10,48}`, Severity: 3},
	{Name: "Stripe API Key", Regex: `sk_live_[0-9a-zA-Z]{24}`, Severity: 3},
	{Name: "GitHub Token", Regex: `gh[pousr]_[0-9a-zA-Z]{36,}`, Severity: 3},
	{Name: "Heroku API Key", Regex: `[hH]eroku(.{0,20})?['"][0-9a-fA-F]{32}['"]`, Severity: 3},
	{Name: "Generic API Key", Regex: `(?i)api[_-]?key(.{0,20})?['"][0-9a-zA-Z]{32,45}['"]`, Severity: 2},
	{Name: "Generic Secret", Regex: `(?i)secret(.{0,20})?['"][0-9a-zA-Z/+]{32,45}['"]`, Severity: 2},
	{Name: "Private RSA Key", Regex: `-----BEGIN RSA PRIVATE KEY-----`, Severity: 3},
	{Name: "Private DSA Key", Regex: `-----BEGIN DSA PRIVATE KEY-----`, Severity: 3},
	{Name: "Private EC Key", Regex: `-----BEGIN EC PRIVATE KEY-----`, Severity: 3},
	{Name: "PGP Private Key Block", Regex: `-----BEGIN PGP PRIVATE KEY BLOCK-----`, Severity: 3},
	{Name: "OpenSSH Private Key", Regex: `-----BEGIN OPENSSH PRIVATE KEY-----`, Severity: 3},
	{Name: "JWT", Regex: `eyJ[A-Za-z0-9-_]+\.[A-Za-z0-9-_]+\.[A-Za-z0-9-_]+`, Severity: 2},
	{Name: "Facebook Access Token", Regex: `EAACEdEose0cBA[0-9A-Za-z]+`, Severity: 3},
	{Name: "Basic Auth in URL", Regex: `https?:\/\/[^:\/\s]+:[^@\/\s]+@`, Severity: 2},
	{Name: "Google OAuth Token", Regex: `ya29\.[0-9A-Za-z\-_]+`, Severity: 3},
	{Name: "Password in Variable", Regex: `(?i)password\s*=\s*['"][^'"]+['"]`, Severity: 2},
	{Name: "Password in URL", Regex: `(?i)://[^:]+:[^@]+@`, Severity: 2},
	{Name: "RSA Public Key", Regex: `-----BEGIN PUBLIC KEY-----`, Severity: 1},
	{Name: "Azure Storage Key", Regex: `(?i)azure(.{0,20})?['"][0-9a-zA-Z/+]{88}['"]`, Severity: 3},
	{Name: "Twilio API Key", Regex: `SK[0-9a-fA-F]{32}`, Severity: 3},
	{Name: "SendGrid API Key", Regex: `SG\.[A-Za-z0-9\-_]{22}\.[A-Za-z0-9\-_]{43}`, Severity: 3},
	{Name: "Mailchimp API Key", Regex: `[0-9a-f]{32}-us[0-9]{1,2}`, Severity: 3},
	{Name: "Square Access Token", Regex: `sq0atp-[0-9A-Za-z\-_]{22}`, Severity: 3},
	{Name: "Square OAuth Secret", Regex: `sq0csp-[0-9A-Za-z\-_]{43}`, Severity: 3},
	{Name: "PayPal Braintree Token", Regex: `access_token\$production\$[0-9a-z]{16}\$[0-9a-f]{32}`, Severity: 3},
	{Name: "LDAP Password", Regex: `(?i)ldap(.{0,20})?['"][^'"]{8,}['"]`, Severity: 2},
	{Name: "MongoDB URI with credentials", Regex: `mongodb(\+srv)?:\/\/[^:]+:[^@]+@`, Severity: 2},
	{Name: "MySQL URI with credentials", Regex: `mysql:\/\/[^:]+:[^@]+@`, Severity: 2},
	{Name: "PostgreSQL URI with credentials", Regex: `postgresql:\/\/[^:]+:[^@]+@`, Severity: 2},
	{Name: "Docker Hub Token", Regex: `(?i)docker(.{0,20})?['"][A-Za-z0-9_\-]{20,}['"]`, Severity: 2},
	{Name: "Stripe Test Key", Regex: `sk_test_[0-9a-zA-Z]{24}`, Severity: 1},
	{Name: "Square Sandbox Token", Regex: `sq0atb-[0-9A-Za-z\-_]{22}`, Severity: 1},
	{Name: "Generic Bearer Token", Regex: `(?i)Bearer\s+[A-Za-z0-9\-._~+/]+=*`, Severity: 2},
	{Name: "JSON Web Token Header", Regex: `"alg"\s*:\s*"HS256"`, Severity: 1},
	{Name: "Firebase API Key", Regex: `AIza[0-9A-Za-z\\-_]{35}`, Severity: 3},
	{Name: "Algolia API Key", Regex: `(?i)algolia(.{0,20})?['"][a-z0-9]{32}['"]`, Severity: 2},
	{Name: "Twitch API Key", Regex: `twitch[_-]?api[_-]?key['"]?\s*[:=]\s*['"][a-z0-9]{30,}['"]`, Severity: 3},
	{Name: "Discord Token", Regex: `mfa\.[0-9a-z\-_]{84}`, Severity: 3},
	{Name: "Firebase Auth Domain", Regex: `.*\.firebaseapp\.com`, Severity: 1},
	{Name: "Basic Auth Header", Regex: `Authorization:\s*Basic\s+[A-Za-z0-9+/=]+`, Severity: 2},
	{Name: "Authorization Bearer", Regex: `Authorization:\s*Bearer\s+[A-Za-z0-9\-._~+/]+=*`, Severity: 2},
	{Name: "Generic Token", Regex: `(?i)token\s*=\s*['"][a-zA-Z0-9_\-]{20,}['"]`, Severity: 2},
	{Name: "Generic Username/Password Pair", Regex: `"username"\s*:\s*".+?",\s*"password"\s*:\s*".+?"`, Severity: 2},
	{Name: "Secret in ENV", Regex: `(?i)(SECRET|TOKEN|KEY|PWD|PASS)[A-Z0-9_]*=.+`, Severity: 2},
	{Name: "Private Key JSON", Regex: `"private_key"\s*:\s*"-----BEGIN PRIVATE KEY-----`, Severity: 3},
	{Name: "OAuth Client Secret", Regex: `"client_secret"\s*:\s*"[^"]+"`, Severity: 3},
	{Name: "OAuth Client ID", Regex: `"client_id"\s*:\s*"[^"]+"`, Severity: 2},
	{Name: "Shopify Access Token", Regex: `shpat_[a-fA-F0-9]{32}`, Severity: 3},
	{Name: "Shopify Shared Secret", Regex: `shpss_[a-fA-F0-9]{32}`, Severity: 3},
	{Name: "Shopify Custom App Token", Regex: `shpca_[a-fA-F0-9]{32}`, Severity: 3},
	{Name: "Telegram Bot Token", Regex: `[0-9]+:AA[0-9A-Za-z\-_]{33}`, Severity: 3},
	{Name: "Bitbucket OAuth Token", Regex: `(?i)bitbucket(.{0,20})?['"][a-z0-9]{32,40}['"]`, Severity: 2},
	{Name: "Dropbox API Key", Regex: `sl\.[A-Za-z0-9-_]{15,}`, Severity: 3},
	{Name: "NPM Token", Regex: `(?i)_authToken\s*=\s*['"]?[a-z0-9\-]{36}['"]?`, Severity: 3},
	{Name: "Yandex API Key", Regex: `AQVN[A-Za-z0-9\-_]{35,}`, Severity: 2},
	{Name: "GitLab Personal Token", Regex: `glpat-[0-9a-zA-Z\-_]{20,}`, Severity: 3},
	{Name: "Confluent Cloud API Key", Regex: `(?i)confluent(.{0,20})?['"][a-zA-Z0-9]{16,}['"]`, Severity: 2},
	{Name: "Rollbar Token", Regex: `(?i)rollbar(.{0,20})?['"][a-f0-9]{32}['"]`, Severity: 2},
	{Name: "Honeycomb API Key", Regex: `(?i)honeycomb(.{0,20})?['"][a-z0-9]{32}['"]`, Severity: 2},
	{Name: "Cloudinary API Key", Regex: `(?i)cloudinary(.{0,20})?['"][0-9]{15}['"]`, Severity: 2},
	{Name: "NewRelic License Key", Regex: `NRAK-[A-Za-z0-9]{27}`, Severity: 3},
	{Name: "Postman API Key", Regex: `PMAK-[a-f0-9]{24,32}-[a-f0-9]{24,32}`, Severity: 3},
	{Name: "Asana Personal Access Token", Regex: `(?i)asana(.{0,20})?['"][0-9a-z]{32}['"]`, Severity: 2},
	{Name: "Google Cloud Service Account", Regex: `"type": "service_account",\s*"project_id"`, Severity: 2},
	{Name: "Zendesk Token", Regex: `(?i)zendesk(.{0,20})?['"][a-z0-9]{32,45}['"]`, Severity: 2},
	{Name: "CircleCI Token", Regex: `circleci_token['"]?\s*[:=]\s*['"][0-9a-f]{40}['"]`, Severity: 2},
	{Name: "Netlify Token", Regex: `(?i)netlify(.{0,20})?['"][a-z0-9]{40}['"]`, Severity: 2},
	{Name: "GitHub App Private Key", Regex: `-----BEGIN RSA PRIVATE KEY-----[\s\S]+?-----END RSA PRIVATE KEY-----`, Severity: 3},
	{Name: "Azure Client Secret", Regex: `"clientSecret"\s*:\s*"[^"]+"`, Severity: 3},
	{Name: "Azure Client ID", Regex: `"clientId"\s*:\s*"[^"]+"`, Severity: 2},
	{Name: "Azure Tenant ID", Regex: `"tenantId"\s*:\s*"[^"]+"`, Severity: 2},
	{Name: "Fastly API Token", Regex: `(?i)fastly(.{0,20})?['"][A-Za-z0-9\-_=]{32,}['"]`, Severity: 3},
	{Name: "Bitly Generic Access Token", Regex: `(?i)bitly(.{0,20})?['"][a-z0-9]{32,}['"]`, Severity: 2},
	{Name: "OpenAI API Key", Regex: `sk-[A-Za-z0-9]{32,}`, Severity: 3},
	{Name: "Firebase Web API Key", Regex: `AIza[0-9A-Za-z\-_]{35}`, Severity: 2},
	{Name: "ReCaptcha Site Key", Regex: `6L[0-9A-Za-z_-]{38}`, Severity: 1},
	{Name: "ReCaptcha Secret Key", Regex: `6L[0-9A-Za-z_-]{38}`, Severity: 3},
	{Name: "OAuth Redirect URI", Regex: `https?:\/\/[^\s]+\/oauth\/callback`, Severity: 1},
	{Name: "Generic OAuth Token", Regex: `(?i)oauth(_token)?['"]?\s*[:=]\s*['"][a-z0-9\-_]{20,}['"]`, Severity: 2},
	{Name: "Kafka Connection String", Regex: `(?i)kafka(.{0,20})?(\/\/)?[^:]+:[^@]+@`, Severity: 2},
	{Name: "RabbitMQ URI with Credentials", Regex: `amqp:\/\/[^:]+:[^@]+@`, Severity: 2},
	{Name: "Basic Auth Credentials", Regex: `(?i)(user(name)?|login)[=: ]['"][^'"]+['"]\s*(password|pass)[=: ]['"][^'"]+['"]`, Severity: 2},
	{Name: "DigitalOcean Token", Regex: `dop_v1_[a-f0-9]{64}`, Severity: 3},
	{Name: "Clojars API Token", Regex: `(?i)clojars(.{0,20})?['"][a-z0-9]{40}['"]`, Severity: 2},
	{Name: "Artifactory Token", Regex: `(?i)artifactory(.{0,20})?['"][a-z0-9-_]{32,}['"]`, Severity: 2},
	{Name: "Ansible Vault Encrypted Block", Regex: `\$ANSIBLE_VAULT;.*\n[0-9a-f\n]+`, Severity: 3},
	{Name: "Generic Client Secret", Regex: `client_secret\s*=\s*['"][^'"]+['"]`, Severity: 3},
	{Name: "Generic Access Token", Regex: `access_token\s*=\s*['"][^'"]+['"]`, Severity: 3},
	{Name: "Generic Refresh Token", Regex: `refresh_token\s*=\s*['"][^'"]+['"]`, Severity: 2},
	{Name: "Email + Password in Log", Regex: `[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,4}:.+`, Severity: 2},
	{Name: "Firebase Admin SDK", Regex: `"project_id": ".+?",\s*"private_key_id": ".+?"`, Severity: 3},
	{Name: "Pusher Secret Key", Regex: `(?i)pusher(.{0,20})?['"][a-z0-9]{32}['"]`, Severity: 2},
	{Name: "RDS Endpoint", Regex: `[a-zA-Z0-9\-]+\.rds\.amazonaws\.com`, Severity: 1},
	{Name: "Snowflake Credential", Regex: `snowflake:\/\/[^:]+:[^@]+@`, Severity: 2},
	{Name: "Cloudflare API Token", Regex: `(?i)cloudflare(.{0,20})?['"][a-z0-9_\-]{40,}['"]`, Severity: 3},
	{Name: "Environment Variable Export", Regex: `(?i)export\s+(SECRET|TOKEN|KEY|PWD|PASS)[A-Z0-9_]*=.+`, Severity: 2},
	{Name: "PEM Encoded Private Key", Regex: `-----BEGIN PRIVATE KEY-----[\s\S]+?-----END PRIVATE KEY-----`, Severity: 3},
	{Name: "Basic Shell Credential", Regex: `(?i)(token|secret|password|pwd|key)=['"]?[A-Za-z0-9+/=]{8,}['"]?`, Severity: 2},
	{Name: "Terraform Variable with Secret", Regex: `variable\s+".*(_secret|_key|_token)"`, Severity: 2},
	{Name: "Kubernetes Secret YAML", Regex: `apiVersion: v1\s+kind: Secret`, Severity: 2},
	//{Name: "Sensitive JSON Key", Regex: `"?(auth|token|secret|key|pass|pwd)"?\s*:\s*"[^"]{8,}"`, Severity: 2},
}

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "generates the initial configuration",
	Long: `
Run:

blacklight init

Should generate all configuration and scanner regexes.
`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("initializing configuration")

		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Println("Error getting home directory:", err)
			return
		}

		if err := os.MkdirAll(path.Join(home, ".blacklight"), os.FileMode(0755)); err != nil {
			fmt.Println("Error creating directory:", err)
			return
		}

		join := path.Join(home, ".blacklight", "config.json")

		d, err := json.Marshal(preConfig)
		if err != nil {
			fmt.Println("Error marshalling configuration:", err)
			return
		}

		err = os.WriteFile(join, d, os.FileMode(0755))
		if err != nil {
			fmt.Println("Error reading directory:", err)
			return
		}

		fmt.Println("=> created configuration " + join)

	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// initCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// initCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
