// Copyright (c) 2018 Joakim Kennedy
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
	"testing"

	"github.com/TcM1911/jamsonic"
	"github.com/boltdb/bolt"
	"github.com/stretchr/testify/assert"
)

func TestArtists(t *testing.T) {
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

	// Mocked data.
	artist1 := &jamsonic.Artist{
		Name: "Artist1",
		ID:   "Artist1",
		Albums: []*jamsonic.Album{
			&jamsonic.Album{Artist: "Artist1", ID: "A1", Year: uint32(1990),
				Tracks: []*jamsonic.Track{
					&jamsonic.Track{Album: "A1", Title: "T1 title"},
					&jamsonic.Track{Album: "A1", Title: "T2 title"},
					&jamsonic.Track{Album: "A1", Title: "T3 title"},
				},
			},
			&jamsonic.Album{Artist: "Artist1", ID: "A2", Year: uint32(1990),
				Tracks: []*jamsonic.Track{
					&jamsonic.Track{Album: "A2", Title: "T1 title"},
					&jamsonic.Track{Album: "A2", Title: "T2 title"},
					&jamsonic.Track{Album: "A2", Title: "T3 title"},
				},
			},
		},
	}
	artist2 := &jamsonic.Artist{
		Name: "Artist2",
		ID:   "Artist2",
		Albums: []*jamsonic.Album{
			&jamsonic.Album{Artist: "Artist2", ID: "A1", Year: uint32(1990),
				Tracks: []*jamsonic.Track{
					&jamsonic.Track{Album: "A1", Title: "T1 title"},
					&jamsonic.Track{Album: "A1", Title: "T2 title"},
					&jamsonic.Track{Album: "A1", Title: "T3 title"},
				},
			},
			&jamsonic.Album{Artist: "Artist2", ID: "A2", Year: uint32(1990),
				Tracks: []*jamsonic.Track{
					&jamsonic.Track{Album: "A2", Title: "T1 title"},
					&jamsonic.Track{Album: "A2", Title: "T2 title"},
					&jamsonic.Track{Album: "A2", Title: "T3 title"},
				},
			},
		},
	}
	artists := []*jamsonic.Artist{artist1, artist2}

	// Tests
	t.Run("save_artists", func(t *testing.T) {
		err := db.SaveArtists(artists)
		assert.NoError(err, "Save artist with no error.")
	})

	t.Run("get_artists", func(t *testing.T) {
		actualArtists, err := db.Artists()
		assert.NoError(err, "Should get artists without an error")
		assert.NotEmpty(actualArtists, "Artists should not be empty")
		assert.Equal(artist1, actualArtists[0], "Should return same A1 values")
		assert.Equal(artist1.Albums[0], actualArtists[0].Albums[0], "Should return same album values")
		assert.Equal(artist1.Albums[0].Tracks[0], actualArtists[0].Albums[0].Tracks[0], "Should return same track values")
		assert.Equal(artist1.Albums[0].Tracks[1], actualArtists[0].Albums[0].Tracks[1], "Should return same track values")
		assert.Equal(artist1.Albums[0].Tracks[2], actualArtists[0].Albums[0].Tracks[2], "Should return same track values")

		assert.Equal(artist2, actualArtists[1], "Should return same A1 values")
		assert.Equal(artist2.Albums[0], actualArtists[1].Albums[0], "Should return same album values")
		assert.Equal(artist2.Albums[0].Tracks[0], actualArtists[1].Albums[0].Tracks[0], "Should return same track values")
		assert.Equal(artist2.Albums[0].Tracks[1], actualArtists[1].Albums[0].Tracks[1], "Should return same track values")
		assert.Equal(artist2.Albums[0].Tracks[2], actualArtists[1].Albums[0].Tracks[2], "Should return same track values")
	})
}
