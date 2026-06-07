//go:build integration

package testcontainers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok)
	return filepath.Clean(filepath.Join(filepath.Dir(file), "../.."))
}

func startFlakyMock(t *testing.T) (baseURL string, cleanup func()) {
	t.Helper()
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context: repoRoot(t),
		},
		ExposedPorts: []string{"8080/tcp"},
		WaitingFor:   wait.ForHTTP("/health").WithPort("8080/tcp").WithStartupTimeout(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	host, err := container.Host(ctx)
	require.NoError(t, err)
	port, err := container.MappedPort(ctx, "8080/tcp")
	require.NoError(t, err)

	return "http://" + host + ":" + port.Port(), func() {
		_ = container.Terminate(ctx)
	}
}

func TestRetryScenarioWithVerification(t *testing.T) {
	baseURL, cleanup := startFlakyMock(t)
	defer cleanup()

	scenario := []map[string]interface{}{
		{
			"path":   "/payments",
			"method": "POST",
			"responses": []map[string]interface{}{
				{"status": 503, "body": `{"error":"unavailable"}`},
				{"status": 503, "body": `{"error":"unavailable"}`},
				{"status": 200, "body": `{"status":"paid"}`},
			},
		},
	}
	body, err := json.Marshal(scenario)
	require.NoError(t, err)

	resp, err := http.Post(baseURL+"/scenario", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)
	_ = resp.Body.Close()

	client := &http.Client{Timeout: 2 * time.Second}
	var lastStatus int
	for attempt := 1; attempt <= 5; attempt++ {
		req, err := http.NewRequest(http.MethodPost, baseURL+"/payments", nil)
		require.NoError(t, err)
		callResp, err := client.Do(req)
		require.NoError(t, err)
		lastStatus = callResp.StatusCode
		_, _ = io.Copy(io.Discard, callResp.Body)
		_ = callResp.Body.Close()
		if callResp.StatusCode == 200 {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	require.Equal(t, 200, lastStatus)

	verifyResp, err := http.Get(baseURL + "/api/verify/requests?path=/payments&method=POST&min=3")
	require.NoError(t, err)
	defer verifyResp.Body.Close()

	var verify map[string]interface{}
	require.NoError(t, json.NewDecoder(verifyResp.Body).Decode(&verify))
	require.Equal(t, true, verify["matched"])
	require.GreaterOrEqual(t, int(verify["count"].(float64)), 3)

	resetResp, err := http.Post(baseURL+"/api/control/reset-scenarios?path=/payments&method=POST", "application/json", nil)
	require.NoError(t, err)
	require.Equal(t, 200, resetResp.StatusCode)
	_ = resetResp.Body.Close()

	retryResp, err := http.Post(baseURL+"/payments", "application/json", nil)
	require.NoError(t, err)
	defer retryResp.Body.Close()
	require.Equal(t, 503, retryResp.StatusCode, "scenario should restart from first response after reset")
}
