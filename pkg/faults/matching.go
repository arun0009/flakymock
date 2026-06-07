package faults

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"regexp"
	"strings"

	"github.com/arun0009/flakymock/pkg/config"
)

// matchesRequest checks if the request matches the scenario's rules
func matchesRequest(s *config.Scenario, r *http.Request) bool {
	// 1. Headers
	for k, v := range s.Matches.Headers {
		if r.Header.Get(k) != v {
			return false
		}
	}

	// 2. Query Params
	query := r.URL.Query()
	for k, v := range s.Matches.Query {
		if query.Get(k) != v {
			return false
		}
	}

	// 3. Body Matching
	if len(s.Matches.Body) > 0 {
		if r.Body == nil {
			return false
		}
		// Read body (and restore it)
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			return false
		}
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		// Convert the expected body to string for comparison
		expectedBody := string(s.Matches.Body)
		expectedBody = strings.TrimSpace(expectedBody)

		// 3a. Regex Match (if starts/ends with /)
		// expectedBody from JSON might be quoted, e.g. "/pattern/"
		// We need to be careful. Matches.Body satisfies the JSONBody Unmarshal, so it's raw bytes.
		// If it was a JSON string in YAML, it's just the string content.

		// Let's strip potential outer quotes if it was a JSON string
		cleanExpected := strings.Trim(expectedBody, `"`)
		if len(cleanExpected) > 2 && cleanExpected[0] == '/' && cleanExpected[len(cleanExpected)-1] == '/' {
			pattern := cleanExpected[1 : len(cleanExpected)-1]
			matched, _ := regexp.MatchString(pattern, string(bodyBytes))
			return matched
		}

		// 3b. Smart JSON Match
		// Try to unmarshal both as JSON. If both succeed, do a deep compare.
		var expectedJSON, actualJSON interface{}
		isExpectedJSON := false
		isActualJSON := false

		if err := json.Unmarshal(s.Matches.Body, &expectedJSON); err == nil {
			isExpectedJSON = true
		} else {
			// It might be a raw string that happens to be JSON, or just a string
			// s.Matches.Body is []byte.
			if err := json.Unmarshal([]byte(cleanExpected), &expectedJSON); err == nil {
				isExpectedJSON = true
			}
		}

		if err := json.Unmarshal(bodyBytes, &actualJSON); err == nil {
			isActualJSON = true
		}

		if isExpectedJSON && isActualJSON {
			// Both are valid JSON. Check if actual is a superset or equal.
			// For a mock, usually we want "If request contains these fields".
			// But for simplicity/strictness, let's start with DeepEqual (Exact Match)
			// TODO(feature): Implement detailed subset matching (e.g. partial JSON)
			// For now, let's use reflect.DeepEqual which is standard for "these documents are the same"
			if !jsonDeepEqual(expectedJSON, actualJSON) {
				return false
			}
			return true
		}

		// 3c. Fallback: String Contains
		// If exact JSON matching isn't possible, fall back to string containment
		// We use cleanExpected (no quotes) for this.
		if !strings.Contains(string(bodyBytes), cleanExpected) {
			return false
		}
	}

	return true
}

// jsonDeepEqual compares two JSON objects/arrays.
// It relies on reflect.DeepEqual but handles numeric type differences common in unmarshalling (float64 vs int)
// For simple unmarshalling, everything is float64 or map[string]interface{}.
func jsonDeepEqual(v1, v2 interface{}) bool {
	return reflect.DeepEqual(v1, v2)
}
