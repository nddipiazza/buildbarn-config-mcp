package jsonnet

import (
"context"
"encoding/json"

"github.com/mark3labs/mcp-go/mcp"
)

// RenderFileHandler evaluates a Jsonnet file
func RenderFileHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
projectID, err := request.RequireString("project_id")
if err != nil {
return mcp.NewToolResultError(err.Error()), nil
}

configSetName, err := request.RequireString("config_set_name")
if err != nil {
return mcp.NewToolResultError(err.Error()), nil
}

fileName, err := request.RequireString("file_name")
if err != nil {
return mcp.NewToolResultError(err.Error()), nil
}

extVars := make(map[string]string)
if ev := request.GetArguments()["ext_vars"]; ev != nil {
if evMap, ok := ev.(map[string]interface{}); ok {
for k, v := range evMap {
if str, ok := v.(string); ok {
extVars[k] = str
}
}
}
}

// TODO: Implement actual Jsonnet evaluation with go-jsonnet
response := map[string]interface{}{
"project_id":      projectID,
"config_set_name": configSetName,
"file_name":       fileName,
"ext_vars":        extVars,
"json_output": map[string]interface{}{
"contentAddressableStorage": map[string]interface{}{
"sharding": map[string]interface{}{
"shards": map[string]interface{}{
"0": map[string]interface{}{
"backend": map[string]string{"address": "storage-0:8981"},
"weight":  1,
},
},
},
},
},
"note": "🚧 Mock output - Jsonnet evaluator (go-jsonnet) not yet implemented",
}

jsonData, _ := json.MarshalIndent(response, "", "  ")
return mcp.NewToolResultText(string(jsonData)), nil
}
