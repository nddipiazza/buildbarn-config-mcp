package main

import (
"context"
"encoding/json"
"fmt"
"net/http"
"os"
"strings"

"github.com/hermetiq/buildbarn-config-mcp/internal/configedit"
bbgithub "github.com/hermetiq/buildbarn-config-mcp/internal/github"
"github.com/hermetiq/buildbarn-config-mcp/internal/configservice"
"github.com/hermetiq/buildbarn-config-mcp/internal/proto"
"github.com/mark3labs/mcp-go/mcp"
"github.com/mark3labs/mcp-go/server"
"github.com/rs/cors"
)

func main() {
transport := os.Getenv("MCP_TRANSPORT")
if transport == "" {
transport = "stdio"
}
port := os.Getenv("MCP_PORT")
if port == "" {
port = "8080"
}

var ghClient *bbgithub.Client
if os.Getenv("GITHUB_TOKEN") != "" {
var err error
ghClient, err = bbgithub.NewClient()
if err != nil {
fmt.Fprintf(os.Stderr, "GitHub client unavailable: %v\n", err)
}
}

s := server.NewMCPServer(
"Buildbarn Config MCP",
"0.2.0",
server.WithToolCapabilities(true),
server.WithResourceCapabilities(true, false),
)

registerProtoTools(s)
registerBrowseTools(s, ghClient)
registerMutationTools(s)
configservice.RegisterBrowserTools(s, ghClient)

switch strings.ToLower(transport) {
case "http", "streamablehttp":
fmt.Fprintf(os.Stderr, "Buildbarn Config MCP Server starting on HTTP port %s\n", port)
httpServer := server.NewStreamableHTTPServer(s)
corsMiddleware := cors.New(cors.Options{
AllowedOrigins: []string{
"http://localhost:5173",
"http://localhost:3000",
"http://localhost:4173",
},
AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
AllowedHeaders:   []string{"Content-Type", "Authorization", "Accept", "Mcp-Session-Id"},
ExposedHeaders:   []string{"Mcp-Session-Id"},
AllowCredentials: false,
})
handler := corsMiddleware.Handler(httpServer)
if err := http.ListenAndServe(":"+port, handler); err != nil {
fmt.Fprintf(os.Stderr, "HTTP server error: %v\n", err)
os.Exit(1)
}
default:
fmt.Fprintln(os.Stderr, "Buildbarn Config MCP Server starting on stdio")
if err := server.ServeStdio(s); err != nil {
fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
os.Exit(1)
}
}
}

func registerProtoTools(s *server.MCPServer) {
s.AddTool(mcp.NewTool(
"search_protos",
mcp.WithDescription("Search Buildbarn proto definitions"),
mcp.WithString("query", mcp.Required(), mcp.Description("Search query")),
mcp.WithNumber("max_results", mcp.Description("Maximum results (default: 10)")),
), proto.SearchProtosHandler)

s.AddTool(mcp.NewTool(
"describe_message",
mcp.WithDescription("Get documentation for a protobuf message type"),
mcp.WithString("fqn", mcp.Required(), mcp.Description("Fully-qualified name")),
mcp.WithNumber("depth", mcp.Description("Expansion depth (default: 2)")),
), proto.DescribeMessageHandler)

s.AddTool(mcp.NewTool(
"get_field_path",
mcp.WithDescription("Resolve a dotted field path to its proto type chain"),
mcp.WithString("root_message", mcp.Required(), mcp.Description("FQN of root config message")),
mcp.WithString("field_path", mcp.Required(), mcp.Description("Dot-separated path")),
), proto.GetFieldPathHandler)

s.AddTool(mcp.NewTool(
"list_config_messages",
mcp.WithDescription("List top-level Buildbarn configuration messages"),
), proto.ListConfigMessagesHandler)
}

func registerBrowseTools(s *server.MCPServer, gh *bbgithub.Client) {
s.AddTool(mcp.NewTool(
"list_collections",
mcp.WithDescription("List all Buildbarn config collections in Hermetiq/bb-config"),
), func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
if gh == nil {
return mcp.NewToolResultText(`["mock-collection-1","mock-collection-2"]`), nil
}
cols, err := gh.ListCollections(ctx)
if err != nil {
return mcp.NewToolResultError(err.Error()), nil
}
b, _ := json.Marshal(cols)
return mcp.NewToolResultText(string(b)), nil
})

s.AddTool(mcp.NewTool(
"list_files",
mcp.WithDescription("List .jsonnet files in a Buildbarn config collection"),
mcp.WithString("collection_key", mcp.Required(), mcp.Description("Collection folder name")),
), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
key := getString(req, "collection_key")
if key == "" {
return mcp.NewToolResultError("collection_key is required"), nil
}
if gh == nil {
return mcp.NewToolResultText(`["storage.jsonnet","worker.jsonnet"]`), nil
}
files, err := gh.ListFiles(ctx, key)
if err != nil {
return mcp.NewToolResultError(err.Error()), nil
}
b, _ := json.Marshal(files)
return mcp.NewToolResultText(string(b)), nil
})

s.AddTool(mcp.NewTool(
"read_file",
mcp.WithDescription("Read a Buildbarn jsonnet config file from Hermetiq/bb-config"),
mcp.WithString("collection_key", mcp.Required(), mcp.Description("Collection folder name")),
mcp.WithString("file_name", mcp.Required(), mcp.Description("File name (e.g., storage.jsonnet)")),
), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
key := getString(req, "collection_key")
file := getString(req, "file_name")
if key == "" || file == "" {
return mcp.NewToolResultError("collection_key and file_name are required"), nil
}
if gh == nil {
return mcp.NewToolResultText(`{"source":"{ grpcServers: [{listenAddresses: [\"0.0.0.0:8980\"]}] }","sha":"mock-sha"}`), nil
}
src, sha, err := gh.ReadFile(ctx, key, file)
if err != nil {
return mcp.NewToolResultError(err.Error()), nil
}
out, _ := json.Marshal(map[string]string{"source": src, "sha": sha})
return mcp.NewToolResultText(string(out)), nil
})

s.AddTool(mcp.NewTool(
"write_file",
mcp.WithDescription("Commit updated jsonnet content to Hermetiq/bb-config"),
mcp.WithString("collection_key", mcp.Required(), mcp.Description("Collection folder name")),
mcp.WithString("file_name", mcp.Required(), mcp.Description("File name")),
mcp.WithString("content", mcp.Required(), mcp.Description("New jsonnet content")),
mcp.WithString("sha", mcp.Required(), mcp.Description("Blob SHA from read_file")),
mcp.WithString("commit_message", mcp.Description("Commit message")),
), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
key := getString(req, "collection_key")
file := getString(req, "file_name")
content := getString(req, "content")
sha := getString(req, "sha")
msg := getString(req, "commit_message")
if msg == "" {
msg = "Update " + file
}
if key == "" || file == "" || content == "" || sha == "" {
return mcp.NewToolResultError("collection_key, file_name, content, and sha are required"), nil
}
if gh == nil {
return mcp.NewToolResultText(`{"status":"mock-write-ok"}`), nil
}
if err := gh.WriteFile(ctx, key, file, content, sha, msg); err != nil {
return mcp.NewToolResultError(err.Error()), nil
}
return mcp.NewToolResultText(`{"status":"ok"}`), nil
})
}

func registerMutationTools(s *server.MCPServer) {
s.AddTool(mcp.NewTool(
"bb_config_update_field",
mcp.WithDescription("Update a scalar value at a specific path in a Buildbarn jsonnet config"),
mcp.WithString("jsonnet_source", mcp.Required(), mcp.Description("Full jsonnet source")),
mcp.WithString("path", mcp.Required(), mcp.Description(`JSON array path, e.g. ["grpcServers","0","listenAddresses","0"]`)),
mcp.WithString("new_value", mcp.Required(), mcp.Description(`JSON-encoded new value`)),
), mutationHandler(func(data interface{}, req mcp.CallToolRequest) (interface{}, error) {
rawVal := getString(req, "new_value")
var newVal interface{}
if err := json.Unmarshal([]byte(rawVal), &newVal); err != nil {
return nil, fmt.Errorf("new_value must be JSON-encoded: %v", err)
}
path, err := configedit.ParsePath(getString(req, "path"))
if err != nil {
return nil, err
}
return configedit.SetAt(data, path, newVal)
}))

s.AddTool(mcp.NewTool(
"bb_config_add_child",
mcp.WithDescription("Append to array or add key to object in a Buildbarn jsonnet config"),
mcp.WithString("jsonnet_source", mcp.Required(), mcp.Description("Full jsonnet source")),
mcp.WithString("path", mcp.Required(), mcp.Description("JSON array path")),
mcp.WithString("field_name", mcp.Description("For objects: key to add")),
mcp.WithString("initial_value", mcp.Description("JSON-encoded initial value (default: null)")),
), mutationHandler(func(data interface{}, req mcp.CallToolRequest) (interface{}, error) {
path, err := configedit.ParsePath(getString(req, "path"))
if err != nil {
return nil, err
}
fieldName := getString(req, "field_name")
var initVal interface{}
if rawInit := getString(req, "initial_value"); rawInit != "" {
if err := json.Unmarshal([]byte(rawInit), &initVal); err != nil {
return nil, fmt.Errorf("initial_value must be JSON-encoded: %v", err)
}
}
target, err := configedit.GetAt(data, path)
if err != nil {
return nil, err
}
var newTarget interface{}
switch v := target.(type) {
case []interface{}:
clone := make([]interface{}, len(v)+1)
copy(clone, v)
clone[len(v)] = initVal
newTarget = clone
case map[string]interface{}:
if fieldName == "" {
return nil, fmt.Errorf("field_name is required when adding to an object")
}
if _, exists := v[fieldName]; exists {
return nil, fmt.Errorf("key %q already exists; use bb_config_update_field", fieldName)
}
clone := make(map[string]interface{}, len(v)+1)
for k, val := range v {
clone[k] = val
}
clone[fieldName] = initVal
newTarget = clone
default:
return nil, fmt.Errorf("cannot add child to %T", target)
}
return configedit.SetAt(data, path, newTarget)
}))

s.AddTool(mcp.NewTool(
"bb_config_remove_field",
mcp.WithDescription("Remove a field or array element from a Buildbarn jsonnet config"),
mcp.WithString("jsonnet_source", mcp.Required(), mcp.Description("Full jsonnet source")),
mcp.WithString("path", mcp.Required(), mcp.Description("JSON array path to remove")),
), mutationHandler(func(data interface{}, req mcp.CallToolRequest) (interface{}, error) {
path, err := configedit.ParsePath(getString(req, "path"))
if err != nil {
return nil, err
}
return configedit.DeleteAt(data, path)
}))

s.AddTool(mcp.NewTool(
"bb_config_change_oneof",
mcp.WithDescription("Switch the active OneOf case in a Buildbarn jsonnet config"),
mcp.WithString("jsonnet_source", mcp.Required(), mcp.Description("Full jsonnet source")),
mcp.WithString("path", mcp.Required(), mcp.Description("JSON array path to the OneOf node")),
mcp.WithString("new_case", mcp.Required(), mcp.Description("New case field name")),
), mutationHandler(func(data interface{}, req mcp.CallToolRequest) (interface{}, error) {
path, err := configedit.ParsePath(getString(req, "path"))
if err != nil {
return nil, err
}
newCase := getString(req, "new_case")
if newCase == "" {
return nil, fmt.Errorf("new_case is required")
}
return configedit.SetAt(data, path, map[string]interface{}{newCase: map[string]interface{}{}})
}))
}

func mutationHandler(
mutate func(data interface{}, req mcp.CallToolRequest) (interface{}, error),
) server.ToolHandlerFunc {
return func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
src := strings.TrimSpace(getString(req, "jsonnet_source"))
if src == "" {
return mcp.NewToolResultError("jsonnet_source is required"), nil
}
data, err := configedit.EvalJsonnet(src)
if err != nil {
return mcp.NewToolResultError(err.Error()), nil
}
updated, err := mutate(data, req)
if err != nil {
return mcp.NewToolResultError(err.Error()), nil
}
out, err := configedit.DataToJsonnet(updated)
if err != nil {
return mcp.NewToolResultError(err.Error()), nil
}
result, _ := json.Marshal(map[string]string{"updated_jsonnet": out})
return mcp.NewToolResultText(string(result)), nil
}
}

func getString(req mcp.CallToolRequest, key string) string {
if v, ok := req.GetArguments()[key].(string); ok {
return v
}
return ""
}
