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

package music

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/TcM1911/jamsonic"
	"github.com/boltdb/bolt"
)

type BTrack struct {
	Artist         string
	DiscNumber     uint8
	TrackNumber    uint32
	DurationMillis string
	EstimatedSize  string
	ID             string
	PlayCount      uint32
	Title          string
	Year           uint32
}

func RefreshLibrary(db *bolt.DB, provider jamsonic.Provider) error {
	//db, err := bolt.Open(fullDbPath(), 0600, nil)
	//checkErr(err)
	//defer db.Close()
	var err error
	tracks, err := provider.ListTracks()
	if err != nil {
		return err
	}
	playlists, err := provider.ListPlaylists()
	if err != nil {
		return err
	}
	entries, err := provider.ListPlaylistEntries()
	if err != nil {
		return err
	}
	err = addTracks(db, tracks)
	if err != nil {
		return err
	}
	err = addPlaylists(db, provider, playlists, entries)
	if err != nil {
		return err
	}
	return err
}

func addPlaylists(db *bolt.DB, provider jamsonic.Provider, playlists []*jamsonic.Playlist, entries []*jamsonic.PlaylistEntry) error {
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

			bt := BTrack{track.Artist, track.DiscNumber, track.TrackNumber, track.DurationMillis,
				track.EstimatedSize, entry.TrackId, track.PlayCount, track.Title, track.Year}
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
func addTracks(db *bolt.DB, tracks []*jamsonic.Track) error {
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
					bt := BTrack{tracks[i].Artist, tracks[i].DiscNumber, tracks[i].TrackNumber, tracks[i].DurationMillis,
						tracks[i].EstimatedSize, tracks[i].ID, tracks[i].PlayCount, tracks[i].Title, tracks[i].Year}
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

			bt := BTrack{tracks[i].Artist, tracks[i].DiscNumber, tracks[i].TrackNumber, tracks[i].DurationMillis,
				tracks[i].EstimatedSize, tracks[i].ID, tracks[i].PlayCount, tracks[i].Title, tracks[i].Year}
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
