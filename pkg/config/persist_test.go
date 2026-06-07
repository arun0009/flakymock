package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPersistAndLoadMappings(t *testing.T) {
	dir := t.TempDir()
	ClearAllScenarios()

	s := &Scenario{
		ID:     "test-get-1",
		Path:   "/persisted",
		Method: "GET",
		Responses: []Response{
			{Status: 503, Body: JSONBody(`"nope"`)},
			{Status: 200, Body: JSONBody(`"ok"`)},
		},
	}
	AddScenario(s)

	written, err := PersistMappingsToDir(dir)
	require.NoError(t, err)
	assert.Equal(t, 1, written)

	ClearAllScenarios()
	loaded, err := LoadMappingsFromDir(dir)
	require.NoError(t, err)
	require.Len(t, loaded, 1)
	assert.Equal(t, "/persisted", loaded[0].Path)
	assert.Equal(t, 503, loaded[0].Responses[0].Status)

	require.NoError(t, DeleteMappingFile(dir, "test-get-1"))
	_, err = os.Stat(filepath.Join(MappingsDir(dir), "test-get-1.json"))
	assert.True(t, os.IsNotExist(err))
}
