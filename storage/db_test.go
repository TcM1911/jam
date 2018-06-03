// Copyright (c) 2018 Joakim Kennedy
// Copyright (c) 2016, 2017 Evgeny Badin
//
// This file is part of Jamsonic.
//
// Jamsonic is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Jamsonic is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Jamsonic.  If not, see <http://www.gnu.org/licenses/>.

package storage

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/TcM1911/jamsonic"
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
		assert.Equal(jamsonic.ErrNoCredentialsStored, err, "Wrong error returned")
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
