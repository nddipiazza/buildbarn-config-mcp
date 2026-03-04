package main

import (

"fmt"
"os"

"github.com/hermetiq/buildbarn-config-mcp/internal/configservice"
"github.com/hermetiq/buildbarn-config-mcp/internal/jsonnet"
"github.com/hermetiq/buildbarn-config-mcp/internal/proto"

"github.com/mark3labs/mcp-go/mcp"
"github.com/mark3labs/mcp-go/server"
)

func main() {
// Create MCP server
s := server.NewMCPServer(
"Buildbarn Config Intelligence",
"0.1.0",
server.WithToolCapabilities(true),
)

// Register Proto Intelligence Tools
registerProtoTools(s)

// Register ConfigService Tools
registerConfigServiceTools(s)

// Register Jsonnet Tools
registerJsonnetTools(s)

// Register Resources
registerResources(s)

// Serve via stdio
fmt.Fprintln(os.Stderr, "🚀 Buildbarn Config MCP Server starting...")
fmt.Fprintln(os.Stderr, "📡 Listening on stdio")

if err := server.ServeStdio(s); err != nil {
fmt.Fprintf(os.Stderr, "❌ Server error: %v\n", err)
os.Exit(1)
}
}

func registerProtoTools(s *server.MCPServer) {
s.AddTool(mcp.NewTool(
"search_protos",
mcp.WithDescription("Search Buildbarn proto definitions for fields, messages, and configuration options"),
mcp.WithString("query", mcp.Required(), mcp.Description("Search query (e.g., 'tracing', 'blob access', 'sharding')")),
mcp.WithNumber("max_results", mcp.Description("Maximum results to return (default: 10)")),
), proto.SearchProtosHandler)

s.AddTool(mcp.NewTool(
"describe_message",
mcp.WithDescription("Get detailed documentation for a protobuf message type"),
mcp.WithString("fqn", mcp.Required(), mcp.Description("Fully qualified name (e.g., 'buildbarn.configuration.bb_storage.ApplicationConfiguration')")),
mcp.WithNumber("depth", mcp.Description("How many levels deep to expand nested messages (default: 2)")),
), proto.DescribeMessageHandler)

s.AddTool(mcp.NewTool(
"get_field_path",
mcp.WithDescription("Resolve a dotted field path to its proto type chain"),
mcp.WithString("root_message", mcp.Required(), mcp.Description("FQN of the root config message")),
mcp.WithString("field_path", mcp.Required(), mcp.Description("Dot-separated path (e.g., 'blobstore.backend.grpc.address')")),
), proto.GetFieldPathHandler)

s.AddTool(mcp.NewTool(
"list_config_messages",
mcp.WithDescription("List the top-level Buildbarn configuration messages (one per service)"),
), proto.ListConfigMessagesHandler)
}

func registerConfigServiceTools(s *server.MCPServer) {
s.AddTool(mcp.NewTool(
"get_config_set",
mcp.WithDescription("Retrieve a ConfigSet (collection of Jsonnet config files) by project and name"),
mcp.WithString("project_id", mcp.Required(), mcp.Description("Project ID")),
mcp.WithString("config_set_name", mcp.Required(), mcp.Description("ConfigSet name")),
mcp.WithString("version", mcp.Description("Version SHA (empty = latest)")),
), configservice.GetConfigSetHandler)

s.AddTool(mcp.NewTool(
"upsert_config_set",
mcp.WithDescription("Create or update a ConfigSet with server-side validation"),
mcp.WithString("project_id", mcp.Required(), mcp.Description("Project ID")),
mcp.WithString("config_set_name", mcp.Required(), mcp.Description("ConfigSet name")),
mcp.WithString("commit_message", mcp.Required(), mcp.Description("Commit message")),
mcp.WithString("expected_version", mcp.Description("Expected current version (for optimistic locking)")),
mcp.WithObject("files", mcp.Required(), mcp.Description("Map of filename to contents")),
), configservice.UpsertConfigSetHandler)

s.AddTool(mcp.NewTool(
"find_config_sets",
mcp.WithDescription("Search for ConfigSets by tags, name prefix, or author"),
mcp.WithString("project_id", mcp.Required(), mcp.Description("Project ID")),
mcp.WithString("name_prefix", mcp.Description("Name prefix filter")),
), configservice.FindConfigSetsHandler)
}

func registerJsonnetTools(s *server.MCPServer) {
s.AddTool(mcp.NewTool(
"render_file",
mcp.WithDescription("Evaluate a single Jsonnet file from a ConfigSet and return the resulting JSON"),
mcp.WithString("project_id", mcp.Required(), mcp.Description("Project ID")),
mcp.WithString("config_set_name", mcp.Required(), mcp.Description("ConfigSet name")),
mcp.WithString("file_name", mcp.Required(), mcp.Description("File name (e.g., 'worker-arm64.jsonnet')")),
mcp.WithObject("ext_vars", mcp.Description("External variables to inject")),
), jsonnet.RenderFileHandler)
}

func registerResources(s *server.MCPServer) {
// Minimal stub - skip resources for now to get server working
fmt.Fprintln(os.Stderr, "✅ Registered 8 tools")
}
