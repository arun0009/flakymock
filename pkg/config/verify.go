package config

import (
	"strings"
	"time"
)

// VerifyFilter selects request history records for verification.
type VerifyFilter struct {
	Path         string
	Method       string
	Status       int
	Since        time.Time
	BodyContains string
	Headers      map[string]string
}

// VerifyResult is the outcome of a verification check.
type VerifyResult struct {
	Count               int                      `json:"count"`
	Matched             bool                     `json:"matched"`
	Min                 int                      `json:"min,omitempty"`
	Max                 int                      `json:"max,omitempty"`
	MinInterval         string                   `json:"minInterval,omitempty"`
	MinIntervalObserved string                   `json:"minIntervalObserved,omitempty"`
	IntervalMatched     bool                     `json:"intervalMatched,omitempty"`
	Requests            []map[string]interface{} `json:"requests,omitempty"`
}

// VerifyStep is one step in an ordered sequence check.
type VerifyStep struct {
	Path   string `json:"path"`
	Method string `json:"method"`
	Status int    `json:"status,omitempty"`
}

// FindMatchingRequests returns history records matching the filter.
func FindMatchingRequests(f VerifyFilter) []RequestRecord {
	HistoryMutex.Lock()
	defer HistoryMutex.Unlock()

	var out []RequestRecord
	for _, rec := range RequestHistory {
		if f.Path != "" && rec.Path != f.Path {
			continue
		}
		if f.Method != "" && rec.Method != f.Method {
			continue
		}
		if f.Status != 0 && rec.StatusCode != f.Status {
			continue
		}
		if !f.Since.IsZero() && rec.Timestamp.Before(f.Since) {
			continue
		}
		if f.BodyContains != "" && !strings.Contains(rec.BodySnippet, f.BodyContains) {
			continue
		}
		if len(f.Headers) > 0 {
			ok := true
			for k, v := range f.Headers {
				if rec.Headers.Get(k) != v {
					ok = false
					break
				}
			}
			if !ok {
				continue
			}
		}
		out = append(out, rec)
	}
	return out
}

// MinGapBetween returns the smallest interval between consecutive records.
func MinGapBetween(records []RequestRecord) time.Duration {
	if len(records) < 2 {
		return 0
	}
	minGap := records[len(records)-1].Timestamp.Sub(records[0].Timestamp)
	for i := 1; i < len(records); i++ {
		gap := records[i].Timestamp.Sub(records[i-1].Timestamp)
		if gap < minGap {
			minGap = gap
		}
	}
	return minGap
}

// VerifyRequests runs count and optional interval checks.
func VerifyRequests(f VerifyFilter, minCount, maxCount int, minInterval time.Duration, includeDetail bool) VerifyResult {
	records := FindMatchingRequests(f)
	count := len(records)

	matched := (minCount <= 0 || count >= minCount) && (maxCount <= 0 || count <= maxCount)

	result := VerifyResult{
		Count:   count,
		Matched: matched,
		Min:     minCount,
		Max:     maxCount,
	}

	if minInterval > 0 {
		result.MinInterval = minInterval.String()
		observed := MinGapBetween(records)
		result.MinIntervalObserved = observed.String()
		result.IntervalMatched = count < 2 || observed >= minInterval
		if !result.IntervalMatched {
			result.Matched = false
		}
	}

	if includeDetail {
		result.Requests = make([]map[string]interface{}, 0, len(records))
		for _, rec := range records {
			entry := map[string]interface{}{
				"id":     rec.ID,
				"time":   rec.Timestamp.Format(time.RFC3339),
				"method": rec.Method,
				"path":   rec.Path,
				"status": rec.StatusCode,
			}
			if rec.BodySnippet != "" {
				entry["body"] = rec.BodySnippet
			}
			result.Requests = append(result.Requests, entry)
		}
	}

	return result
}

// VerifySequence checks that steps appear in order within request history.
func VerifySequence(steps []VerifyStep) (bool, int) {
	if len(steps) == 0 {
		return true, 0
	}

	HistoryMutex.Lock()
	defer HistoryMutex.Unlock()

	stepIdx := 0
	matchedCount := 0
	for _, rec := range RequestHistory {
		step := steps[stepIdx]
		method := strings.ToUpper(step.Method)
		if method == "" {
			method = rec.Method
		}
		if rec.Path != step.Path {
			continue
		}
		if method != "" && rec.Method != method {
			continue
		}
		if step.Status != 0 && rec.StatusCode != step.Status {
			continue
		}
		stepIdx++
		matchedCount++
		if stepIdx >= len(steps) {
			return true, matchedCount
		}
	}
	return false, matchedCount
}

// CountMatchingRequests returns how many history records match the given filters.
func CountMatchingRequests(path, method string, status int, since time.Time) int {
	return len(FindMatchingRequests(VerifyFilter{
		Path:   path,
		Method: method,
		Status: status,
		Since:  since,
	}))
}
