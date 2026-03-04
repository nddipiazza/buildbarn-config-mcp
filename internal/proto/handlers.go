package proto

import (
"context"
"encoding/json"

"github.com/mark3labs/mcp-go/mcp"
)

// SearchProtosHandler searches Buildbarn proto definitions
func SearchProtosHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
query, err := request.RequireString("query")
if err != nil {
return mcp.NewToolResultError(err.Error()), nil
}

maxResults := request.GetInt("max_results", 10)

// TODO: Implement actual proto search
results := []map[string]interface{}{
{
"fqn":           "buildbarn.configuration.global.TracingConfiguration",
"kind":          "message",
"parent":        "buildbarn.configuration.global.Configuration",
"comment":       "Configuration for OpenTelemetry tracing",
"match_context": "Tracing configuration for distributed tracing",
},
{
"fqn":           "buildbarn.configuration.blobstore.ShardingBlobAccessConfiguration",
"kind":          "message",
"parent":        "buildbarn.configuration.blobstore.BlobAccessConfiguration",
"comment":       "Shard blob access across multiple backends",
"match_context": "Configuration for sharding blobs across storage backends",
},
}

response := map[string]interface{}{
"query":       query,
"max_results": maxResults,
"count":       len(results),
"results":     results,
"note":        "🚧 Mock data - proto index not yet implemented",
}

jsonData, _ := json.MarshalIndent(response, "", "  ")
return mcp.NewToolResultText(string(jsonData)), nil
}

// DescribeMessageHandler returns detailed proto message documentation
func DescribeMessageHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
fqn, err := request.RequireString("fqn")
if err != nil {
return mcp.NewToolResultError(err.Error()), nil
}

depth := request.GetInt("depth", 2)

// TODO: Implement actual proto introspection
response := map[string]interface{}{
"name":    fqn,
"comment": "Top-level configuration for bb-storage service",
"fields": []map[string]interface{}{
{
"name":    "contentAddressableStorage",
"type":    "BlobAccessConfiguration",
"label":   "optional",
"comment": "Backend for the Content Addressable Storage",
},
{
"name":    "actionCache",
"type":    "ActionCacheConfiguration",
"label":   "optional",
"comment": "Configuration for the Action Cache",
},
{
"name":    "global",
"type":    "GlobalConfiguration",
"label":   "optional",
"comment": "Global configuration (diagnostics, tracing)",
},
},
"depth": depth,
"note":  "🚧 Mock data - proto descriptor parsing not yet implemented",
}

jsonData, _ := json.MarshalIndent(response, "", "  ")
return mcp.NewToolResultText(string(jsonData)), nil
}

// GetFieldPathHandler resolves a dotted field path
func GetFieldPathHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
rootMessage, err := request.RequireString("root_message")
if err != nil {
return mcp.NewToolResultError(err.Error()), nil
}

fieldPath, err := request.RequireString("field_path")
if err != nil {
return mcp.NewToolResultError(err.Error()), nil
}

// TODO: Implement actual field path resolution
response := map[string]interface{}{
"root_message": rootMessage,
"field_path":   fieldPath,
"steps": []map[string]interface{}{
{
"field_name":     "blobstore",
"parent_message": rootMessage,
"field_type":     "BlobstoreConfiguration",
"comment":        "Blobstore configuration",
},
{
"field_name":     "backend",
"parent_message": "BlobstoreConfiguration",
"field_type":     "BlobAccessConfiguration",
"comment":        "Backend configuration (oneof: local, grpc, sharding)",
},
},
"leaf_type":    "string",
"leaf_comment": "gRPC address for remote blob access",
"note":         "🚧 Mock data - field path resolver not yet implemented",
}

jsonData, _ := json.MarshalIndent(response, "", "  ")
return mcp.NewToolResultText(string(jsonData)), nil
}

// ListConfigMessagesHandler lists top-level config messages
func ListConfigMessagesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
// TODO: Extract from proto descriptors
messages := []map[string]interface{}{
{
"fqn":     "buildbarn.configuration.bb_storage.ApplicationConfiguration",
"service": "bb-storage",
"comment": "Top-level configuration for bb-storage",
"top_level_fields": []string{
"contentAddressableStorage",
"actionCache",
"global",
"grpcServers",
"schedulers",
},
},
{
"fqn":     "buildbarn.configuration.bb_scheduler.ApplicationConfiguration",
"service": "bb-scheduler",
"comment": "Top-level configuration for bb-scheduler",
"top_level_fields": []string{
"contentAddressableStorage",
"global",
"clientGrpcServers",
"workerGrpcServers",
},
},
{
"fqn":     "buildbarn.configuration.bb_worker.ApplicationConfiguration",
"service": "bb-worker",
"comment": "Top-level configuration for bb-worker",
"top_level_fields": []string{
"blobstore",
"scheduler",
"buildDirectories",
"global",
},
},
}

response := map[string]interface{}{
"count":    len(messages),
"messages": messages,
"note":     "🚧 Mock data - proto descriptors not yet loaded",
}

jsonData, _ := json.MarshalIndent(response, "", "  ")
return mcp.NewToolResultText(string(jsonData)), nil
}
