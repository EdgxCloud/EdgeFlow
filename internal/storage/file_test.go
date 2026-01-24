package storage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileStorageCreation(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.json")

	storage, err := NewFileStorage(dbPath)
	require.NoError(t, err)
	assert.NotNil(t, storage)

	defer storage.Close()

	assert.FileExists(t, dbPath)
}

func TestFileSaveAndLoadFlow(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.json")

	storage, err := NewFileStorage(dbPath)
	require.NoError(t, err)
	defer storage.Close()

	flow := &Flow{
		ID:   "flow-1",
		Name: "Test Flow",
	}

	err = storage.SaveFlow(flow)
	require.NoError(t, err)

	loaded, err := storage.GetFlow("flow-1")
	require.NoError(t, err)
	assert.Equal(t, flow.ID, loaded.ID)
	assert.Equal(t, flow.Name, loaded.Name)
}

func TestFileListFlows(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.json")

	storage, err := NewFileStorage(dbPath)
	require.NoError(t, err)
	defer storage.Close()

	flow1 := &Flow{ID: "flow-1", Name: "Flow 1"}
	flow2 := &Flow{ID: "flow-2", Name: "Flow 2"}

	storage.SaveFlow(flow1)
	storage.SaveFlow(flow2)

	flows, err := storage.ListFlows()
	require.NoError(t, err)
	assert.Len(t, flows, 2)
}

func TestFileDeleteFlow(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.json")

	storage, err := NewFileStorage(dbPath)
	require.NoError(t, err)
	defer storage.Close()

	flow := &Flow{ID: "flow-1", Name: "Test Flow"}
	storage.SaveFlow(flow)

	err = storage.DeleteFlow("flow-1")
	require.NoError(t, err)

	_, err = storage.GetFlow("flow-1")
	assert.Error(t, err)
}

func TestFileConcurrentAccess(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.json")

	storage, err := NewFileStorage(dbPath)
	require.NoError(t, err)
	defer storage.Close()

	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func(id int) {
			flow := &Flow{
				ID:   fmt.Sprintf("flow-%d", id),
				Name: fmt.Sprintf("Flow %d", id),
			}
			storage.SaveFlow(flow)
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	flows, err := storage.ListFlows()
	require.NoError(t, err)
	assert.Len(t, flows, 10)
}
