package config

import (
	"fmt"
	"strings"
	"sync/atomic"
	"time"
)

// ScenarioSummary is a read-only view of a loaded scenario.
type ScenarioSummary struct {
	ID            string `json:"id"`
	Path          string `json:"path"`
	Method        string `json:"method"`
	ResponseCount int    `json:"responseCount"`
	CurrentIndex  int    `json:"currentIndex"`
}

// ScenarioRootDir is the root directory for file persistence (mappings/).
var ScenarioRootDir = "."

// SetScenarioRootDir sets the persistence root directory.
func SetScenarioRootDir(root string) {
	if root != "" {
		ScenarioRootDir = root
	}
}

// scenarioMapKey builds the in-memory lookup key for a scenario.
func scenarioMapKey(path, method string) string {
	return path + "_" + strings.ToUpper(method)
}

func ensureScenarioID(s *Scenario) {
	if s.ID != "" {
		return
	}
	safe := strings.NewReplacer("/", "-", "{", "", "}", "", " ", "-").Replace(s.Path)
	s.ID = fmt.Sprintf("%s-%s-%d", safe, strings.ToLower(s.Method), time.Now().UnixNano())
}

// ScenarioForPersist returns a scenario copy without runtime fields.
func ScenarioForPersist(s *Scenario) Scenario {
	return Scenario{
		ID:             s.ID,
		Path:           s.Path,
		Method:         strings.ToUpper(s.Method),
		Matches:        s.Matches,
		Responses:      s.Responses,
		CircuitBreaker: s.CircuitBreaker,
	}
}

// ListScenarios returns summaries of all loaded scenarios.
func ListScenarios() []ScenarioSummary {
	var out []ScenarioSummary
	scenarios.Range(func(_, value interface{}) bool {
		for _, s := range value.([]*Scenario) {
			out = append(out, ScenarioSummary{
				ID:            s.ID,
				Path:          s.Path,
				Method:        s.Method,
				ResponseCount: len(s.Responses),
				CurrentIndex:  int(atomic.LoadInt32(&s.Index)),
			})
		}
		return true
	})
	return out
}

// GetScenarioByID finds a scenario by ID.
func GetScenarioByID(id string) (*Scenario, bool) {
	var found *Scenario
	scenarios.Range(func(_, value interface{}) bool {
		for _, s := range value.([]*Scenario) {
			if s.ID == id {
				found = s
				return false
			}
		}
		return true
	})
	return found, found != nil
}

// ReplaceScenarioByID replaces an existing scenario in place, preserving runtime state.
func ReplaceScenarioByID(id string, updated Scenario) error {
	configLock.Lock()
	defer configLock.Unlock()

	updated.Method = strings.ToUpper(updated.Method)
	if updated.ID == "" {
		updated.ID = id
	}

	var replaced bool
	scenarios.Range(func(key, value interface{}) bool {
		list := value.([]*Scenario)
		for i, s := range list {
			if s.ID != id {
				continue
			}
			updated.CBState = s.CBState
			if updated.CBState == nil {
				updated.CBState = &CircuitBreakerState{State: "closed"}
			}
			atomic.StoreInt32(&updated.Index, atomic.LoadInt32(&s.Index))

			newList := make([]*Scenario, len(list))
			copy(newList, list)
			updatedCopy := updated
			newList[i] = &updatedCopy

			newKey := scenarioMapKey(updated.Path, updated.Method)
			if newKey != key.(string) {
				if len(newList) == 1 {
					scenarios.Delete(key)
				} else {
					trimmed := append([]*Scenario{}, newList[:i]...)
					trimmed = append(trimmed, newList[i+1:]...)
					scenarios.Store(key, trimmed)
				}
				if v, ok := scenarios.Load(newKey); ok {
					old := v.([]*Scenario)
					scenarios.Store(newKey, append(old, &updatedCopy))
				} else {
					scenarios.Store(newKey, []*Scenario{&updatedCopy})
				}
			} else {
				scenarios.Store(key, newList)
			}
			replaced = true
			return false
		}
		return true
	})
	if !replaced {
		return fmt.Errorf("scenario not found: %s", id)
	}
	return nil
}

// DeleteScenarioByID removes a scenario by ID.
func DeleteScenarioByID(id string) bool {
	configLock.Lock()
	defer configLock.Unlock()
	return deleteScenarioByIDLocked(id)
}

func deleteScenarioByIDLocked(id string) bool {
	var deleted bool
	scenarios.Range(func(key, value interface{}) bool {
		list := value.([]*Scenario)
		for i, s := range list {
			if s.ID != id {
				continue
			}
			if len(list) == 1 {
				scenarios.Delete(key)
			} else {
				newList := append([]*Scenario{}, list[:i]...)
				newList = append(newList, list[i+1:]...)
				scenarios.Store(key, newList)
			}
			deleted = true
			return false
		}
		return true
	})
	return deleted
}

// DeleteScenariosByPathMethod removes all scenarios for a path and method.
func DeleteScenariosByPathMethod(path, method string) int {
	configLock.Lock()
	defer configLock.Unlock()

	key := scenarioMapKey(path, method)
	v, ok := scenarios.Load(key)
	if !ok {
		return 0
	}
	count := len(v.([]*Scenario))
	scenarios.Delete(key)
	return count
}

// ClearAllScenarios removes every loaded scenario.
func ClearAllScenarios() {
	configLock.Lock()
	defer configLock.Unlock()
	scenarios.Range(func(key, _ interface{}) bool {
		scenarios.Delete(key)
		return true
	})
}
