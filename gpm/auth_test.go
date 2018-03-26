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

package gpm

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
