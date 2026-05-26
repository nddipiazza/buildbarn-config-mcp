// Package configservice provides MCP tool handlers that implement the
// React ConfigBrowser's expected tool protocol:
//
//   InitBuildbarnConfig       — called when a file is opened in the editor
//   SetBuildbarnConfigValue   — update a scalar field
//   SetBuildbarnConfigOneOf   — switch a OneOf case
//   AppendBuildbarnConfigItem — append to a repeated field
//   RemoveBuildbarnConfigItem — remove from a repeated field
//
// All tools maintain per-file session state (source + SHA) and optionally
// commit changes back to nddipiazza/bb-config via the GitHub client.
package configservice

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/hermetiq/buildbarn-config-mcp/internal/configedit"
	bbgithub "github.com/hermetiq/buildbarn-config-mcp/internal/github"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// fileSession holds the in-memory editor state for one open file.
type fileSession struct {
	Source  string
	SHA     string
	Data    interface{}
	Version string
}

// sessions maps "project_id:config_set_name:file_name" → *fileSession.
var sessions sync.Map

func sessionKey(projectID, configSetName, fileName string) string {
	return projectID + ":" + configSetName + ":" + fileName
}

func newVersion() string {
	return fmt.Sprintf("v%d", time.Now().UnixMilli())
}

// RegisterBrowserTools adds the React ConfigBrowser-compatible tools to the server.
// Passing a nil gh client disables GitHub commits (local-only mode).
func RegisterBrowserTools(s *server.MCPServer, gh *bbgithub.Client) {

	// ── InitBuildbarnConfig ──────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool(
			"InitBuildbarnConfig",
			mcp.WithDescription("Initialize a config file session in the editor"),
			mcp.WithString("project_id", mcp.Required()),
			mcp.WithString("config_set_name", mcp.Required()),
			mcp.WithString("file_name", mcp.Required()),
			mcp.WithString("initial_content", mcp.Required(), mcp.Description("JSON-serialised evaluated config")),
			mcp.WithString("jsonnet_source", mcp.Description("Raw jsonnet source (optional)")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			projectID := getString(req, "project_id")
			configSetName := getString(req, "config_set_name")
			rawFileName := getString(req, "file_name")
			initialContent := getString(req, "initial_content")
			jsonnetSource := getString(req, "jsonnet_source")

			// Normalise the file name (may arrive as "foo.jsonnet" or just "foo")
			fileName := normaliseFileName(rawFileName)

			// Parse the initial evaluated JSON so we can expose it to later tools
			var data interface{}
			if err := json.Unmarshal([]byte(initialContent), &data); err != nil {
				return mcp.NewToolResultError("invalid initial_content: " + err.Error()), nil
			}

			// Determine SHA from GitHub so we can write back later
			sha := ""
			if gh != nil {
				_, ghSHA, err := gh.ReadFile(ctx, projectID, fileName)
				if err == nil {
					sha = ghSHA
				}
				// If the file doesn't exist in GitHub yet, we'll create it
			}

			// Prefer the incoming jsonnet source; fall back to pretty-printed JSON
			source := jsonnetSource
			if source == "" {
				b, _ := json.MarshalIndent(data, "", "  ")
				source = string(b)
			}

			version := newVersion()
			sessions.Store(sessionKey(projectID, configSetName, fileName), &fileSession{
				Source:  source,
				SHA:     sha,
				Data:    data,
				Version: version,
			})

			result, _ := json.Marshal(map[string]string{
				"version":     version,
				"fileContent": initialContent,
			})
			return mcp.NewToolResultText(string(result)), nil
		},
	)

	// ── SetBuildbarnConfigValue ──────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool(
			"SetBuildbarnConfigValue",
			mcp.WithDescription("Update a scalar field value in the config"),
			mcp.WithString("project_id", mcp.Required()),
			mcp.WithString("config_set_name", mcp.Required()),
			mcp.WithString("file_name", mcp.Required()),
			mcp.WithString("field_path", mcp.Required(), mcp.Description("Dot-separated field path")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			projectID := getString(req, "project_id")
			configSetName := getString(req, "config_set_name")
			rawFileName := getString(req, "file_name")
			fieldPath := getString(req, "field_path")
			value := req.GetArguments()["value"]

			fileName := normaliseFileName(rawFileName)

			sess, err := getSession(projectID, configSetName, fileName)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			// Re-evaluate the stored source to get fresh data
			data, evalErr := configedit.EvalJsonnet(sess.Source)
			if evalErr != nil {
				// Fall back to stored data
				data = sess.Data
			}

			// Apply the field change
			path := dotPathToSegs(fieldPath)
			updated, err := configedit.SetAt(data, path, value)
			if err != nil {
				return mcp.NewToolResultError("SetAt: " + err.Error()), nil
			}

			// Re-serialise as JSON (valid jsonnet)
			updatedSource, err := configedit.DataToJsonnet(updated)
			if err != nil {
				return mcp.NewToolResultError("serialise: " + err.Error()), nil
			}

			version := newVersion()
			fileContent := updatedSource

			// Write to GitHub if client is available
			newSHA := sess.SHA
			if gh != nil {
				commitMsg := fmt.Sprintf("Update %s: set %s", fileName, fieldPath)
				if writeErr := gh.WriteFile(ctx, projectID, fileName, updatedSource, sess.SHA, commitMsg); writeErr != nil {
					fmt.Printf("[configservice] GitHub write failed: %v\n", writeErr)
				} else {
					fmt.Printf("[configservice] ✅ GitHub commit: %s/%s — set %s\n", projectID, fileName, fieldPath)
					// Re-read to get the new SHA
					if _, newGHSHA, readErr := gh.ReadFile(ctx, projectID, fileName); readErr == nil {
						newSHA = newGHSHA
					}
				}
			}

			// Update session
			sessions.Store(sessionKey(projectID, configSetName, fileName), &fileSession{
				Source:  updatedSource,
				SHA:     newSHA,
				Data:    updated,
				Version: version,
			})

			result, _ := json.Marshal(map[string]string{
				"version":       version,
				"fileContent":   fileContent,
				"updatedSource": updatedSource,
			})
			return mcp.NewToolResultText(string(result)), nil
		},
	)

	// ── SetBuildbarnConfigOneOf ──────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool(
			"SetBuildbarnConfigOneOf",
			mcp.WithDescription("Switch the active OneOf case in the config"),
			mcp.WithString("project_id", mcp.Required()),
			mcp.WithString("config_set_name", mcp.Required()),
			mcp.WithString("file_name", mcp.Required()),
			mcp.WithString("field_path", mcp.Required()),
			mcp.WithString("next_case", mcp.Required()),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			projectID := getString(req, "project_id")
			configSetName := getString(req, "config_set_name")
			rawFileName := getString(req, "file_name")
			fieldPath := getString(req, "field_path")
			nextCase := getString(req, "next_case")

			fileName := normaliseFileName(rawFileName)
			sess, err := getSession(projectID, configSetName, fileName)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			data, _ := configedit.EvalJsonnet(sess.Source)
			if data == nil {
				data = sess.Data
			}

			path := dotPathToSegs(fieldPath)
			updated, err := configedit.SetAt(data, path, map[string]interface{}{nextCase: map[string]interface{}{}})
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			return commitAndRespond(ctx, gh, projectID, configSetName, fileName, sess, updated,
				fmt.Sprintf("Update %s: switch OneOf at %s to %s", fileName, fieldPath, nextCase))
		},
	)

	// ── AppendBuildbarnConfigItem ────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool(
			"AppendBuildbarnConfigItem",
			mcp.WithDescription("Append an item to a repeated field in the config"),
			mcp.WithString("project_id", mcp.Required()),
			mcp.WithString("config_set_name", mcp.Required()),
			mcp.WithString("file_name", mcp.Required()),
			mcp.WithString("field_path", mcp.Required()),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			projectID := getString(req, "project_id")
			configSetName := getString(req, "config_set_name")
			rawFileName := getString(req, "file_name")
			fieldPath := getString(req, "field_path")
			defaultValue := req.GetArguments()["default_value"]

			fileName := normaliseFileName(rawFileName)
			sess, err := getSession(projectID, configSetName, fileName)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			data, _ := configedit.EvalJsonnet(sess.Source)
			if data == nil {
				data = sess.Data
			}

			path := dotPathToSegs(fieldPath)
			target, _ := configedit.GetAt(data, path)
			var newTarget interface{}
			switch v := target.(type) {
			case []interface{}:
				clone := make([]interface{}, len(v)+1)
				copy(clone, v)
				if defaultValue == nil {
					defaultValue = map[string]interface{}{}
				}
				clone[len(v)] = defaultValue
				newTarget = clone
			default:
				newTarget = []interface{}{defaultValue}
			}

			updated, err := configedit.SetAt(data, path, newTarget)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			return commitAndRespond(ctx, gh, projectID, configSetName, fileName, sess, updated,
				fmt.Sprintf("Update %s: append to %s", fileName, fieldPath))
		},
	)

	// ── RemoveBuildbarnConfigItem ────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool(
			"RemoveBuildbarnConfigItem",
			mcp.WithDescription("Remove an item from a repeated field in the config"),
			mcp.WithString("project_id", mcp.Required()),
			mcp.WithString("config_set_name", mcp.Required()),
			mcp.WithString("file_name", mcp.Required()),
			mcp.WithString("field_path", mcp.Required()),
			mcp.WithNumber("index", mcp.Required()),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			projectID := getString(req, "project_id")
			configSetName := getString(req, "config_set_name")
			rawFileName := getString(req, "file_name")
			fieldPath := getString(req, "field_path")
			indexFloat, _ := req.GetArguments()["index"].(float64)
			index := int(indexFloat)

			fileName := normaliseFileName(rawFileName)
			sess, err := getSession(projectID, configSetName, fileName)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			data, _ := configedit.EvalJsonnet(sess.Source)
			if data == nil {
				data = sess.Data
			}

			path := append(dotPathToSegs(fieldPath), strconv.Itoa(index))
			updated, err := configedit.DeleteAt(data, path)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			return commitAndRespond(ctx, gh, projectID, configSetName, fileName, sess, updated,
				fmt.Sprintf("Update %s: remove index %d from %s", fileName, index, fieldPath))
		},
	)
}

// ── helpers ──────────────────────────────────────────────────────────────────

func normaliseFileName(name string) string {
	if strings.HasSuffix(name, ".jsonnet") {
		return name
	}
	return name + ".jsonnet"
}

func getString(req mcp.CallToolRequest, key string) string {
	if v, ok := req.GetArguments()[key].(string); ok {
		return v
	}
	return ""
}

func dotPathToSegs(dotPath string) []string {
	if dotPath == "" {
		return nil
	}
	raw := strings.ReplaceAll(dotPath, "[", ".")
	raw = strings.ReplaceAll(raw, "]", "")
	parts := strings.Split(raw, ".")
	result := parts[:0]
	for _, p := range parts {
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

func getSession(projectID, configSetName, fileName string) (*fileSession, error) {
	key := sessionKey(projectID, configSetName, fileName)
	if v, ok := sessions.Load(key); ok {
		return v.(*fileSession), nil
	}
	return nil, fmt.Errorf("no session for %q — call InitBuildbarnConfig first", key)
}

func commitAndRespond(
	ctx context.Context,
	gh *bbgithub.Client,
	projectID, configSetName, fileName string,
	sess *fileSession,
	updated interface{},
	commitMsg string,
) (*mcp.CallToolResult, error) {
	updatedSource, err := configedit.DataToJsonnet(updated)
	if err != nil {
		return mcp.NewToolResultError("serialise: " + err.Error()), nil
	}

	version := newVersion()
	newSHA := sess.SHA

	if gh != nil {
		if writeErr := gh.WriteFile(ctx, projectID, fileName, updatedSource, sess.SHA, commitMsg); writeErr != nil {
			fmt.Printf("[configservice] GitHub write failed: %v\n", writeErr)
		} else {
			fmt.Printf("[configservice] ✅ GitHub commit: %s\n", commitMsg)
			if _, ghSHA, readErr := gh.ReadFile(ctx, projectID, fileName); readErr == nil {
				newSHA = ghSHA
			}
		}
	}

	sessions.Store(sessionKey(projectID, configSetName, fileName), &fileSession{
		Source:  updatedSource,
		SHA:     newSHA,
		Data:    updated,
		Version: version,
	})

	result, _ := json.Marshal(map[string]string{
		"version":       version,
		"fileContent":   updatedSource,
		"updatedSource": updatedSource,
	})
	return mcp.NewToolResultText(string(result)), nil
}
