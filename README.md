# Buildbarn Config MCP Server

🚀 **Status:** Minimal Working Version (v0.1.0)

## What This Is

An MCP (Model Context Protocol) server that provides AI agents with tools to intelligently navigate and edit Buildbarn Jsonnet configurations.

### Current Status: Minimal Working Demo

✅ **Working:**
- 8 MCP tools with mock implementations
- 4 MCP resources (static and dynamic)
- Runs via stdio transport
- Testable with MCP Inspector

🚧 **Not Yet Implemented:**
- Proto descriptor loading and indexing
- ConfigService gRPC client
- Jsonnet evaluator (go-jsonnet)
- HTTP/SSE transport

## Quick Start

### Build and Run

```bash
# Build
go build -o bin/mcp-server ./cmd/mcp-server

# Run (stdio mode)
./bin/mcp-server

# Or run directly
go run ./cmd/mcp-server
```

### Test with MCP Inspector

```bash
# Install Node.js if you don't have it, then:
npx @modelcontextprotocol/inspector go run ./cmd/mcp-server
```

This opens a web UI where you can:
- See all 8 tools
- Call tools with test inputs
- View mock responses
- See JSON-RPC messages

## Available Tools (8)

### Proto Intelligence (4 tools)
1. **search_protos** - Search proto definitions (returns mock data)
2. **describe_message** - Get proto message docs (returns mock data)
3. **get_field_path** - Resolve field paths (returns mock data)
4. **list_config_messages** - List top-level configs (returns mock data)

### ConfigService (3 tools)
5. **get_config_set** - Get ConfigSet files (returns mock data)
6. **upsert_config_set** - Create/update ConfigSet (returns mock data)
7. **find_config_sets** - Search ConfigSets (returns mock data)

### Jsonnet (1 tool)
8. **render_file** - Evaluate Jsonnet (returns mock data)

## Available Resources (4)

1. **buildbarn://patterns/jsonnet** - Jsonnet patterns (excerpt)
2. **jsonnet://spec/language-reference** - Jsonnet language ref (minimal)
3. **buildbarn://protos/{service}/schema** - Proto schema (mock)
4. **configset://{project}/{name}/files** - ConfigSet files (mock)

## Project Structure

```
buildbarn-config-mcp/
├── cmd/
│   └── mcp-server/
│       └── main.go              # Entry point, registers all tools/resources
├── internal/
│   ├── proto/
│   │   └── handlers.go          # Proto intelligence tools (mocks)
│   ├── configservice/
│   │   └── handlers.go          # ConfigService tools (mocks)
│   ├── jsonnet/
│   │   └── handlers.go          # Jsonnet evaluation tool (mock)
│   └── resources/
│       └── resources.go         # Static and dynamic resources
├── bin/
│   └── mcp-server               # Built binary
├── go.mod
├── go.sum
└── README.md
```

## Example: Call a Tool

Using the MCP Inspector:

1. Open inspector: `npx @modelcontextprotocol/inspector go run ./cmd/mcp-server`
2. Select tool: `search_protos`
3. Enter parameters:
   ```json
   {
     "query": "tracing",
     "max_results": 5
   }
   ```
4. See mock response:
   ```json
   {
     "query": "tracing",
     "max_results": 5,
     "count": 2,
     "results": [...],
     "note": "🚧 Mock data - proto index not yet implemented"
   }
   ```

## Next Steps (Implementation Roadmap)

See `/home/ndipiazza/source/hermetiq/hermetiq-genai-agent/projects/buildbarn-config-mcp/` for full docs:

1. **Proto System** - Vendor Buildbarn protos, generate descriptors
2. **Proto Index** - Implement actual search/describe logic
3. **ConfigService Client** - Connect to real gRPC backend
4. **Jsonnet Evaluator** - Use go-jsonnet to evaluate files
5. **HTTP/SSE Transport** - For production remote access

## Dependencies

- **github.com/mark3labs/mcp-go** - MCP protocol implementation
- Go 1.21+

## Development

```bash
# Add dependencies
go mod tidy

# Build
go build -o bin/mcp-server ./cmd/mcp-server

# Test
npx @modelcontextprotocol/inspector go run ./cmd/mcp-server

# Format
go fmt ./...
```

## Resources

- **Full Project Docs:** `/home/ndipiazza/source/hermetiq/hermetiq-genai-agent/projects/buildbarn-config-mcp/`
- **MCP Protocol:** https://modelcontextprotocol.io/
- **mcp-go Library:** https://github.com/mark3labs/mcp-go
- **MCP Inspector:** https://github.com/modelcontextprotocol/inspector

---

**This is a working skeleton!** All the tools respond with mock data that demonstrates the expected structure. Now you can incrementally replace mocks with real implementations.
