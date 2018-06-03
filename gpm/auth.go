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

package gpm

/*
import (
	"fmt"

	"github.com/TcM1911/jamsonic"
	"github.com/TcM1911/jamsonic/lastfm"
	"github.com/TcM1911/jamsonic/storage"
	"github.com/boltdb/bolt"
	"github.com/budkin/gmusic"
)

func loginFromDatabase(db *bolt.DB) (*gmusic.GMusic, error) {
	auth, deviceID, err := storage.ReadCredentials(db)
	if err != nil {
		return nil, err
	}
	return &gmusic.GMusic{
		Auth:     string(auth),
		DeviceID: string(deviceID),
	}, nil
}

func CheckCreds(d *storage.BoltDB, lastFM *bool) (*Client, *lastfm.Client, string, error) {
	db := d.Bolt
	gm, err := loginFromDatabase(db)
	client := &Client{GMusic: gm}
	if err != nil {
		gm, err = authenticate()
		if err != nil {
			return nil, nil, "", err
		}
		err = storage.WriteCredentials(db, gm.Auth, gm.DeviceID)
		if err != nil {
			return nil, nil, "", err
		}
		fmt.Println("Syncing database, may take a few seconds (will take longer if you have a lot of playlists)")
		err = jamsonic.RefreshLibrary(d, client)
	}
	if err != nil {
		return nil, nil, "", err
	}

	lmclient := lastfm.New(
		string([]byte{0x62, 0x39, 0x30, 0x36, 0x65, 0x62, 0x63, 0x35, 0x39, 0x35, 0x34, 0x63, 0x37, 0x65, 0x63, 0x39, 0x66, 0x39, 0x65, 0x63, 0x64, 0x32, 0x66, 0x66, 0x35, 0x63, 0x30, 0x62, 0x65, 0x33, 0x64, 0x34}),
		string([]byte{0x39, 0x36, 0x66, 0x63, 0x63, 0x33, 0x33, 0x33, 0x33, 0x61, 0x39, 0x61, 0x30, 0x33, 0x37, 0x66, 0x63, 0x65, 0x35, 0x31, 0x65, 0x63, 0x33, 0x62, 0x37, 0x62, 0x34, 0x37, 0x66, 0x66, 0x62, 0x37}))

	sk, err := storage.ReadLastFM(db)
	if *lastFM && err != nil {

		token, err := lmclient.GetToken()
		if err != nil {
			return nil, nil, "", err
		}

		lmclient.LoginWithToken(token)
		sk = lmclient.Api.GetSessionKey()

		err = storage.WriteLastFM([]byte(sk), db)
		if err != nil {
			return nil, nil, "", err
		}
	}

	if sk != "" {
		lmclient.Api.SetSession(sk)
		*lastFM = true
		return client, lmclient, sk, nil

	}

	return client, nil, "", nil
}

func authenticate() (*gmusic.GMusic, error) {
	email := jamsonic.AskForUsername()
	password, err := jamsonic.AskForPassword()
	if err != nil {
		return nil, err
	}
	return gmusic.Login(email, string(password))
}
*/
