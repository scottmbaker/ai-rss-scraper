# AI RSS Scraper

AI RSS Scraper is a tool designed to watch RSS feeds and use AI to determine which articles are of 
interest to you (or more specifically to me! I wrote it for myself...). It does the following:

* Downloads articles from an RSS feed
* Stores them in a local SQLite database
* Uses an AI model to score the articles based on your preferences
* Sends you an email with the top articles

The tool may be run from a command line as a one time job, or deployed persistently in a container 
using docker and/or kubernetes.

To use this, you will need to have an OpenAI provider and provide an API key.  You will also need to 
provide a prompt that tells the AI what you are interested in.  Finally, you will need to provide a 
list of RSS feeds to scrape.  You can provide these on the command line or in a config file. The defaults
are:

* OpenAI Provider: poe.com
* Model: gemini-3-flash
* Prompt: stuff related to scott's vintage computing hobby
* RSS Feed: Hackaday


## Installation

### Prerequisites

-   Go 1.25.5
-   An API key for an OpenAI-compatible service (e.g., Poe, OpenAI).
-   A SMTP server to send emails, if you want to use the email reporting feature. 
    Writing the output to an html file is an alternative.

### Build

```bash
git clone https://github.com/scottmbaker/ai-rss-scraper.git
cd ai-rss-scraper
make build
```

## Configuration

The application can be configured via a YAML config file (`$HOME/.ai-rss-scraper.yaml`), 
environment variables, or CLI flags.

The config file has the least precedence, followed by environment variables, and then CLI flags.

### Environment Variables

-   `API_KEY`: Your API key.
-   `BASE_URL`: OpenAI-compatible API URL.
-   `MODEL`: LLM Model to use.
-   `FEED_URL`: RSS Feed URL.
-   `PROMPT`: Custom prompt string or path to a file (prefixed with `@`).
-   `EMAIL_SMARTHOST`: SMTP server address.
-   `EMAIL_IDENTITY`: SMTP auth identity.
-   `EMAIL_TO`: Recipient address.
-   `EMAIL_FROM`: Sender address.
-   `EMAIL_SUBJECT`: Email subject.
-   `EMAIL_USERNAME`: SMTP username.
-   `EMAIL_PASSWORD`: SMTP password.

### Flags

Global flags available for all commands:

-   `--api-key`: API Key (overrides env var).
-   `--base-url`: OpenAI-compatible API URL (default: `https://api.poe.com/v1`).
-   `--model`: LLM Model to use (default: `gemini-3-flash`).
-   `--feed-url`: RSS Feed URL (default: `https://hackaday.com/blog/feed/`).
-   `--db-path`: Path to SQLite database (default: `rss_history.db`).
-   `--prompt`: Custom prompt string or path to a file (prefixed with `@`, e.g., `@prompt.txt`).
-   `--email-smarthost`: SMTP Smarthost.
-   `--email-identity`: Email Identity.
-   `--email-to`: Email To.
-   `--email-from`: Email From.
-   `--email-subject`: Email Subject.
-   `--email-username`: Email Username.
-   `--email-password`: Email Password.

## AI Provider selection

The default provider is Poe, but you can use any OpenAI-compatible API.  You can also use the 
`listmodels` command to list available models from the configured API endpoint.

Examples:

```
# Google
--base-url https://generativelanguage.googleapis.com/v1beta/openai/ --model gemini-flash-latest 
```

## Usage

### Fetch and Score and Report (Run)

This is the *do everything* command.  It will fetch new articles, score them, and generate a report.

```bash
./bin/ai-rss-scraper run
```

If you give it an interval, it will run in a loop.

```bash
./bin/ai-rss-scraper run --interval 1h
```

**Options:**
-   `--interval <duration>`: Interval to run the scraper loop (e.g. `1h`, `30m`). 0 means run once.
-   `--no-fetch`: Don't fetch new articles.
-   `--no-score`: Don't score articles.
-   `--no-report`: Don't generate report.
-   `--age <days>`: Age of articles in days to include in report (default: 7).
-   `--threshold <score>`: Score threshold for report (default: 50).
-   `--out <filename>`: Output filename for the report.
-   `--send-email`: Send report via email.

Either `--out` or `--send-email` must be specified.

### Fetch Only

Fetch articles and save them to the database without scoring.

```bash
./bin/ai-rss-scraper fetch
```

### Score Articles Only

Score articles that are currently in the database but haven't been scored yet.

```bash
./bin/ai-rss-scraper score
```

**Options:**
-   `--refresh <pattern>`: Rescore articles with titles matching the wildcard pattern (e.g., `*Retro*`).
-   `--showresponse`: Print the raw JSON response from the LLM for debugging.

```bash
./bin/ai-rss-scraper score --refresh "Lazarus*" --showresponse
```

### Report Only

Generate a report of the articles.

```bash
./bin/ai-rss-scraper report
```

**Options:**
-   `--age <days>`: Age of articles in days to include in report (default: 7).
-   `--threshold <score>`: Score threshold for report (default: 50).
-   `--out <filename>`: Output filename for the report (default: `report.html`).
-   `--send-email`: Send report via email.
-   `--always`: Include articles that have already been reported.

Either `--out` or `--send-email` must be specified.

### List Articles in Database

List recent articles and their scores.

```bash
./bin/ai-rss-scraper list
```

### Dump Database

Dump the full verbose contents of the database, including full article text and analysis.

```bash
./bin/ai-rss-scraper dump
```

### List Models

List available models from the configured API endpoint. This is useful when switching providers, 
as the names of models may change.

```bash
./bin/ai-rss-scraper listmodels
```

### Reset Reporting State

Reset the reporting state of articles in the database. This is useful while developing, so that 
emails can be resent. It takes a wildcard pattern to match article titles to reset.

```bash
./bin/ai-rss-scraper reset-reported "*"
```

## Prompting

You can customize the scoring logic by providing a custom prompt template. The template can use 
Go template variables:

-   `{{.Title}}`
-   `{{.Description}}`
-   `{{.Content}}`

Example `prompt.txt`:
```text
Rate this article on a scale of 0-100 based on how relevant it is to embedded systems.
Title: {{.Title}}
Content: {{.Content}}
```

Prompts may be more complex. For example, the built-in prompt that Scott uses with hackaday is:
```text
Scott likes projects relating to vintage computers and speech synthesizers. In particular he like 
the 4004, 8008, 8080, 8085, 8086, z80, and z8000 cpus. He ikes unique display technologies like 
nixie tubes. He likes raspberry pi and microcontroller projects if there is something unique or 
retro about them. He likes restoring old or rare computers. Produce a numeric score between 0 and 
100 based on how much scott will like this project, please exactly three bullet points on what he 
will like.

Title: {{.Title}}
Description: {{.Description}}
Content: {{.Content}}
```

Usage:
```bash
./bin/ai-rss-scraper score --prompt "@prompt.txt"
```

## Deploy with Helm

A Helm chart is included for deploying the application to Kubernetes.

### Prerequisites

-   A Kubernetes cluster
-   Helm 3+ installed

### Installation

1.  Navigate to the helm directory:
    ```bash
    cd helm/ai-rss-scraper
    ```

2.  Install the chart:
    ```bash
    helm install ai-rss-scraper . \
      --set config.apiKey="your-api-key" \
      --set config.email.smarthost="smtp.example.com:587" \
      --set config.email.username="your-username" \
      --set config.email.password="your-password" \
      --set config.email.to="recipient@example.com" \
      --set config.email.from="sender@example.com"
    ```

### Configuration

You can customize the deployment by modifying `values.yaml` or using `--set` flags.

|Params|Description|Default|
|---|---|---|
| `image.repository` | Docker image repository | `smbaker/ai-rss-scraper` |
| `image.tag` | Docker image tag | `0.0.1` |
| `storage.size` | PVC size | `1Gi` |
| `storage.className` | Storage class name | `standard` |
| `storage.hostPath.enabled` | Enable HostPath for local dev | `false` |
| `storage.hostPath.path` | HostPath directory | `/tmp/ai-rss-scraper-data` |
| `config.feedUrl` | RSS Feed URL | `https://hackaday.com/blog/feed/` |
| `config.model` | LLM Model | `gemini-3-flash` |
| `config.baseUrl` | API Base URL | `https://api.poe.com/v1` |
| `config.runInterval` | Loop interval | `1h` |
| `config.apiKey` | API Key | `""` |
| `config.email.*` | Email settings | (empty) |

**Note:** You will need to update all of the email settings and the API key.

**Note:** Using a persistent volume is recommended for persistence of data.

## License

Apache 2.0
