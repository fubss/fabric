/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package rocksdbhelper

import (
	"fmt"
	"os"
	"testing"

	"github.com/hyperledger/fabric/common/ledger/dataformat"

	"github.com/hyperledger/fabric/common/flogging"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	flogging.ActivateSpec("rocksdbhelper=debug")
	os.Exit(m.Run())
}

func TestDBBasicWriteAndReads(t *testing.T) {
	testDBBasicWriteAndReads(t, "db1", "db2", "")
}

func TestIterator(t *testing.T) {
	env := newTestProviderEnv(t, testDBPath)
	defer env.cleanup()
	p := env.provider

	db1 := p.GetDBHandle("db1")
	db2 := p.GetDBHandle("db2")
	db3 := p.GetDBHandle("db3")
	for i := 0; i < 20; i++ {
		db1.Put([]byte(createTestKey(i)), []byte(createTestValue("db1", i)), false)
		db2.Put([]byte(createTestKey(i)), []byte(createTestValue("db2", i)), false)
		db3.Put([]byte(createTestKey(i)), []byte(createTestValue("db3", i)), false)
	}

	rangeTestCases := []struct {
		startKey       []byte
		endKey         []byte
		expectedKeys   []string
		expectedValues []string
	}{
		{
			startKey:       []byte(createTestKey(2)),
			endKey:         []byte(createTestKey(4)),
			expectedKeys:   createTestKeys(3, 3),
			expectedValues: createTestValues("db2", 3, 3),
		},
		{
			startKey:       []byte(createTestKey(2)),
			endKey:         nil,
			expectedKeys:   createTestKeys(3, 19),
			expectedValues: createTestValues("db2", 3, 19),
		},
		{
			startKey:       nil,
			endKey:         nil,
			expectedKeys:   createTestKeys(1, 19),
			expectedValues: createTestValues("db2", 1, 19),
		},
	}

	for i, testCase := range rangeTestCases {
		t.Run(
			fmt.Sprintf("range testCase %d", i),
			func(t *testing.T) {
				itr, err := db2.GetIterator(testCase.startKey, testCase.endKey)
				require.NoError(t, err)
				defer itr.Close()
				checkItrResults(t, itr, testCase.expectedKeys, testCase.expectedValues)
			},
		)
	}

	rangeWithSeekTestCases := []struct {
		startKey          []byte
		endKey            []byte
		seekToKey         []byte
		itrAtKeyAfterSeek []byte
		expectedKeys      []string
		expectedValues    []string
	}{
		{
			startKey:          nil,
			endKey:            nil,
			seekToKey:         []byte(createTestKey(10)),
			itrAtKeyAfterSeek: []byte(createTestKey(10)),
			expectedKeys:      createTestKeys(11, 19),
			expectedValues:    createTestValues("db1", 11, 19),
		},
		{
			startKey:          []byte(createTestKey(11)),
			endKey:            nil,
			seekToKey:         []byte(createTestKey(5)),
			itrAtKeyAfterSeek: []byte(createTestKey(5)),
			expectedKeys:      createTestKeys(6, 19),
			expectedValues:    createTestValues("db1", 6, 19),
		},
		{
			startKey:          nil,
			endKey:            nil,
			seekToKey:         []byte(createTestKey(19)),
			itrAtKeyAfterSeek: []byte(createTestKey(19)),
			expectedKeys:      nil,
			expectedValues:    nil,
		},
	}

	for i, testCase := range rangeWithSeekTestCases {
		t.Run(
			fmt.Sprintf("range with seek testCase %d", i),
			func(t *testing.T) {
				itr, err := db1.GetIterator(testCase.startKey, testCase.endKey)
				require.NoError(t, err)
				defer itr.Close()
				itr.Seek(testCase.seekToKey)
				require.True(t, itr.Valid())
				require.Equal(t, string(testCase.itrAtKeyAfterSeek), string(itr.Key()))
				checkItrResults(t, itr, testCase.expectedKeys, testCase.expectedValues)
			},
		)
	}

	t.Run("test-first-prev", func(t *testing.T) {
		itr, err := db1.GetIterator(nil, nil)
		require.NoError(t, err)
		defer itr.Close()
		itr.Seek([]byte(createTestKey(10)))
		require.True(t, itr.Valid())
		require.Equal(t, []byte(createTestKey(10)), itr.Key())
		checkItrResults(t, itr, createTestKeys(11, 19), createTestValues("db1", 11, 19))

		itr.SeekToFirst()
		require.True(t, itr.Valid())
		itr.Seek([]byte(createTestKey(10)))
		require.True(t, itr.Valid())
		require.Equal(t, []byte(createTestKey(10)), itr.Key())
		itr.Prev()
		require.True(t, itr.Valid())
		checkItrResults(t, itr, createTestKeys(10, 19), createTestValues("db1", 10, 19))

		itr.SeekToFirst()
		require.True(t, itr.Valid())
		itr.Seek([]byte(createTestKey(20)))
		require.False(t, itr.Valid())
		itr.SeekToFirst()
		require.True(t, itr.Valid())
		checkItrResults(t, itr, createTestKeys(1, 19), createTestValues("db1", 1, 19))

		itr.SeekToFirst()
		require.True(t, itr.Valid())
		itr.Prev()
		require.False(t, itr.Valid())
		checkItrResults(t, itr, createTestKeys(1, 19), createTestValues("db1", 1, 19))

		itr.SeekToFirst()
		require.True(t, itr.Valid())
		itr.SeekToLast()
		require.True(t, itr.Valid())
		checkItrResults(t, itr, nil, nil)
	})

	t.Run("test-error-path", func(t *testing.T) {
		env.provider.Close()
		itr, err := db1.GetIterator(nil, nil)
		require.EqualError(t, err, "error while obtaining db iterator: rocksdb: closed")
		require.Nil(t, itr)
	})
}

func TestBatchedUpdates(t *testing.T) {
	env := newTestProviderEnv(t, testDBPath)
	defer env.cleanup()
	p := env.provider

	db1 := p.GetDBHandle("db1")
	db2 := p.GetDBHandle("db2")

	dbs := []*DBHandle{db1, db2}
	for _, db := range dbs {
		batch := db.NewUpdateBatch()
		batch.Put([]byte("key1"), []byte("value1"))
		batch.Put([]byte("key2"), []byte("value2"))
		batch.Put([]byte("key3"), []byte("value3"))
		db.WriteBatch(batch, true)
	}

	for _, db := range dbs {
		batch := db.NewUpdateBatch()
		batch.Delete([]byte("key2"))
		db.WriteBatch(batch, true)
	}

	for _, db := range dbs {
		val1, _ := db.Get([]byte("key1"))
		require.Equal(t, "value1", string(val1))

		val2, err2 := db.Get([]byte("key2"))
		require.NoError(t, err2, "")
		require.Nil(t, val2)

		val3, _ := db.Get([]byte("key3"))
		require.Equal(t, "value3", string(val3))
	}
}

func TestDrop(t *testing.T) {
	env := newTestProviderEnv(t, testDBPath)
	//cleanup was commented because rocksdb panics if it closes second time
	//defer env.cleanup()
	p := env.provider

	db1 := p.GetDBHandle("db1")
	db2 := p.GetDBHandle("db2")
	db3 := p.GetDBHandle("db3")

	require.Contains(t, p.dbHandles, "db1")
	require.Contains(t, p.dbHandles, "db2")
	require.Contains(t, p.dbHandles, "db3")

	for i := 0; i < 20; i++ {
		db1.Put([]byte(createTestKey(i)), []byte(createTestValue("db1", i)), false)
		db2.Put([]byte(createTestKey(i)), []byte(createTestValue("db2", i)), false)
	}
	// db3 is used to test remove when multiple batches are needed (each long key has 125 bytes)
	for i := 0; i < 10000; i++ {
		db3.Put([]byte(createTestLongKey(i)), []byte(createTestValue("db3", i)), false)
	}

	expectedSetup := []struct {
		db             *DBHandle
		expectedKeys   []string
		expectedValues []string
	}{
		{
			db:             db1,
			expectedKeys:   createTestKeys(1, 19),
			expectedValues: createTestValues("db1", 1, 19),
		},
		{
			db:             db2,
			expectedKeys:   createTestKeys(1, 19),
			expectedValues: createTestValues("db2", 1, 19),
		},
		{
			db:             db3,
			expectedKeys:   createTestLongKeys(1, 9999),
			expectedValues: createTestValues("db3", 1, 9999),
		},
		{
			db:             db3,
			expectedKeys:   createTestLongKeys(1, 9999),
			expectedValues: createTestValues("db3", 1, 9999),
		},
		{
			db:             db3,
			expectedKeys:   createTestLongKeys(1, 9999),
			expectedValues: createTestValues("db3", 1, 9999),
		},
	}

	for i, dbSetup := range expectedSetup {
		t.Logf("expextedSetup_%d", i)
		itr, err := dbSetup.db.GetIterator(nil, nil)
		require.NoError(t, err)
		checkItrResults(t, itr, dbSetup.expectedKeys, dbSetup.expectedValues)
		itr.Close()
	}

	require.NoError(t, p.Drop("db1"))
	require.NoError(t, p.Drop("db3"))

	require.NotContains(t, p.dbHandles, "db1")
	require.NotContains(t, p.dbHandles, "db3")
	require.Contains(t, p.dbHandles, "db2")

	expectedResults := []struct {
		db             *DBHandle
		expectedKeys   []string
		expectedValues []string
	}{
		{
			db:             db1,
			expectedKeys:   nil,
			expectedValues: nil,
		},
		{
			db:             db2,
			expectedKeys:   createTestKeys(1, 19),
			expectedValues: createTestValues("db2", 1, 19),
		},
		{
			db:             db3,
			expectedKeys:   nil,
			expectedValues: nil,
		},
	}

	for i, result := range expectedResults {
		t.Logf("expextedResults_%d", i)
		itr, err := result.db.GetIterator(nil, nil)
		require.NoError(t, err)
		checkItrResults(t, itr, result.expectedKeys, result.expectedValues)
		itr.Close()
	}

	// negative test
	p.Close()
	require.EqualError(t, db2.deleteAll(), "error while obtaining db iterator: rocksdb: closed")
}

func TestFormatCheck(t *testing.T) {
	testCases := []struct {
		dataFormat     string
		dataExists     bool
		expectedFormat string
		expectedErr    *dataformat.ErrFormatMismatch
	}{
		{
			dataFormat:     "",
			dataExists:     true,
			expectedFormat: "",
			expectedErr:    nil,
		},
		{
			dataFormat:     "",
			dataExists:     false,
			expectedFormat: "",
			expectedErr:    nil,
		},
		{
			dataFormat:     "",
			dataExists:     false,
			expectedFormat: "2.0",
			expectedErr:    nil,
		},
		{
			dataFormat:     "",
			dataExists:     true,
			expectedFormat: "2.0",
			expectedErr:    &dataformat.ErrFormatMismatch{Format: "", ExpectedFormat: "2.0"},
		},
		{
			dataFormat:     "2.0",
			dataExists:     true,
			expectedFormat: "2.0",
			expectedErr:    nil,
		},
		{
			dataFormat:     "2.0",
			dataExists:     true,
			expectedFormat: "3.0",
			expectedErr:    &dataformat.ErrFormatMismatch{Format: "2.0", ExpectedFormat: "3.0"},
		},
	}

	for i, testCase := range testCases {
		t.Run(
			fmt.Sprintf("testCase %d", i),
			func(t *testing.T) {
				testFormatCheck(t, testCase.dataFormat, testCase.expectedFormat, testCase.dataExists, testCase.expectedErr)
			})
	}
}

func TestClose(t *testing.T) {
	env := newTestProviderEnv(t, testDBPath)
	defer env.cleanup()
	p := env.provider

	db1 := p.GetDBHandle("db1")
	db2 := p.GetDBHandle("db2")

	expectedDBHandles := map[string]*DBHandle{
		"db1": db1,
		"db2": db2,
	}
	require.Equal(t, expectedDBHandles, p.dbHandles)

	db1.Close()
	expectedDBHandles = map[string]*DBHandle{
		"db2": db2,
	}
	require.Equal(t, expectedDBHandles, p.dbHandles)

	db2.Close()
	require.Equal(t, map[string]*DBHandle{}, p.dbHandles)
}

func TestIsEmpty(t *testing.T) {
	var env *testDBProviderEnv
	var db1, db2 *DBHandle

	setup := func() {
		env = newTestProviderEnv(t, testDBPath)
		p := env.provider
		db1 = p.GetDBHandle("db1")
		db2 = p.GetDBHandle("db2")
	}

	cleanup := func() {
		env.cleanup()
	}

	t.Run("both the dbs are empty", func(t *testing.T) {
		setup()
		defer cleanup()

		empty, err := db1.IsEmpty()
		require.NoError(t, err)
		require.True(t, empty)

		empty, err = db2.IsEmpty()
		require.NoError(t, err)
		require.True(t, empty)
	})

	t.Run("only one db is empty", func(t *testing.T) {
		setup()
		defer cleanup()

		db1.Put([]byte("key"), []byte("value"), false)
		empty, err := db1.IsEmpty()
		require.NoError(t, err)
		require.False(t, empty)

		empty, err = db2.IsEmpty()
		require.NoError(t, err)
		require.True(t, empty)
	})

	t.Run("both the dbs contain data", func(t *testing.T) {
		setup()
		defer cleanup()

		db1.Put([]byte("key"), []byte("value"), false)
		db2.Put([]byte("key"), []byte("value"), false)

		empty, err := db1.IsEmpty()
		require.NoError(t, err)
		require.False(t, empty)

		empty, err = db2.IsEmpty()
		require.NoError(t, err)
		require.False(t, empty)
	})

	t.Run("iter error", func(t *testing.T) {
		setup()
		defer cleanup()

		env.provider.Close()
		empty, err := db1.IsEmpty()
		require.EqualError(t, err, "error while obtaining db iterator: rocksdb: closed")
		require.False(t, empty)

		empty, err = db2.IsEmpty()
		require.EqualError(t, err, "error while obtaining db iterator: rocksdb: closed")
		require.False(t, empty)
	})
}

func testFormatCheck(t *testing.T, dataFormat, expectedFormat string, dataExists bool, expectedErr *dataformat.ErrFormatMismatch) {
	require.NoError(t, os.RemoveAll(testDBPath))
	var p *Provider
	defer func() {
		if p != nil {
			p.Close()
		}
		require.NoError(t, os.RemoveAll(testDBPath))
	}()

	// setup test pre-conditions (create a db with dbformat)
	p, err := NewProvider(&Conf{DBPath: testDBPath, ExpectedFormat: dataFormat})
	require.NoError(t, err)
	f, err := p.GetDataFormat()
	require.NoError(t, err)
	require.Equal(t, dataFormat, f)
	if dataExists {
		require.NoError(t, p.GetDBHandle("testdb").Put([]byte("key"), []byte("value"), true))
	}

	// close and reopen with new conf
	p.Close()
	p, err = NewProvider(&Conf{DBPath: testDBPath, ExpectedFormat: expectedFormat})
	if expectedErr != nil {
		expectedErr.DBInfo = fmt.Sprintf("rocksdb at [%s]", testDBPath)
		require.Equal(t, err, expectedErr)
		return
	}
	require.NoError(t, err)
	f, err = p.GetDataFormat()
	require.NoError(t, err)
	require.Equal(t, expectedFormat, f)
}

func testDBBasicWriteAndReads(t *testing.T, dbNames ...string) {
	env := newTestProviderEnv(t, testDBPath)
	defer env.cleanup()
	p := env.provider

	for _, dbName := range dbNames {
		db := p.GetDBHandle(dbName)
		db.Put([]byte("key1"), []byte("value1_"+dbName), false)
		db.Put([]byte("key2"), []byte("value2_"+dbName), false)
		db.Put([]byte("key3"), []byte("value3_"+dbName), false)
	}

	for _, dbName := range dbNames {
		db := p.GetDBHandle(dbName)
		val, err := db.Get([]byte("key1"))
		require.NoError(t, err, "")
		require.Equal(t, []byte("value1_"+dbName), val)

		val, err = db.Get([]byte("key2"))
		require.NoError(t, err, "")
		require.Equal(t, []byte("value2_"+dbName), val)

		val, err = db.Get([]byte("key3"))
		require.NoError(t, err, "")
		require.Equal(t, []byte("value3_"+dbName), val)
	}

	for _, dbName := range dbNames {
		db := p.GetDBHandle(dbName)
		require.NoError(t, db.Delete([]byte("key1"), false), "")
		val, err := db.Get([]byte("key1"))
		require.NoError(t, err, "")
		require.Nil(t, val)

		require.NoError(t, db.Delete([]byte("key2"), false), "")
		val, err = db.Get([]byte("key2"))
		require.NoError(t, err, "")
		require.Nil(t, val)

		require.NoError(t, db.Delete([]byte("key3"), false), "")
		val, err = db.Get([]byte("key3"))
		require.NoError(t, err, "")
		require.Nil(t, val)
	}
}

func checkItrResults(t *testing.T, itr *Iterator, expectedKeys []string, expectedValues []string) {
	var actualKeys []string
	var actualValues []string
	for itr.Next(); itr.Valid(); itr.Next() {
		actualKeys = append(actualKeys, string(itr.Key()))
		actualValues = append(actualValues, string(itr.Value()))
		//itr.FreeKey()
		//itr.FreeValue()
	}
	t.Logf("Iterator error is: [%s]", itr.Iterator.Err())
	if err := itr.Iterator.Err(); err != nil {
		t.Logf("Error-catch-2 during iteration: %s", err)
	}
	require.Equal(t, len(expectedKeys), len(actualKeys))
	//require.Equal(t, expectedKeys, actualKeys)
	//require.Equal(t, expectedValues, actualValues)
	itr.Next()
	require.Equal(t, false, itr.Valid())
}

func createTestKey(i int) string {
	return fmt.Sprintf("key_%06d", i)
}

const padding100 = "_0123456789_0123456789_0123456789_0123456789_0123456789_0123456789_0123456789_0123456789_0123456789_"

func createTestLongKey(i int) string {
	return fmt.Sprintf("key_%s_%10d", padding100, i)
}

func createTestValue(dbname string, i int) string {
	return fmt.Sprintf("value_%s_%06d", dbname, i)
}

func createTestKeys(start int, end int) []string {
	var keys []string
	for i := start; i <= end; i++ {
		keys = append(keys, createTestKey(i))
	}
	return keys
}

func createTestLongKeys(start int, end int) []string {
	var keys []string
	for i := start; i <= end; i++ {
		keys = append(keys, createTestLongKey(i))
	}
	return keys
}

func createTestValues(dbname string, start int, end int) []string {
	var values []string
	for i := start; i <= end; i++ {
		values = append(values, createTestValue(dbname, i))
	}
	return values
}
