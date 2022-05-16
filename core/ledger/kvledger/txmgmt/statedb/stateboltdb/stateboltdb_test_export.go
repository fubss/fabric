/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package stateboltdb

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestVDBEnv provides a bolt db backed versioned db for testing
type TestVDBEnv struct {
	t          testing.TB
	DBProvider *VersionedDBProvider
	dbPath     string
}

// NewTestVDBEnv instantiates and new bolt db backed TestVDB
func NewTestVDBEnv(t testing.TB) *TestVDBEnv {
	t.Logf("Creating new TestVDBEnv")
	dbPath, err := ioutil.TempDir("", "statebltdb")
	if err != nil {
		t.Fatalf("Failed to create boltdb directory: %s", err)
	}
	dbProvider, err := NewVersionedDBProvider(dbPath)
	require.NoError(t, err)
	return &TestVDBEnv{t, dbProvider, dbPath}
}

// Cleanup closes the db and removes the db folder
func (env *TestVDBEnv) Cleanup() {
	env.t.Logf("Cleaningup TestVDBEnv")
	env.DBProvider.Close()
	os.RemoveAll(env.dbPath)
}
