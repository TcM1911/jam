// Copyright (c) 2018 Joakim Kennedy
// Copyright (c) 2016, 2017 Evgeny Badin

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package storage

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/boltdb/bolt"
	"github.com/stretchr/testify/assert"
)

func Test_FullPathDirectory(t *testing.T) {
	dir := fullDbPath()
	if !strings.HasSuffix(dir, ".local/share/jamsonicdb") {
		t.Fatalf("Invalid directory: %s\n", dir)
	}
}

func TestCredentials(t *testing.T) {
	assert := assert.New(t)
	// Get a temp file for testing database.
	tmpFolder := os.TempDir()
	f, err := ioutil.TempFile(tmpFolder, "jamsonic-test")
	fileName := f.Name()
	f.Close()
	defer os.Remove(fileName)
	if err != nil {
		assert.FailNow("Failed to create a temp file.")
	}

	b, err := bolt.Open(fileName, 0600, nil)
	if err != nil {
		assert.FailNow("Failed to open test database.")
	}
	db := &BoltDB{Bolt: b, LibName: []byte("testLibrary")}

	expectedPassword := []byte("test password")
	credkey := []byte("credKey")
	// Tests
	t.Run("handle_no_creds_saved", func(t *testing.T) {
		_, err := db.GetCredentials(credkey)
		assert.Equal(ErrNoCredentialsStored, err, "Wrong error returned")
	})
	t.Run("save", func(t *testing.T) {
		err := db.SaveCredentials(credkey, expectedPassword)
		assert.NoError(err, "Should save creds without an error.")
	})
	t.Run("retrieve_stored_creds", func(t *testing.T) {
		actual, err := db.GetCredentials(credkey)
		assert.NoError(err, "Should return creds without an error")
		assert.Equal(expectedPassword, actual, "wrong creds returned")
	})
}
