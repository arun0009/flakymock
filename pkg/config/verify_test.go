package config

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestVerifyRequestsIntervalAndBody(t *testing.T) {
	HistoryMutex.Lock()
	RequestHistory = []RequestRecord{
		{
			ID: "1", Timestamp: time.Now(), Method: "POST", Path: "/pay",
			StatusCode: 503, BodySnippet: `{"retry":true}`,
			Headers:    http.Header{"X-Retry": []string{"1"}},
		},
		{
			ID: "2", Timestamp: time.Now().Add(150 * time.Millisecond), Method: "POST", Path: "/pay",
			StatusCode: 200, BodySnippet: `{"retry":true}`,
			Headers:    http.Header{"X-Retry": []string{"2"}},
		},
	}
	HistoryMutex.Unlock()

	result := VerifyRequests(VerifyFilter{
		Path:         "/pay",
		Method:       "POST",
		BodyContains: "retry",
		Headers:      map[string]string{"X-Retry": "1"},
	}, 1, 0, 100*time.Millisecond, false)
	assert.True(t, result.Matched)
	assert.Equal(t, 1, result.Count)

	resultAll := VerifyRequests(VerifyFilter{
		Path:   "/pay",
		Method: "POST",
	}, 2, 0, 100*time.Millisecond, false)
	assert.True(t, resultAll.Matched)
	assert.True(t, resultAll.IntervalMatched)
}

func TestVerifySequence(t *testing.T) {
	HistoryMutex.Lock()
	RequestHistory = []RequestRecord{
		{Method: "GET", Path: "/health", StatusCode: 200},
		{Method: "POST", Path: "/pay", StatusCode: 503},
		{Method: "POST", Path: "/pay", StatusCode: 200},
	}
	HistoryMutex.Unlock()

	ok, progress := VerifySequence([]VerifyStep{
		{Path: "/pay", Method: "POST", Status: 503},
		{Path: "/pay", Method: "POST", Status: 200},
	})
	assert.True(t, ok)
	assert.Equal(t, 2, progress)
}
