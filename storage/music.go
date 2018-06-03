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
	"encoding/json"
	"errors"
	"log"
	"sort"
	"strconv"
	"strings"

	"github.com/TcM1911/jamsonic"
	"github.com/boltdb/bolt"
)

var (
	musicLibrary = []byte("MusicLib")
)

var (
	ErrNoLibraryFound  = errors.New("no library found")
	ErrLibraryNotFound = errors.New("library not found")
)

// AddPlaylists stores the playlists in the database.
func (d *BoltDB) AddPlaylists(provider jamsonic.Provider, playlists []*jamsonic.Playlist, entries []*jamsonic.PlaylistEntry) error {
	db := d.Bolt
	var pl *bolt.Bucket
	var track *jamsonic.Track
	var plName string
	var buf []byte
	var count int
	err := db.Update(func(tx *bolt.Tx) error {
		tx.DeleteBucket([]byte("Playlists"))

		pls, err := tx.CreateBucket([]byte("Playlists"))
		if err != nil {
			return err
		}

		for _, entry := range entries {
			track, err = provider.GetTrackInfo(entry.TrackId)
			if err != nil {
				err = nil
				continue
			}

			for _, plEntry := range playlists {
				if entry.PlaylistId == plEntry.ID {
					plName = plEntry.Name
					break
				}
			}

			pl, err = pls.CreateBucketIfNotExists([]byte(plName))
			if err != nil {
				return err
			}

			bt := jamsonic.Track{
				Artist:         track.Artist,
				DiscNumber:     track.DiscNumber,
				TrackNumber:    track.TrackNumber,
				DurationMillis: track.DurationMillis,
				EstimatedSize:  track.EstimatedSize,
				ID:             entry.TrackId,
				PlayCount:      track.PlayCount,
				Title:          track.Title,
				Year:           track.Year,
			}
			buf, err = json.Marshal(bt)
			if err != nil {
				return err
			}

			err = pl.Put([]byte(strconv.Itoa(count)), buf)
			if err != nil {
				return err
			}
			count++
		}

		return err

	})
	return err
}

// GetPlaylists returns the playlists. This code was moved from
// the UI package.
// This code is depricated.
func GetPlaylists(d *BoltDB) sort.StringSlice {
	playlists := sort.StringSlice{}
	d.Bolt.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Playlists"))
		c := b.Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			playlists = append(playlists, string(k))
		}
		return nil
	})
	return playlists
}

// GetArtistsAndAlbums returns the artists and albums. This code was moved from
// the UI package.
// This code is depricated.
func GetArtistsAndAlbums(d *BoltDB) (sort.StringSlice, map[string]bool, map[string][]string) {
	artists := sort.StringSlice{}
	artistsMap := make(map[string]bool)
	albums := make(map[string][]string)
	d.Bolt.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("Library"))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			if !artistsMap[string(k)] {
				artistsMap[string(k)] = false
			}
			if v == nil {
				if err := b.Bucket(k).ForEach(func(kk []byte, vv []byte) error {
					albums[string(k)] = append(albums[string(k)], string(kk))

					return nil
				}); err != nil {
					log.Fatalf("Can't populate artists: %s", err)
				}
			}
		}
		for k := range artistsMap {
			artists = append(artists, k)
		}
		artists.Sort()
		// log.Printf("Artists: %s", app.Artists)
		return nil
	})
	return artists, artistsMap, albums
}

func (d *BoltDB) Artists() ([]*jamsonic.Artist, error) {
	artists := make([]*jamsonic.Artist, 0)
	err := d.Bolt.View(func(tx *bolt.Tx) error {
		mainBucket := tx.Bucket(musicLibrary)
		if mainBucket == nil {
			return ErrNoLibraryFound
		}
		b := mainBucket.Bucket(d.LibName)
		if b == nil {
			return ErrLibraryNotFound
		}
		b.ForEach(func(k []byte, v []byte) error {
			var artist jamsonic.Artist
			err := json.Unmarshal(v, &artist)
			if err != nil {
				return err
			}
			artists = append(artists, &artist)
			return nil
		})
		return nil
	})
	return artists, err
}

func (d *BoltDB) SaveArtists(artists []*jamsonic.Artist) error {
	err := d.Bolt.Update(func(tx *bolt.Tx) error {
		mainBucket, err := tx.CreateBucketIfNotExists(musicLibrary)
		if err != nil {
			return err
		}
		// Remove old cached data.
		if mainBucket.Bucket(d.LibName) != nil {
			err = mainBucket.DeleteBucket(d.LibName)
			if err != nil {
				return err
			}
		}
		b, err := mainBucket.CreateBucket(d.LibName)
		if err != nil {
			return err
		}
		for _, artist := range artists {
			buf, err := json.Marshal(&artist)
			if err != nil {
				return err
			}
			err = b.Put([]byte(artist.ID), buf)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

// GetTracks returns the tracks. This code was moved from
// the UI package.
// This code is depricated.
func GetTracks(d *BoltDB, what []string, view int, curPos, scrOffset map[bool]int, numAlb func(k int) int) map[string][]string {
	songs := make(map[string][]string)
	if err := d.Bolt.View(func(tx *bolt.Tx) error {
		var b *bolt.Bucket
		var c *bolt.Cursor
		if view == 0 {
			i := curPos[false] - 1 + scrOffset[false]
			b = tx.Bucket([]byte("Library")).Bucket([]byte(what[i-numAlb(i)]))
		} else if view == 1 {
			b = tx.Bucket([]byte("Playlists"))
		}
		c = b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			if v == nil {
				cc := b.Bucket(k).Cursor()
				for kk, vv := cc.First(); kk != nil; kk, vv = cc.Next() {
					songs[string(k)] = append(songs[string(k)], string(vv))
				}
			}
		}
		return nil
	}); err != nil {
		log.Fatalf("Can't populate songs: %s", err)
	}
	return songs
}

// AddTracks stores the tracks to the database.
func (d *BoltDB) AddTracks(tracks []*jamsonic.Track) error {
	db := d.Bolt
	var artist *bolt.Bucket
	var mixedAlbum bool
	var buf []byte
	var key string
	err := db.Update(func(tx *bolt.Tx) error {
		tx.DeleteBucket([]byte("Library"))

		lib, err := tx.CreateBucket([]byte("Library"))
		if err != nil {
			return err
		}
		for i := 0; i < len(tracks); i++ {
			temp := i + 1
			for mixedAlbum = false; temp < len(tracks) && tracks[temp].Album == tracks[i].Album; temp++ {
				if tracks[temp].Artist != tracks[i].Artist {
					mixedAlbum = true
				}
			}

			if mixedAlbum {
				artist, err = lib.CreateBucketIfNotExists([]byte("Various Artists"))
				if err != nil {
					return err
				}

				if tracks[i].Album == "" {
					tracks[i].Album = "Unknown Album"
				}
				album, err := artist.CreateBucketIfNotExists([]byte(tracks[i].Album))
				if err != nil {
					return err
				}
				for i < temp {
					bt := jamsonic.Track{
						Artist:         tracks[i].Artist,
						DiscNumber:     tracks[i].DiscNumber,
						TrackNumber:    tracks[i].TrackNumber,
						DurationMillis: tracks[i].DurationMillis,
						EstimatedSize:  tracks[i].EstimatedSize,
						ID:             tracks[i].ID,
						PlayCount:      tracks[i].PlayCount,
						Title:          tracks[i].Title,
						Year:           tracks[i].Year,
					}
					buf, err = json.Marshal(bt)
					if err != nil {
						return err
					}

					if tracks[i].TrackNumber < 10 {
						key = strconv.Itoa(int(tracks[i].DiscNumber)) + "0" + strconv.Itoa(int(tracks[i].TrackNumber))
					} else {
						key = strconv.Itoa(int(tracks[i].DiscNumber)) + strconv.Itoa(int(tracks[i].TrackNumber))
					}

					err = album.Put([]byte(key), buf)
					if err != nil {
						return err
					}
					i++
				}
				continue
			} else {
				if tracks[i].Artist == "" {
					if tracks[i].AlbumArtist == "" {
						tracks[i].Artist = "Unknown Artist"
					} else {
						tracks[i].Artist = tracks[i].AlbumArtist
					}
				}
				tracks[i].Artist = strings.Title(strings.ToLower(tracks[i].Artist))
				artist, err = lib.CreateBucketIfNotExists([]byte(tracks[i].Artist))
				if err != nil {
					return err
				}
			}
			if tracks[i].Album == "" {
				tracks[i].Album = "Unknown Album"
			}
			album, err := artist.CreateBucketIfNotExists([]byte(tracks[i].Album))
			if err != nil {
				return err
			}

			bt := jamsonic.Track{
				Artist:         tracks[i].Artist,
				DiscNumber:     tracks[i].DiscNumber,
				TrackNumber:    tracks[i].TrackNumber,
				DurationMillis: tracks[i].DurationMillis,
				EstimatedSize:  tracks[i].EstimatedSize,
				ID:             tracks[i].ID,
				PlayCount:      tracks[i].PlayCount,
				Title:          tracks[i].Title,
				Year:           tracks[i].Year,
			}
			buf, err = json.Marshal(bt)
			if err != nil {
				return err
			}

			if tracks[i].TrackNumber < 10 {
				key = strconv.Itoa(int(tracks[i].DiscNumber)) + "0" + strconv.Itoa(int(tracks[i].TrackNumber))
			} else {
				key = strconv.Itoa(int(tracks[i].DiscNumber)) + strconv.Itoa(int(tracks[i].TrackNumber))
			}

			err = album.Put([]byte(key), buf)
			if err != nil {
				return err
			}
		}

		return nil
	})
	return err
}
