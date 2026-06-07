package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/arun0009/flakymock/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestIDMiddleware(t *testing.T) {
	handler := requestIDMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Context().Value(RequestIDKey)
		assert.NotNil(t, id, "Request ID not found in context")
	}))

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
}

func TestCorsMiddleware(t *testing.T) {
	// Ensure CORS is enabled in config
	cfg := config.GetConfig()
	cfg.EnableCORS = true
	// Note: In a real scenario we might need to inject config, but here we rely on global state which is tricky.
	// For this unit test, we assume default config or current state allows CORS.

	handler := corsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	req := httptest.NewRequest("OPTIONS", "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Result().StatusCode, "Expected 200 OK for OPTIONS")
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"), "Expected Access-Control-Allow-Origin: *")
}

func TestDynamicPathParameters_Reproduction(t *testing.T) {
	// Setup server
	cfg := config.GetConfig()
	router := NewRouter(cfg)
	ts := httptest.NewServer(router)
	defer ts.Close()

	// 1. Add a dynamic scenario with a path parameter via API
	// The server should register this.
	scenario := config.Scenario{
		Path:   "/api/items/{id}",
		Method: "GET",
		Responses: []config.Response{
			{
				Status: 200,
				Body:   config.JSONBody(`{"item_id": "{{.Request.PathVars.id}}"}`),
			},
		},
	}

	body, _ := json.Marshal([]config.Scenario{scenario})
	resp, err := http.Post(ts.URL+"/scenario", "application/json", bytes.NewReader(body))
	require.NoError(t, err, "Failed to add scenario")
	require.Equal(t, 200, resp.StatusCode, "Failed to add scenario")

	// 2. Request a matching path
	// This relies on the catch-all handler finding the scenario.
	// Suspected bug: The catch-all handler looks up by exact path key, so "/api/items/123" won't match "/api/items/{id}"
	req, err := http.NewRequest("GET", ts.URL+"/api/items/123", nil)
	require.NoError(t, err)

	client := &http.Client{}
	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// 3. Assert
	assert.Equal(t, 200, resp.StatusCode, "Expected 200 OK for dynamic path match")
}

func TestVerifyRequests(t *testing.T) {
	scenarios := config.GetScenarios()
	scenarios.Range(func(key, value interface{}) bool {
		scenarios.Delete(key)
		return true
	})

	cfg := config.GetConfig()
	router := NewRouter(cfg)
	ts := httptest.NewServer(router)
	defer ts.Close()

	scenario := config.Scenario{
		Path:   "/retry",
		Method: "GET",
		Responses: []config.Response{
			{Status: 503, Body: config.JSONBody(`"fail"`)},
			{Status: 200, Body: config.JSONBody(`"ok"`)},
		},
	}
	config.AddScenario(&scenario)

	_, _ = http.Get(ts.URL + "/retry")
	_, _ = http.Get(ts.URL + "/retry")

	verifyResp, err := http.Get(ts.URL + "/api/verify/requests?path=/retry&method=GET&min=2")
	require.NoError(t, err)
	defer verifyResp.Body.Close()

	var result map[string]interface{}
	require.NoError(t, json.NewDecoder(verifyResp.Body).Decode(&result))
	assert.Equal(t, float64(2), result["count"])
	assert.Equal(t, true, result["matched"])
}

func TestResetScenarios(t *testing.T) {
	scenarios := config.GetScenarios()
	scenarios.Range(func(key, value interface{}) bool {
		scenarios.Delete(key)
		return true
	})

	cfg := config.GetConfig()
	router := NewRouter(cfg)
	ts := httptest.NewServer(router)
	defer ts.Close()

	scenario := config.Scenario{
		Path:   "/reset-me",
		Method: "GET",
		Responses: []config.Response{
			{Status: 503, Body: config.JSONBody(`"fail"`)},
			{Status: 200, Body: config.JSONBody(`"ok"`)},
		},
	}
	config.AddScenario(&scenario)

	first, err := http.Get(ts.URL + "/reset-me")
	require.NoError(t, err)
	_ = first.Body.Close()
	assert.Equal(t, 503, first.StatusCode)

	resetResp, err := http.Post(ts.URL+"/api/control/reset-scenarios?path=/reset-me&method=GET", "application/json", nil)
	require.NoError(t, err)
	defer resetResp.Body.Close()
	assert.Equal(t, 200, resetResp.StatusCode)

	again, err := http.Get(ts.URL + "/reset-me")
	require.NoError(t, err)
	_ = again.Body.Close()
	assert.Equal(t, 503, again.StatusCode, "scenario should restart from first response after reset")
}

func TestScenarioCRUDAndPersist(t *testing.T) {
	config.ClearAllScenarios()
	root := t.TempDir()
	config.SetScenarioRootDir(root)

	cfg := config.GetConfig()
	router := NewRouter(cfg)
	ts := httptest.NewServer(router)
	defer ts.Close()

	scenario := config.Scenario{
		Path:   "/crud",
		Method: "GET",
		Responses: []config.Response{
			{Status: 503, Body: config.JSONBody(`"fail"`)},
		},
	}
	body, _ := json.Marshal(scenario)
	resp, err := http.Post(ts.URL+"/scenario", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	var addResult map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&addResult))
	_ = resp.Body.Close()
	ids, ok := addResult["ids"].([]interface{})
	require.True(t, ok)
	require.Len(t, ids, 1)
	id := ids[0].(string)

	listResp, err := http.Get(ts.URL + "/scenarios")
	require.NoError(t, err)
	var listed []config.ScenarioSummary
	require.NoError(t, json.NewDecoder(listResp.Body).Decode(&listed))
	_ = listResp.Body.Close()
	require.Len(t, listed, 1)
	assert.Equal(t, id, listed[0].ID)

	persistResp, err := http.Post(ts.URL+"/api/control/persist-scenarios", "application/json", nil)
	require.NoError(t, err)
	require.Equal(t, 200, persistResp.StatusCode)
	_ = persistResp.Body.Close()

	config.ClearAllScenarios()
	loaded, err := config.LoadMappingsFromDir(root)
	require.NoError(t, err)
	require.Len(t, loaded, 1)
	config.AddScenario(&loaded[0])

	delResp, err := http.NewRequest(http.MethodDelete, ts.URL+"/scenarios?id="+id, nil)
	require.NoError(t, err)
	delResult, err := http.DefaultClient.Do(delResp)
	require.NoError(t, err)
	require.Equal(t, 200, delResult.StatusCode)
	_ = delResult.Body.Close()

	listResp2, err := http.Get(ts.URL + "/scenarios")
	require.NoError(t, err)
	var listed2 []config.ScenarioSummary
	require.NoError(t, json.NewDecoder(listResp2.Body).Decode(&listed2))
	_ = listResp2.Body.Close()
	assert.Len(t, listed2, 0)
}

func TestVerifyRequestsRich(t *testing.T) {
	config.ClearAllScenarios()
	cfg := config.GetConfig()
	router := NewRouter(cfg)
	ts := httptest.NewServer(router)
	defer ts.Close()

	req, err := http.NewRequest(http.MethodPost, ts.URL+"/echo", bytes.NewReader([]byte(`{"retry":true}`)))
	require.NoError(t, err)
	req.Header.Set("X-Retry", "1")
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	_ = resp.Body.Close()

	verifyURL := ts.URL + "/api/verify/requests?path=/echo&method=POST&body_contains=retry&header=X-Retry:1&min=1"
	verifyResp, err := http.Get(verifyURL)
	require.NoError(t, err)
	defer verifyResp.Body.Close()

	var result config.VerifyResult
	require.NoError(t, json.NewDecoder(verifyResp.Body).Decode(&result))
	assert.True(t, result.Matched)
	assert.GreaterOrEqual(t, result.Count, 1)
}
