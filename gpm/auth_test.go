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

package gpm

/*
import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/boltdb/bolt"
)

// tempfile returns a temporary file path.
func tempfile() (string, error) {
	f, _ := ioutil.TempFile("", "jamsonicdb")
	err := f.Close()
	if err != nil {
		return "", err
	}
	err = os.Remove(f.Name())
	if err != nil {
		return "", err
	}
	return f.Name(), nil
}

func getDatabase(t *testing.T) *bolt.DB {
	tf, err := tempfile()
	if err != nil {
		t.Fatalf("Can't create temporary file: %v", err)
	}
	db, err := bolt.Open(tf, 0600, nil)
	if err != nil {
		t.Fatalf("Can't create BoltDB test database.")
	}
	return db
}

func Test_LoginWithEmptyConfiguration(t *testing.T) {
	db := getDatabase(t)
	_, err := loginFromDatabase(db)
	if err == nil {
		t.Fatalf("Invalid process for auth: %s", err)
	}
}
*/
