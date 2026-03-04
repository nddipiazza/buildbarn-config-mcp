package resources

import (
"encoding/json"
)

// GetJsonnetPatterns returns the Jsonnet patterns resource
func GetJsonnetPatterns() string {
return `# Buildbarn Jsonnet Configuration Patterns (Excerpt)

## File Structure

A ConfigSet contains:
- common.libsonnet - Shared config (blobstore topology, scalars)
- storage.jsonnet, frontend.jsonnet - Service configs
- worker-*.jsonnet - Worker variants

## What Goes in common.libsonnet

✅ Blobstore topology (CAS/AC shards)
✅ Shared scalars (browserUrl, maximumMessageSizeBytes)
✅ Global diagnostics config
❌ Service-specific ports/addresses
❌ Per-service tracing (service.name differs)
❌ Worker-specific settings

## Worker Factory Pattern

` + "```jsonnet\n" + `// worker.libsonnet
local mkWorker(arch) = { /* base config */ };
{ mkWorker: mkWorker }

// worker-arm64.jsonnet
local worker = import 'worker.libsonnet';
worker.mkWorker('arm64')
` + "```\n" + `
## Numeric Expressions

` + "```jsonnet\n" + `// ✅ CORRECT - readable
sizeBytes: 550 * 1024 * 1024 * 1024,  // 550 GiB

// ❌ WRONG - opaque
sizeBytes: 590558003200,
` + "```\n" + `
🚧 This is an excerpt - full patterns would be much longer`
}

// GetJsonnetLanguageReference returns the Jsonnet language reference
func GetJsonnetLanguageReference() string {
return `# Jsonnet Language Reference (Curated)

## Object Composition

` + "```jsonnet\n" + `// Merge objects
a + b  // Right side wins on conflicts

// Deep merge
a +: b  // Makes field overridable by children
` + "```\n" + `
## Imports

` + "```jsonnet\n" + `local common = import 'common.libsonnet';
local worker = import 'worker.libsonnet';
` + "```\n" + `
## ExtVars

` + "```jsonnet\n" + `{
  workerId: {
    pod: std.extVar('POD_NAME'),
    node: std.extVar('NODE_NAME'),
  },
}
` + "```\n" + `
## Functions

` + "```jsonnet\n" + `local mkWorker(arch) = {
  platform: { arch: arch },
};

mkWorker('arm64')
` + "```\n" + `
🚧 This is a minimal reference - full version would cover all features`
}

// GetProtoSchema returns proto schema for a service
func GetProtoSchema(uri string) (string, error) {
// TODO: Parse URI, extract service name, render proto schema
schema := map[string]interface{}{
"service": "bb-storage",
"fqn":     "buildbarn.configuration.bb_storage.ApplicationConfiguration",
"fields": []map[string]interface{}{
{
"name":    "contentAddressableStorage",
"type":    "BlobAccessConfiguration",
"comment": "Backend for CAS",
},
{
"name":    "actionCache",
"type":    "ActionCacheConfiguration",
"comment": "Action cache configuration",
},
},
"note": "🚧 Mock schema - proto descriptor rendering not implemented",
}

jsonData, _ := json.MarshalIndent(schema, "", "  ")
return string(jsonData), nil
}

// GetConfigSetFiles returns ConfigSet files as a resource
func GetConfigSetFiles(uri string) (string, error) {
// TODO: Parse URI, extract project/name, call GetConfigSet
response := map[string]interface{}{
"uri":  uri,
"note": "🚧 Mock response - ConfigSet resource not implemented",
"files": []map[string]string{
{"name": "common.libsonnet", "contents": "{ /* common config */ }"},
{"name": "storage.jsonnet", "contents": "local common = import 'common.libsonnet'; {}"},
},
}

jsonData, _ := json.MarshalIndent(response, "", "  ")
return string(jsonData), nil
}
