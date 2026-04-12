package store

import (
	"os"
	"testing"

	"github.com/mxbossard/tui-journal/internal/immutxtdb/bucket"
	"github.com/mxbossard/utilz/filez"
	"github.com/mxbossard/utilz/ztring"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStore_NewTwoPhasesStore(t *testing.T) {
	tmpDir := filez.MkdirTempOrPanic(t.Name())
	defer os.RemoveAll(tmpDir)

	s, err := NewTwoPhasesStore(tmpDir, "salt")
	assert.NoError(t, err)
	assert.NotNil(t, s)
}

func TestStore_NewBucket(t *testing.T) {
	tmpDir := filez.MkdirTempOrPanic(t.Name())
	defer os.RemoveAll(tmpDir)

	s, err := NewTwoPhasesStore(tmpDir, "salt")
	assert.NoError(t, err)
	assert.NotNil(t, s)

	expectedName := "b1"
	expectedLabels := bucket.NewLabels("foo", "bar")
	b := s.NewBucket(expectedName, expectedLabels)
	assert.NotNil(t, b)

	assert.Equal(t, expectedName, b.Header.Name)
}

func TestStore_Save(t *testing.T) {
	tmpDir := filez.MkdirTempOrPanic(t.Name())
	defer os.RemoveAll(tmpDir)

	s, err := NewTwoPhasesStore(tmpDir, "salt")
	assert.NoError(t, err)
	assert.NotNil(t, s)

	expectedName := "b1"
	expectedLabels := bucket.NewLabels("foo", "bar")
	expectedTxt := ztring.LoremIpsumWords(5)

	b := s.NewBucket(expectedName, expectedLabels)
	assert.NotNil(t, b)

	err = b.WriteText(expectedTxt)
	assert.NoError(t, err)

	err = s.Save(b)
	assert.NoError(t, err)
}

func TestStore_Commit(t *testing.T) {
	//TODO
}

func TestStore_Squashing_Commit(t *testing.T) {
	// Perform multiple saves
	// Check multiple layers exists
	// Commit squashing
	// Get & test only one layer exists
	//TODO
}

func TestStore_Names(t *testing.T) {
	// Save some buckets
	// Commit some buckets
	// Check names
	//TODO
}

func TestStore_Get(t *testing.T) {
	//TODO
	tmpDir := filez.MkdirTempOrPanic(t.Name())
	defer os.RemoveAll(tmpDir)

	// Use a first service
	s, err := NewTwoPhasesStore(tmpDir, "salt")
	assert.NoError(t, err)
	assert.NotNil(t, s)

	expectedName := "b1"
	expectedLabels := bucket.NewLabels("foo", "bar")
	expectedTxt := ztring.LoremIpsumWords(5)

	b := s.NewBucket(expectedName, expectedLabels)
	assert.NotNil(t, b)
	err = b.WriteText(expectedTxt)
	assert.NoError(t, err)

	// Get after Save
	err = s.Save(b)
	assert.NoError(t, err)

	// Use a second service
	s2, err := NewTwoPhasesStore(tmpDir, "salt")
	assert.NoError(t, err)
	assert.NotNil(t, s2)

	b1, err := s2.Get(b.Header.Uid)
	assert.NoError(t, err)
	require.NotNil(t, b1)
	require.NotNil(t, b1.Header)
	assert.Equal(t, expectedName, b1.Header.Name)

	// Use a third service
	s3, err := NewTwoPhasesStore(tmpDir, "salt")
	assert.NoError(t, err)
	assert.NotNil(t, s3)

	// Get after Commit
	err = s3.Commit(b, false)
	assert.NoError(t, err)

	// Use a fourth service
	s4, err := NewTwoPhasesStore(tmpDir, "salt")
	assert.NoError(t, err)
	assert.NotNil(t, s4)

	b2, err := s4.Get(b.Header.Uid)
	assert.NoError(t, err)
	assert.NotNil(t, b2)
	assert.Equal(t, expectedName, b2.Header.Name)
}

func TestStore_Filter(t *testing.T) {
	// Add multiple buckets
	// Save them all & commit some
	// Test filtering
	//TODO
}
