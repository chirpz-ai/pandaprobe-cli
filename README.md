# pandaprobe-cli

An **agent-first** command-line interface for the [PandaProbe](https://pandaprobe.com)
LLM observability platform. It reads traces, sessions and spans, lists and submits
evaluation scores, and orchestrates evaluation runs against the PandaProbe backend.

The compiled binary is named **`pandaprobe`**.

> The PandaProbe SDK owns the *write path* (collecting and ingesting traces).
> This CLI owns the *read + evaluation path* — the work usually done in a dashboard.

## Why "agent-first"?

This CLI is designed to be driven by AI coding agents (Claude Code, Cursor, Cline, …)
as much as by humans:

- **JSON by default.** Every command prints JSON to stdout. Pass `--format table`
  for human-readable output.
- **Data on stdout, errors on stderr.** Always. So `… | jq` just works.
- **Machine-parseable errors.** Errors are JSON objects with a stable shape and a
  meaningful process exit code.
- **No interactivity.** No prompts, spinners, pagers, or colors unless you ask for
  a table.
- **Explicit filters.** Rich server-side filtering keeps responses (and token
  usage) small.

## Install

### macOS / Linux

```bash
curl -fsSL https://cli.pandaprobe.com/install.sh | sh
```

### Windows (PowerShell)

```powershell
irm https://cli.pandaprobe.com/install.ps1 | iex
```

The installers download the prebuilt binary for your OS/arch from the latest
GitHub release, verify its checksum, and place `pandaprobe` on your PATH. Pin a
version or change the location with environment variables:

```bash
PANDAPROBE_VERSION=v0.2.0 PANDAPROBE_INSTALL_DIR="$HOME/.local/bin" \
  curl -fsSL https://cli.pandaprobe.com/install.sh | sh
```

> Until the `cli.pandaprobe.com` vanity host is live, the same scripts work
> directly from the repo, e.g.
> `curl -fsSL https://raw.githubusercontent.com/chirpz-ai/pandaprobe-cli/main/scripts/install.sh | sh`.

### From source

```bash
go install github.com/chirpz-ai/pandaprobe-cli@latest   # installs the `pandaprobe` binary

# or build locally
make build && ./pandaprobe version
```

Prebuilt archives for Linux, macOS and Windows (amd64/arm64) are also attached to
each [GitHub release](https://github.com/chirpz-ai/pandaprobe-cli/releases).

## Quick start

```bash
# 1. Configure credentials (or use env vars / flags — see Configuration)
pandaprobe config set api_key      sk_pp_xxxxxxxx
pandaprobe config set project_name my-agent-project

# 2. List the most recent failed traces as JSON
pandaprobe traces list --status ERROR --limit 20

# 3. Pull one trace and pipe its spans into jq
pandaprobe traces get <trace_id> | jq '.spans[] | {name, kind, status}'

# 4. Submit a programmatic score
pandaprobe evals scores submit --trace-id <trace_id> --name accuracy --value 0.92
```

## Configuration

Resolution precedence (highest to lowest):

1. Command-line flags
2. `PANDAPROBE_*` environment variables
3. Config file `~/.pandaprobe/config.yaml`
4. Built-in defaults

| Setting       | Flag          | Env var                   | Config key     | Default                      |
|---------------|---------------|---------------------------|----------------|------------------------------|
| API key       | `--api-key`   | `PANDAPROBE_API_KEY`      | `api_key`      | —                            |
| Project name  | `--project`   | `PANDAPROBE_PROJECT_NAME` | `project_name` | —                            |
| Endpoint      | `--endpoint`  | `PANDAPROBE_ENDPOINT`     | `endpoint`     | `https://api.pandaprobe.com` |
| Output format | `--format`    | `PANDAPROBE_FORMAT`       | `format`       | `json`                       |
| Timeout (s)   | —             | `PANDAPROBE_TIMEOUT`      | `timeout`      | `30`                         |

Authentication uses the `X-API-Key` and `X-Project-Name` headers.

```bash
pandaprobe config show              # effective config + where each value came from (key masked)
pandaprobe config show --reveal-secrets
pandaprobe config path              # config file location and whether it exists
```

### Global flags

```
--api-key string     PandaProbe API key
--project string     PandaProbe project name
--endpoint string    API endpoint URL
--format string      Output format: json (default) or table
--config string      Path to config file (default ~/.pandaprobe/config.yaml)
--verbose            Log request/response summaries to stderr
--debug              Log full HTTP details to stderr (API key masked)
--no-color           Disable color in table output
```

## Exit codes

Agents can branch on the process exit code:

| Code | Meaning                                            |
|------|----------------------------------------------------|
| 0    | Success                                            |
| 1    | General error (network, decode, unexpected)        |
| 2    | Authentication/authorization error (401, 403)      |
| 3    | Not found (404)                                    |
| 4    | Validation error (bad flags, 400, 422)             |
| 5    | Other API error (other 4xx, 5xx)                   |

Error shape (written to stderr):

```json
{
  "error": {
    "code": "validation_error",
    "message": "invalid --status \"BOGUS\": must be one of PENDING, RUNNING, COMPLETED, ERROR",
    "status": 422,
    "request_id": "…",
    "details": [{ "loc": ["query", "status"], "msg": "…", "type": "enum_error" }]
  }
}
```

## Command reference

### Traces

```
pandaprobe traces list   [--status --session-id --user-id --name --tags
                          --started-after --started-before --sort-by --sort-order
                          --limit --offset]
pandaprobe traces get    <trace_id> [--spans-only] [--kind --status]
pandaprobe traces spans  <trace_id> [--kind --status]
```

`traces get` returns the trace with its spans inline. `traces spans` and
`--spans-only` extract and (optionally) filter spans client-side.

### Sessions

```
pandaprobe sessions list [--user-id --has-error --started-after --started-before
                          --tags --query --sort-by --sort-order --limit --offset]
pandaprobe sessions get  <session_id> [--include-traces=false] [--limit --offset]
```

### Evaluations

All `evals` subcommands accept `--target trace|session` (default `trace`).

```
pandaprobe evals metrics

pandaprobe evals runs create   --metrics m1,m2 [--name --model --sampling-rate]
                               # trace filters:   --date-from --date-to --status
                               #                  --session-id --user-id --tags --filter-name
                               # session filters: --user-id --has-error --tags
                               #                  --min-trace-count --signal-weights
pandaprobe evals runs batch    --metrics m1,m2  (--trace-ids … | --session-ids …) [--name --model]
pandaprobe evals runs list     [--status --limit --offset]
pandaprobe evals runs get      <run_id>
pandaprobe evals runs scores   <run_id>

pandaprobe evals scores list   [--name --source --status --eval-run-id --date-from --date-to
                                --trace-id --data-type --environment   # (trace target)
                                --session-id                           # (session target)
                                --limit --offset]
pandaprobe evals scores get    <trace_id|session_id>
pandaprobe evals scores submit --trace-id … --name … --value …
                               [--data-type --source --reason --metadata]   # trace target only
```

Score submission is **trace-only**: `--target session` on `scores submit` returns a
validation error, because the backend exposes no session-score write endpoint.

### Pagination

List endpoints use offset/limit. Pass `--limit` (1–200) and `--offset`; output
includes a `pagination` block:

```json
{ "items": [ … ], "pagination": { "total": 150, "limit": 20, "offset": 0 } }
```

## Development

```bash
make test        # go test -race ./...
make test-cover  # per-package coverage
make cover       # enforce >=80% coverage on internal/
make lint        # golangci-lint
make build       # build ./pandaprobe with version ldflags
make snapshot    # local GoReleaser cross-compile
```

## Scope (v1)

Read + evaluation only. The CLI never ingests traces (`POST /traces` is the SDK's
job) and performs no destructive operations (no `DELETE`/`PATCH`). Organization
project and API-key management require a different (user-token) auth model and are
intentionally out of scope for the API-key–driven CLI.

## License

MIT — see [LICENSE](./LICENSE).
