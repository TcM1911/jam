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
	logger  *jamsonic.Logger
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

func Open(logger *jamsonic.Logger) (*BoltDB, error) {
	dbPath := fullDbPath()
	logger.DebugLog("Opening database stored at " + dbPath)
	db, err := bolt.Open(dbPath, 0600, nil)
	logger.DebugLog("Database opened")
	return &BoltDB{Bolt: db, logger: logger}, err
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
