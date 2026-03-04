# Buildbarn Config MCP Server

An [MCP](https://modelcontextprotocol.io/) server that gives AI agents (GitHub Copilot, Claude, etc.) tools to read, edit, and commit Buildbarn Jsonnet configurations stored in `Hermetiq/bb-config`.

## Tools

| Tool | Description |
|------|-------------|
| `list_collections` | List cluster config collections in `Hermetiq/bb-config` |
| `list_files` | List `.jsonnet` files in a collection |
| `read_file` | Read a file (returns source + blob SHA) |
| `write_file` | Commit updated content (optimistic locking via SHA) |
| `bb_config_update_field` | Update a scalar value at a JSON path |
| `bb_config_add_child` | Append to array or add key to object |
| `bb_config_remove_field` | Remove a field or array element |
| `bb_config_change_oneof` | Switch a OneOf field to a different case |
| `search_protos` | Search proto definitions |
| `describe_message` | Get docs for a proto message type |
| `get_field_path` | Resolve a dotted field path |
| `list_config_messages` | List top-level Buildbarn config messages |

## Quick Start

### Requirements

- Go 1.22+
- `GITHUB_TOKEN` with read/write access to `Hermetiq/bb-config` (optional — server degrades to mock mode without it)

### Run as HTTP server (for web UI / buildbarn-forms)

```bash
GITHUB_TOKEN=ghp_xxx MCP_TRANSPORT=http MCP_PORT=8080 go run ./cmd/mcp-server
```

Then start the `buildbarn-forms` e2e harness pointing at it:

```bash
cd /path/to/buildbarn-forms/e2e
VITE_MCP_BASE_URL=http://localhost:8080 npm run dev
```

### Run via stdio (for VS Code Copilot agent)

Set `GITHUB_TOKEN` in your shell, then VS Code auto-starts the server via `.vscode/mcp.json`.

Or run directly:

```bash
GITHUB_TOKEN=ghp_xxx go run ./cmd/mcp-server
```

## VS Code Integration

The `.vscode/mcp.json` in this repo configures VS Code to auto-start the server in stdio mode when you open the project. Set `GITHUB_TOKEN` in your shell before opening VS Code (or in your shell profile).

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `MCP_TRANSPORT` | `stdio` | Transport: `stdio` or `http` |
| `MCP_PORT` | `8080` | HTTP port (only used when `MCP_TRANSPORT=http`) |
| `GITHUB_TOKEN` | — | GitHub PAT for `Hermetiq/bb-config` access |

## Development

```bash
# Build
go build -o bin/mcp-server ./cmd/mcp-server

# Test
go test ./...

# Run tests with verbose output
go test -v ./internal/configedit/...
```

## Architecture

```
buildbarn-forms (React UI)
  └── JsonnetEditor (mcpBaseURL prop)
        └── mcpClient.ts → POST /mcp
buildbarn-config-mcp (this repo, Go)
  ├── cmd/mcp-server/main.go  — HTTP/stdio transport, tool registration
  ├── internal/configedit/    — stateless jsonnet mutation helpers
  ├── internal/github/        — GitHub API client for Hermetiq/bb-config
  └── internal/proto/         — proto definition tools (mock index)
Hermetiq/bb-config
  └── <collection>/<file>.jsonnet
```

## Security

- Jsonnet evaluation uses `MemoryImporter` (no file/stdlib imports allowed)
- `MaxStack = 100` prevents recursion DoS
- GitHub writes require a blob SHA for optimistic locking

## Related

- `Hermetiq/buildbarn-forms` — React UI library with JsonnetEditor
- `Hermetiq/bb-config` — Buildbarn jsonnet config storage
- `Hermetiq/cloud-native/bep-nats/mcp` — Production MCP server (stateless tools only)
