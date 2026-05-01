package store

import (
	"os"
	"testing"

	"github.com/mxbossard/tui-journal/internal/immutxtdb/bucket"
	"github.com/mxbossard/utilz/filez"
	"github.com/mxbossard/utilz/iterz"
	"github.com/mxbossard/utilz/timez"
	"github.com/mxbossard/utilz/ztring"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStore_NewTwoPhasesStore(t *testing.T) {
	tmpEDir := filez.MkdirTempOrPanic(t.Name())
	defer os.RemoveAll(tmpEDir)
	tmpRDir := filez.MkdirTempOrPanic(t.Name())
	defer os.RemoveAll(tmpRDir)

	s, err := NewTwoPhasesStore(tmpEDir, tmpRDir, "device", "salt")
	assert.NoError(t, err)
	assert.NotNil(t, s)
}

func TestStore_NewBucket(t *testing.T) {
	tmpEDir := filez.MkdirTempOrPanic(t.Name())
	defer os.RemoveAll(tmpEDir)
	tmpRDir := filez.MkdirTempOrPanic(t.Name())
	defer os.RemoveAll(tmpRDir)

	s, err := NewTwoPhasesStore(tmpEDir, tmpRDir, "device", "salt")
	assert.NoError(t, err)
	assert.NotNil(t, s)

	expectedName := "b1"
	expectedLabels := bucket.NewLabels("foo", "bar")
	b := s.NewBucket(expectedName, expectedLabels)
	assert.NotNil(t, b)

	assert.Equal(t, expectedName, b.Header.Name)
}

func TestStore_Save(t *testing.T) {
	tmpEDir := filez.MkdirTempOrPanic(t.Name())
	defer os.RemoveAll(tmpEDir)
	tmpRDir := filez.MkdirTempOrPanic(t.Name())
	defer os.RemoveAll(tmpRDir)

	s, err := NewTwoPhasesStore(tmpEDir, tmpRDir, "device", "salt")
	assert.NoError(t, err)
	assert.NotNil(t, s)

	expectedName := "b1"
	expectedLabels := bucket.NewLabels("foo", "bar")
	expectedTxt := ztring.LoremIpsumWords(5)

	b := s.NewBucket(expectedName, expectedLabels)
	assert.NotNil(t, b)

	err = b.SetText(expectedTxt)
	assert.NoError(t, err)

	err = s.Save(b)
	assert.NoError(t, err)
}

func TestStore_Names(t *testing.T) {
	// Save some buckets
	// Commit some buckets
	// Check names
	//TODO
}

func TestStore_Get(t *testing.T) {
	tmpEDir := filez.MkdirTempOrPanic(t.Name())
	defer os.RemoveAll(tmpEDir)
	tmpRDir := filez.MkdirTempOrPanic(t.Name())
	defer os.RemoveAll(tmpRDir)

	// Use a first service
	s, err := NewTwoPhasesStore(tmpEDir, tmpRDir, "device", "salt")
	assert.NoError(t, err)
	assert.NotNil(t, s)

	expectedName := "b1"
	expectedLabels := bucket.NewLabels("foo", "bar")
	expectedTxt := ztring.LoremIpsumWords(5)

	b := s.NewBucket(expectedName, expectedLabels)
	assert.NotNil(t, b)
	err = b.SetText(expectedTxt)
	assert.NoError(t, err)

	// Get after Save
	err = s.Save(b)
	assert.NoError(t, err)

	// Use a second service
	s2, err := NewTwoPhasesStore(tmpEDir, tmpRDir, "device", "salt")
	assert.NoError(t, err)
	assert.NotNil(t, s2)

	b1, err := s2.Get(b.Header.Uid)
	assert.NoError(t, err)
	require.NotNil(t, b1)
	require.NotNil(t, b1.Header)
	assert.Equal(t, expectedName, b1.Header.Name)

	// Use a third service to Get the Bucket
	s3, err := NewTwoPhasesStore(tmpEDir, tmpRDir, "device", "salt")
	assert.NoError(t, err)
	assert.NotNil(t, s3)

	b2, err := s3.Get(b.Header.Uid)
	assert.NoError(t, err)
	assert.NotNil(t, b2)
	assert.Equal(t, expectedName, b2.Header.Name)
}

func TestStore_Commit(t *testing.T) {
	tmpEDir1 := filez.MkdirTempOrPanic(t.Name())
	defer os.RemoveAll(tmpEDir1)
	tmpEDir2 := filez.MkdirTempOrPanic(t.Name())
	defer os.RemoveAll(tmpEDir2)
	tmpRDir := filez.MkdirTempOrPanic(t.Name())
	defer os.RemoveAll(tmpRDir)

	// Use a first service
	s, err := NewTwoPhasesStore(tmpEDir1, tmpRDir, "device", "salt")
	assert.NoError(t, err)
	assert.NotNil(t, s)

	expectedName := "b1"
	expectedLabels := bucket.NewLabels("foo", "bar")
	expectedTxt := ztring.LoremIpsumWords(5)

	b := s.NewBucket(expectedName, expectedLabels)
	assert.NotNil(t, b)
	err = b.SetText(expectedTxt)
	assert.NoError(t, err)

	// Get after Save
	err = s.Save(b)
	assert.NoError(t, err)

	// Use a second service with different ephemeral dir but same rested dir
	s2, err := NewTwoPhasesStore(tmpEDir2, tmpRDir, "device", "salt")
	assert.NoError(t, err)
	assert.NotNil(t, s2)

	// bkt should not exists in rested dir
	b2, err := s2.Get(b.Header.Uid)
	assert.Nil(t, b2)
	assert.Error(t, err)
	assert.Equal(t, bucket.ErrNotExist, err)

	// Commit bkt in s1
	err = s.Commit(b, false)
	assert.NoError(t, err)

	// bkt should exists in rested dir
	b3, err := s2.Get(b.Header.Uid)
	assert.NoError(t, err)
	require.NotNil(t, b3)
	require.NotNil(t, b3.Header)
	assert.Equal(t, expectedName, b3.Header.Name)
}

func TestStore_Squashing_Commit(t *testing.T) {
	// Perform multiple saves
	// Check multiple layers exists
	// Commit squashing
	// Get & test only one layer exists
	//TODO
}

func TestStore_Filter(t *testing.T) {
	// Add multiple buckets
	// Save them all & commit some
	// Test filtering
	tmpEDir1 := filez.MkdirTempOrPanic(t.Name())
	defer os.RemoveAll(tmpEDir1)
	tmpEDir2 := filez.MkdirTempOrPanic(t.Name())
	defer os.RemoveAll(tmpEDir2)
	tmpRDir := filez.MkdirTempOrPanic(t.Name())
	defer os.RemoveAll(tmpRDir)

	// Use a first service
	s1, err := NewTwoPhasesStore(tmpEDir1, tmpRDir, "device", "salt")
	assert.NoError(t, err)
	assert.NotNil(t, s1)

	expectedLabels := bucket.NewLabels("foo", "bar")
	expectedNameA := "bA"
	expectedTxtA1 := ztring.LoremIpsumWords(2)
	expectedNameB := "bB"
	expectedTxtB1 := ztring.LoremIpsumWords(4)
	expectedNameC := "bC"
	expectedTxtC1 := ztring.LoremIpsumWords(6)
	expectedNameD := "bD"
	expectedTxtD1 := ztring.LoremIpsumWords(8)
	expectedNameE := "bE"
	expectedTxtE1 := ztring.LoremIpsumWords(10)

	t0 := timez.ParseYyyyMmDd("2026-04-30")

	// bA
	t1 := timez.ParseYyyyMmDd("2026-05-01")
	bA := s1.NewBucket(expectedNameA, expectedLabels)
	assert.NotNil(t, bA)
	err = bA.SetText(expectedTxtA1)
	assert.NoError(t, err)
	err = s1.Save(bA, t1)
	assert.NoError(t, err)

	// bB
	t2 := timez.ParseYyyyMmDd("2026-05-02")
	bB := s1.NewBucket(expectedNameB, expectedLabels)
	assert.NotNil(t, bB)
	err = bB.SetText(expectedTxtB1)
	assert.NoError(t, err)
	err = s1.Save(bB, t2)
	assert.NoError(t, err)

	// bC
	t3 := timez.ParseYyyyMmDd("2026-05-03")
	bC := s1.NewBucket(expectedNameC, expectedLabels)
	assert.NotNil(t, bC)
	err = bC.SetText(expectedTxtC1)
	assert.NoError(t, err)
	err = s1.Save(bC, t3)
	assert.NoError(t, err)

	// bD
	t4 := timez.ParseYyyyMmDd("2026-05-04")
	bD := s1.NewBucket(expectedNameD, expectedLabels)
	assert.NotNil(t, bD)
	err = bD.SetText(expectedTxtD1)
	assert.NoError(t, err)
	err = s1.Save(bD, t4)
	assert.NoError(t, err)

	// bE
	t5 := timez.ParseYyyyMmDd("2026-05-05")
	bE := s1.NewBucket(expectedNameE, expectedLabels)
	assert.NotNil(t, bE)
	err = bE.SetText(expectedTxtE1)
	assert.NoError(t, err)
	err = s1.Save(bE, t5)
	assert.NoError(t, err)

	tEnd := timez.ParseYyyyMmDd("2026-05-31")
	_ = t0
	_ = tEnd

	// Use a second service with different ephemeral dir
	s2, err := NewTwoPhasesStore(tmpEDir2, tmpRDir, "device", "salt")
	assert.NoError(t, err)
	assert.NotNil(t, s2)

	// Check bA
	bA1, err := s1.Get(bA.Header.Uid)
	assert.NoError(t, err)
	require.NotNil(t, bA1)
	require.NotNil(t, bA1.Header)
	assert.Equal(t, expectedNameA, bA1.Header.Name)

	bA2, err := s2.Get(bA.Header.Uid)
	assert.Equal(t, bucket.ErrNotExist, err)
	require.Nil(t, bA2)

	// Check bB
	bB1, err := s1.Get(bB.Header.Uid)
	assert.NoError(t, err)
	require.NotNil(t, bB1)
	require.NotNil(t, bB1.Header)
	assert.Equal(t, expectedNameB, bB1.Header.Name)

	// Check second store nothing commited so nothing should be found
	bB2, err := s2.Get(bB.Header.Uid)
	assert.Equal(t, bucket.ErrNotExist, err)
	assert.Nil(t, bB2)

	// Check bC
	bC1, err := s1.Get(bC.Header.Uid)
	assert.NoError(t, err)
	require.NotNil(t, bC1)
	require.NotNil(t, bC1.Header)
	assert.Equal(t, expectedNameC, bC1.Header.Name)

	bC2, err := s2.Get(bC.Header.Uid)
	assert.Equal(t, bucket.ErrNotExist, err)
	assert.Nil(t, bC2)

	// Check bD
	bD1, err := s1.Get(bD.Header.Uid)
	assert.NoError(t, err)
	require.NotNil(t, bD1)
	require.NotNil(t, bD1.Header)
	assert.Equal(t, expectedNameD, bD1.Header.Name)

	// Check second store nothing commited so nothing should be found
	bD2, err := s2.Get(bD.Header.Uid)
	assert.Equal(t, bucket.ErrNotExist, err)
	assert.Nil(t, bD2)

	// Check bE
	bE1, err := s1.Get(bE.Header.Uid)
	assert.NoError(t, err)
	require.NotNil(t, bE1)
	require.NotNil(t, bE1.Header)
	assert.Equal(t, expectedNameE, bE1.Header.Name)

	bE2, err := s2.Get(bE.Header.Uid)
	assert.Equal(t, bucket.ErrNotExist, err)
	assert.Nil(t, bE2)

	// Check Filtering

	// BktA created between t0 & t2
	paginerA1, err := s1.Filter(bucket.OlderFirst, bucket.CreatedBetweenCriterion(t0, t2), 1, 0)
	assert.NoError(t, err)
	require.NotNil(t, paginerA1)
	entriesA1 := iterz.Flatten(paginerA1.All())
	assert.Len(t, entriesA1, 1)
	require.True(t, len(entriesA1) > 0)
	assert.Equal(t, expectedNameA, entriesA1[0].Val().Header.Name)
	txt, err := entriesA1[0].Val().ProjectText(bucket.LatestVersion)
	assert.NoError(t, err)
	assert.Equal(t, expectedTxtA1, txt)

	// No bucket created between t0 & t1 (limits excluded)
	paginerA2, err := s2.Filter(bucket.OlderFirst, bucket.CreatedBetweenCriterion(t0, t1), 1, 0)
	assert.NoError(t, err)
	require.NotNil(t, paginerA1)
	entriesA2 := iterz.Flatten(paginerA2.All())
	assert.Len(t, entriesA2, 0)

	// BktC created between t2 & t4 (BktC)
	paginerC1, err := s1.Filter(bucket.OlderFirst, bucket.CreatedBetweenCriterion(t2, t4), 1, 0)
	assert.NoError(t, err)
	require.NotNil(t, paginerC1)
	entriesC1 := iterz.Flatten(paginerC1.All())
	assert.Len(t, entriesC1, 1)
	require.True(t, len(entriesC1) > 0)
	assert.Equal(t, expectedNameC, entriesC1[0].Val().Header.Name)
	txt, err = entriesC1[0].Val().ProjectText(bucket.LatestVersion)
	assert.NoError(t, err)
	assert.Equal(t, expectedTxtC1, txt)

	// BktB created between t1 & t3 (BktB)
	paginerB1, err := s1.Filter(bucket.OlderFirst, bucket.CreatedBetweenCriterion(t1, t3), 1, 0)
	assert.NoError(t, err)
	require.NotNil(t, paginerB1)
	entriesB1 := iterz.Flatten(paginerB1.All())
	assert.Len(t, entriesB1, 1)
	require.True(t, len(entriesB1) > 0)
	assert.Equal(t, expectedNameB, entriesB1[0].Val().Header.Name)
	txt, err = entriesB1[0].Val().ProjectText(bucket.LatestVersion)
	assert.NoError(t, err)
	assert.Equal(t, expectedTxtB1, txt)

	// Refilter to check consistency
	paginerB2, err := s1.Filter(bucket.OlderFirst, bucket.CreatedBetweenCriterion(t1, t3), 1, 0)
	assert.NoError(t, err)
	require.NotNil(t, paginerB2)
	entriesB2 := iterz.Flatten(paginerB2.All())
	assert.Len(t, entriesB2, 1)
	require.True(t, len(entriesB2) > 0)
	assert.Equal(t, expectedNameB, entriesB2[0].Val().Header.Name)
	txt, err = entriesB2[0].Val().ProjectText(bucket.LatestVersion)
	assert.NoError(t, err)
	assert.Equal(t, expectedTxtB1, txt)

	// Search for updates: all buckets should match
	paginerD1, err := s1.Filter(bucket.OlderFirst, bucket.UpdatedBetweenCriterion(t0, tEnd), 1, 0)
	assert.NoError(t, err)
	require.NotNil(t, paginerD1)
	entriesD1 := iterz.Flatten(paginerD1.All())
	assert.Len(t, entriesD1, 5)

	// Search for updates: no buckets should match (no bucket commited
	paginerD2, err := s2.Filter(bucket.OlderFirst, bucket.UpdatedBetweenCriterion(t0, tEnd), 1, 0)
	assert.NoError(t, err)
	require.NotNil(t, paginerD2)
	entriesD2 := iterz.Flatten(paginerD2.All())
	assert.Len(t, entriesD2, 0)
}

func TestStore_Filter_Commited_Updated(t *testing.T) {
	// Add multiple buckets
	// Save them all & commit some
	// Test filtering
	tmpEDir1 := filez.MkdirTempOrPanic(t.Name())
	defer os.RemoveAll(tmpEDir1)
	tmpEDir2 := filez.MkdirTempOrPanic(t.Name())
	defer os.RemoveAll(tmpEDir2)
	tmpRDir := filez.MkdirTempOrPanic(t.Name())
	defer os.RemoveAll(tmpRDir)

	// Use a first service
	s1, err := NewTwoPhasesStore(tmpEDir1, tmpRDir, "device", "salt")
	assert.NoError(t, err)
	assert.NotNil(t, s1)

	expectedLabels := bucket.NewLabels("foo", "bar")
	expectedNameA := "bA"
	expectedTxtA1 := ztring.LoremIpsumWords(2)
	expectedNameB := "bB"
	expectedTxtB1 := ztring.LoremIpsumWords(4)
	expectedNameC := "bC"
	expectedTxtC1 := ztring.LoremIpsumWords(6)
	expectedNameD := "bD"
	expectedTxtD1 := ztring.LoremIpsumWords(8)
	expectedNameE := "bE"
	expectedTxtE1 := ztring.LoremIpsumWords(10)

	t0 := timez.ParseYyyyMmDd("2026-04-30")

	// bA
	t1 := timez.ParseYyyyMmDd("2026-05-01")
	bA := s1.NewBucket(expectedNameA, expectedLabels)
	assert.NotNil(t, bA)
	err = bA.SetText(expectedTxtA1)
	assert.NoError(t, err)
	err = s1.Save(bA, t1)
	assert.NoError(t, err)

	// bB (commited)
	t2 := timez.ParseYyyyMmDd("2026-05-02")
	bB := s1.NewBucket(expectedNameB, expectedLabels)
	assert.NotNil(t, bB)
	err = bB.SetText(expectedTxtB1)
	assert.NoError(t, err)
	err = s1.Save(bB, t2)
	assert.NoError(t, err)
	err = s1.Commit(bB, false)
	assert.NoError(t, err)
	// Update bB AFTER Commit
	err = bB.SetText(expectedTxtB1 + "updated")
	assert.NoError(t, err)
	err = s1.Save(bB, t2)
	assert.NoError(t, err)

	// bC
	t3 := timez.ParseYyyyMmDd("2026-05-03")
	bC := s1.NewBucket(expectedNameC, expectedLabels)
	assert.NotNil(t, bC)
	err = bC.SetText(expectedTxtC1)
	assert.NoError(t, err)
	err = s1.Save(bC, t3)
	assert.NoError(t, err)

	// bD (commited)
	t4 := timez.ParseYyyyMmDd("2026-05-04")
	bD := s1.NewBucket(expectedNameD, expectedLabels)
	assert.NotNil(t, bD)
	err = bD.SetText(expectedTxtD1)
	assert.NoError(t, err)
	err = s1.Save(bD, t4)
	assert.NoError(t, err)
	// Update bD BEFORE Commit
	err = bD.SetText(expectedTxtD1 + "updated")
	assert.NoError(t, err)
	err = s1.Save(bD, t4)
	assert.NoError(t, err)
	err = s1.Commit(bD, false)
	assert.NoError(t, err)

	// Update bC
	err = bC.SetText(expectedTxtC1 + "updated")
	assert.NoError(t, err)
	err = s1.Save(bC, t4)
	assert.NoError(t, err)

	// bE
	t5 := timez.ParseYyyyMmDd("2026-05-05")
	bE := s1.NewBucket(expectedNameE, expectedLabels)
	assert.NotNil(t, bE)
	err = bE.SetText(expectedTxtE1)
	assert.NoError(t, err)
	err = s1.Save(bE, t5)
	assert.NoError(t, err)

	tEnd := timez.ParseYyyyMmDd("2026-05-31")
	_ = t0
	_ = tEnd

	// Use a second service with different ephemeral dir
	s2, err := NewTwoPhasesStore(tmpEDir2, tmpRDir, "device", "salt")
	assert.NoError(t, err)
	assert.NotNil(t, s2)

	// Check bA
	bA1, err := s1.Get(bA.Header.Uid)
	assert.NoError(t, err)
	require.NotNil(t, bA1)
	require.NotNil(t, bA1.Header)
	assert.Equal(t, expectedNameA, bA1.Header.Name)

	// Check second store only commited bucket should be found
	bA2, err := s2.Get(bA.Header.Uid)
	assert.Equal(t, bucket.ErrNotExist, err)
	require.Nil(t, bA2)

	// Check bB
	bB1, err := s1.Get(bB.Header.Uid)
	assert.NoError(t, err)
	require.NotNil(t, bB1)
	require.NotNil(t, bB1.Header)
	assert.Equal(t, expectedNameB, bB1.Header.Name)

	// Check second store only commited bucket should be found
	bB2, err := s2.Get(bB.Header.Uid)
	assert.NoError(t, err)
	require.NotNil(t, bB2)
	require.NotNil(t, bB2.Header)
	assert.Equal(t, expectedNameB, bB2.Header.Name)

	// Check bC
	bC1, err := s1.Get(bC.Header.Uid)
	assert.NoError(t, err)
	require.NotNil(t, bC1)
	require.NotNil(t, bC1.Header)
	assert.Equal(t, expectedNameC, bC1.Header.Name)

	// Check second store only commited bucket should be found
	bC2, err := s2.Get(bC.Header.Uid)
	assert.Equal(t, bucket.ErrNotExist, err)
	require.Nil(t, bC2)

	// Check bD
	bD1, err := s1.Get(bD.Header.Uid)
	assert.NoError(t, err)
	require.NotNil(t, bD1)
	require.NotNil(t, bD1.Header)
	assert.Equal(t, expectedNameD, bD1.Header.Name)

	// Check second store only commited bucket should be found
	bD2, err := s2.Get(bD.Header.Uid)
	assert.NoError(t, err)
	require.NotNil(t, bD2)
	require.NotNil(t, bD2.Header)
	assert.Equal(t, expectedNameD, bD2.Header.Name)

	// Check bA
	bE1, err := s1.Get(bE.Header.Uid)
	assert.NoError(t, err)
	require.NotNil(t, bE1)
	require.NotNil(t, bE1.Header)
	assert.Equal(t, expectedNameE, bE1.Header.Name)

	// Check second store only commited bucket should be found
	bE2, err := s2.Get(bE.Header.Uid)
	assert.Equal(t, bucket.ErrNotExist, err)
	assert.Nil(t, bE2)

	// Check Filtering

	// BktA created between t0 & t2
	paginerA1, err := s1.Filter(bucket.OlderFirst, bucket.CreatedBetweenCriterion(t0, t2), 1, 0)
	assert.NoError(t, err)
	require.NotNil(t, paginerA1)
	entriesA1 := iterz.Flatten(paginerA1.All())
	assert.Len(t, entriesA1, 1)
	require.True(t, len(entriesA1) > 0)
	assert.Equal(t, expectedNameA, entriesA1[0].Val().Header.Name)
	txt, err := entriesA1[0].Val().ProjectText(bucket.LatestVersion)
	assert.NoError(t, err)
	assert.Equal(t, expectedTxtA1, txt)

	// No bucket created between t0 & t1 (limits excluded)
	paginerA2, err := s1.Filter(bucket.OlderFirst, bucket.CreatedBetweenCriterion(t0, t1), 1, 0)
	assert.NoError(t, err)
	require.NotNil(t, paginerA1)
	entriesA2 := iterz.Flatten(paginerA2.All())
	assert.Len(t, entriesA2, 0)

	// BktC created between t2 & t4 (BktC: updated)
	paginerC1, err := s1.Filter(bucket.OlderFirst, bucket.CreatedBetweenCriterion(t2, t4), 1, 0)
	assert.NoError(t, err)
	require.NotNil(t, paginerC1)
	entriesC1 := iterz.Flatten(paginerC1.All())
	assert.Len(t, entriesC1, 1)
	require.True(t, len(entriesC1) > 0)
	assert.Equal(t, expectedNameC, entriesC1[0].Val().Header.Name)
	txt, err = entriesC1[0].Val().ProjectText(bucket.LatestVersion)
	assert.NoError(t, err)
	assert.Equal(t, expectedTxtC1+"updated", txt)

	// BktB created between t1 & t3 (BktB: commited then updated)
	paginerB1, err := s1.Filter(bucket.OlderFirst, bucket.CreatedBetweenCriterion(t1, t3), 1, 0)
	assert.NoError(t, err)
	require.NotNil(t, paginerB1)
	entriesB1 := iterz.Flatten(paginerB1.All())
	assert.Len(t, entriesB1, 1)
	require.True(t, len(entriesB1) > 0)
	assert.Equal(t, expectedNameB, entriesB1[0].Val().Header.Name)
	txt, err = entriesB1[0].Val().ProjectText(bucket.LatestVersion)
	assert.NoError(t, err)
	assert.Equal(t, expectedTxtB1+"updated", txt)

	// Refilter to check consistency
	paginerB2, err := s1.Filter(bucket.OlderFirst, bucket.CreatedBetweenCriterion(t1, t3), 1, 0)
	assert.NoError(t, err)
	require.NotNil(t, paginerB2)
	entriesB2 := iterz.Flatten(paginerB2.All())
	assert.Len(t, entriesB2, 1)
	require.True(t, len(entriesB2) > 0)
	assert.Equal(t, expectedNameB, entriesB2[0].Val().Header.Name)
	txt, err = entriesB2[0].Val().ProjectText(bucket.LatestVersion)
	assert.NoError(t, err)
	assert.Equal(t, expectedTxtB1+"updated", txt)

	// Search for updates between t3 and end: buckets C D E
	paginerD1, err := s1.Filter(bucket.OlderFirst, bucket.UpdatedBetweenCriterion(t3, tEnd), 1, 0)
	assert.NoError(t, err)
	require.NotNil(t, paginerD1)
	entriesD1 := iterz.Flatten(paginerD1.All())
	assert.Len(t, entriesD1, 3)
	// First Bucket: C
	require.True(t, len(entriesD1) >= 1)
	assert.Equal(t, expectedNameC, entriesD1[0].Val().Header.Name)
	txt, err = entriesD1[0].Val().ProjectText(bucket.LatestVersion)
	assert.NoError(t, err)
	assert.Equal(t, expectedTxtC1+"updated", txt)
	// Second Bucket: D
	require.True(t, len(entriesD1) >= 2)
	assert.Equal(t, expectedNameD, entriesD1[1].Val().Header.Name)
	txt, err = entriesD1[1].Val().ProjectText(bucket.LatestVersion)
	assert.NoError(t, err)
	assert.Equal(t, expectedTxtD1+"updated", txt)
	// // Tird Bucket: D
	// require.True(t, len(entriesD1) >= 3)
	// assert.Equal(t, expectedNameD, entriesD1[2].Val().Header.Name)
	// txt, err = entriesD1[2].Val().ProjectText(bucket.LatestVersion)
	// assert.NoError(t, err)
	// assert.Equal(t, expectedTxtD1, txt)
	// Tird Bucket: E
	require.True(t, len(entriesD1) >= 3)
	assert.Equal(t, expectedNameE, entriesD1[2].Val().Header.Name)
	txt, err = entriesD1[2].Val().ProjectText(bucket.LatestVersion)
	assert.NoError(t, err)
	assert.Equal(t, expectedTxtE1, txt)

	// Search for updates: all buckets should match in s1
	paginerE1, err := s1.Filter(bucket.OlderFirst, bucket.UpdatedBetweenCriterion(t0, tEnd), 1, 0)
	assert.NoError(t, err)
	require.NotNil(t, paginerE1)
	entriesE1 := iterz.Flatten(paginerE1.All())
	assert.Len(t, entriesE1, 5)

	// Search for updates: commited buckets should match in s2
	paginerE2, err := s2.Filter(bucket.OlderFirst, bucket.UpdatedBetweenCriterion(t0, tEnd), 1, 0)
	assert.NoError(t, err)
	require.NotNil(t, paginerE2)
	entriesE2 := iterz.Flatten(paginerE2.All())
	assert.Len(t, entriesE2, 2)
}
