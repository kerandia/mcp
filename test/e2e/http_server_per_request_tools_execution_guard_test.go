// Copyright (c) "Neo4j"
// Neo4j Sweden AB [http://neo4j.com]

//go:build e2e

package e2e

import (
	"context"
	"encoding/base64"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/neo4j/mcp/test/e2e/helpers"

	"github.com/stretchr/testify/require"
)

func TestHTTPPerRequestToolsExecutionGuard(t *testing.T) {
	t.Parallel()

	baseURL := startHTTPModeServer(t)

	t.Run("'write-cypher' call tool request should be blocked when X-Neo4j-MCP-Readonly header is 'true'", func(t *testing.T) {
		t.Parallel()

		cfg := dbs.GetDriverConf()

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		httpClient := newHTTPClient(t, baseURL+"/db/neo4j/mcp", map[string]string{
			"Authorization":        "Basic " + base64.StdEncoding.EncodeToString([]byte(cfg.Username+":"+cfg.Password)),
			"X-Neo4j-MCP-URI":      cfg.URI,
			"X-Neo4j-MCP-ReadOnly": "true",
		})
		defer httpClient.Close()

		require.NoError(t, httpClient.Start(ctx), "http client failed to start")

		_, err := httpClient.Initialize(ctx, helpers.BuildInitializeRequest())
		require.NoError(t, err, "expected initialize to succeed")

		callToolResponse, err := httpClient.CallTool(ctx, mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "write-cypher",
			},
		})
		require.NoError(t, err, "expected no transport error calling write-cypher")
		require.True(t, callToolResponse.IsError, "expected write-cypher to be blocked in read-only mode")

		textContent, ok := callToolResponse.Content[0].(mcp.TextContent)
		require.True(t, ok, "expected text content in error response")
		require.Equal(t, "'write-cypher' is not permitted in read-only mode", textContent.Text)
	})

	t.Run("calling an unknown tool when X-Neo4j-MCP-Readonly header is 'true' should return a protocol-level error", func(t *testing.T) {
		t.Parallel()

		cfg := dbs.GetDriverConf()

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		httpClient := newHTTPClient(t, baseURL+"/db/neo4j/mcp", map[string]string{
			"Authorization":        "Basic " + base64.StdEncoding.EncodeToString([]byte(cfg.Username+":"+cfg.Password)),
			"X-Neo4j-MCP-URI":      cfg.URI,
			"X-Neo4j-MCP-ReadOnly": "true",
		})
		defer httpClient.Close()

		require.NoError(t, httpClient.Start(ctx), "http client failed to start")

		_, err := httpClient.Initialize(ctx, helpers.BuildInitializeRequest())
		require.NoError(t, err, "expected initialize to succeed")

		callToolResponse, err := httpClient.CallTool(ctx, mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "non-existent-tool",
			},
		})
		require.Error(t, err, "expected a protocol-level error for an unknown tool")
		require.Nil(t, callToolResponse)
	})

	t.Run("'read-cypher' call tool request should not be blocked when X-Neo4j-MCP-Readonly header is 'true'", func(t *testing.T) {
		t.Parallel()

		cfg := dbs.GetDriverConf()

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		httpClient := newHTTPClient(t, baseURL+"/db/neo4j/mcp", map[string]string{
			"Authorization":        "Basic " + base64.StdEncoding.EncodeToString([]byte(cfg.Username+":"+cfg.Password)),
			"X-Neo4j-MCP-URI":      cfg.URI,
			"X-Neo4j-MCP-ReadOnly": "true",
		})
		defer httpClient.Close()

		require.NoError(t, httpClient.Start(ctx), "http client failed to start")

		_, err := httpClient.Initialize(ctx, helpers.BuildInitializeRequest())
		require.NoError(t, err, "expected initialize to succeed")

		callToolResponse, err := httpClient.CallTool(ctx, mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "read-cypher",
				Arguments: map[string]any{
					"query": "RETURN 1 AS n",
				},
			},
		})
		require.NoError(t, err, "expected no transport error calling read-cypher")
		require.False(t, callToolResponse.IsError, "expected read-cypher to succeed in read-only mode, response: %+v", callToolResponse)
	})
}
