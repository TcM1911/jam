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
	"errors"
	"os"
	"path/filepath"
	"runtime"

	"github.com/TcM1911/jamsonic"
	"github.com/boltdb/bolt"
)

const (
	gpmAuthBucket = "#AuthDetails"
	gpmAuthKey    = "Auth"
	gpmDeviceKey  = "DeviceID"
	lastFMBucket  = "#LastFM"
	lastFMRecord  = "sk"
)

var (
	credentialBucket = []byte("Credentials")
)

var (
	ErrNoGPMCredentials = errors.New("No #AuthDetails bucket")
	ErrNoLastFMBucket   = errors.New("No #LastFM bucket")
	ErrNoLastFMRecord   = errors.New("No LastFM record in the database")
)

func fullDbPath() string {
	if runtime.GOOS == "windows" {
		return filepath.Join(os.Getenv("APPDATA"), "jamsonicdb")
	}
	path := filepath.Join(os.Getenv("HOME"), ".local", "share")
	_, err := os.Stat(path)

	if os.IsNotExist(err) {
		os.MkdirAll(path, 0755)
	}
	return filepath.Join(path, "jamsonicdb")
}

type BoltDB struct {
	Bolt    *bolt.DB
	LibName []byte
}

func (d *BoltDB) SaveCredentials(key []byte, credStruct []byte) error {
	err := d.Bolt.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(credentialBucket)
		if err != nil {
			return err
		}
		return b.Put(key, credStruct)
	})
	return err
}

func (d *BoltDB) GetCredentials(key []byte) ([]byte, error) {
	var buf []byte
	err := d.Bolt.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(credentialBucket)
		if b == nil {
			return jamsonic.ErrNoCredentialsStored
		}
		creds := b.Get(key)
		buf = make([]byte, len(creds))
		copy(buf, creds)
		return nil
	})
	return buf, err
}

func Open() (*BoltDB, error) {
	db, err := bolt.Open(fullDbPath(), 0600, nil)
	return &BoltDB{Bolt: db}, err
}

func ReadCredentials(db *bolt.DB) ([]byte, []byte, error) {
	var auth []byte
	var deviceID []byte
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(gpmAuthBucket))
		if b == nil {
			return ErrNoGPMCredentials
		}
		auth = b.Get([]byte(gpmAuthKey))
		deviceID = b.Get([]byte(gpmDeviceKey))
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	return auth, deviceID, nil
}

func WriteCredentials(db *bolt.DB, auth string, deviceID string) error {
	return db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(gpmAuthBucket))
		if err != nil {
			return err
		}
		err = b.Put([]byte(gpmAuthKey), []byte(auth))
		err = b.Put([]byte(gpmDeviceKey), []byte(deviceID))
		return err
	})
}

func ReadLastFM(db *bolt.DB) (string, error) {
	var lastfm []byte
	var err error

	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(lastFMBucket))
		if b == nil {
			return ErrNoLastFMBucket
		}
		lastfm = b.Get([]byte(lastFMRecord))
		if string(lastfm) == "" {
			err = ErrNoLastFMRecord
		}

		return err

	})
	return string(lastfm), err
}

func WriteLastFM(lastfm []byte, db *bolt.DB) error {

	return db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(lastFMBucket))
		if err != nil {
			return err
		}

		err = b.Put([]byte(lastFMRecord), lastfm)
		return err
	})
}
