package configservice

import (
"context"
"encoding/json"

"github.com/mark3labs/mcp-go/mcp"
)

// GetConfigSetHandler retrieves a ConfigSet
func GetConfigSetHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
projectID, err := request.RequireString("project_id")
if err != nil {
return mcp.NewToolResultError(err.Error()), nil
}

configSetName, err := request.RequireString("config_set_name")
if err != nil {
return mcp.NewToolResultError(err.Error()), nil
}

version := request.GetString("version", "")

// TODO: Call actual ConfigService gRPC
configSet := map[string]interface{}{
"project_id":      projectID,
"config_set_name": configSetName,
"version":         version,
"current_version": "abc123def456",
"tags":            []string{"production", "main-cluster"},
"files": []map[string]interface{}{
{
"name":     "common.libsonnet",
"contents": "{\n  blobstore: {\n    contentAddressableStorage: {\n      sharding: {\n        shards: {\n          \"0\": { backend: { grpc: { address: 'storage-0:8981' } }, weight: 1 },\n          \"1\": { backend: { grpc: { address: 'storage-1:8981' } }, weight: 1 },\n        },\n      },\n    },\n  },\n  browserUrl: 'https://browser.example.com',\n  maximumMessageSizeBytes: 10 * 1024 * 1024,\n}",
},
{
"name":     "storage.jsonnet",
"contents": "local common = import 'common.libsonnet';\n{\n  contentAddressableStorage: common.blobstore.contentAddressableStorage,\n  global: common.global,\n}",
},
},
"ext_vars": map[string]string{
"POD_NAME":  "validation-pod",
"NODE_NAME": "validation-node",
},
"note": "🚧 Mock data - ConfigService gRPC client not yet implemented",
}

jsonData, _ := json.MarshalIndent(configSet, "", "  ")
return mcp.NewToolResultText(string(jsonData)), nil
}

// UpsertConfigSetHandler creates or updates a ConfigSet
func UpsertConfigSetHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
projectID, err := request.RequireString("project_id")
if err != nil {
return mcp.NewToolResultError(err.Error()), nil
}

configSetName, err := request.RequireString("config_set_name")
if err != nil {
return mcp.NewToolResultError(err.Error()), nil
}

commitMessage, err := request.RequireString("commit_message")
if err != nil {
return mcp.NewToolResultError(err.Error()), nil
}

filesRaw := request.GetArguments()["files"]
filesMap := make(map[string]interface{})
if filesRaw != nil {
if fm, ok := filesRaw.(map[string]interface{}); ok {
filesMap = fm
}
}

expectedVersion := request.GetString("expected_version", "")

// TODO: Call actual ConfigService gRPC with validation
response := map[string]interface{}{
"project_id":        projectID,
"config_set_name":   configSetName,
"commit_message":    commitMessage,
"expected_version":  expectedVersion,
"new_version":       "def789abc012",
"files_count":       len(filesMap),
"validation_status": "✅ Would validate Jsonnet → JSON → proto (not implemented yet)",
"note":              "🚧 Mock response - ConfigService not connected",
}

jsonData, _ := json.MarshalIndent(response, "", "  ")
return mcp.NewToolResultText(string(jsonData)), nil
}

// FindConfigSetsHandler searches for ConfigSets
func FindConfigSetsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
projectID, err := request.RequireString("project_id")
if err != nil {
return mcp.NewToolResultError(err.Error()), nil
}

namePrefix := request.GetString("name_prefix", "")

// TODO: Call actual ConfigService gRPC
results := []map[string]interface{}{
{
"config_set_name": "prod-cluster",
"current_version": "abc123",
"tags":            []string{"production"},
"last_modified":   "2026-02-15T10:30:00Z",
},
{
"config_set_name": "staging-cluster",
"current_version": "def456",
"tags":            []string{"staging"},
"last_modified":   "2026-02-14T14:20:00Z",
},
}

response := map[string]interface{}{
"project_id":  projectID,
"name_prefix": namePrefix,
"count":       len(results),
"results":     results,
"note":        "🚧 Mock data - ConfigService search not implemented",
}

jsonData, _ := json.MarshalIndent(response, "", "  ")
return mcp.NewToolResultText(string(jsonData)), nil
}
