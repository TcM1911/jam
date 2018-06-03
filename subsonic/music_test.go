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

package subsonic

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/TcM1911/jamsonic"
	"github.com/stretchr/testify/assert"
)

func TestGetStream(t *testing.T) {
	assert := assert.New(t)
	stream := []byte("stream of data")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.RequestURI, "id=codefail") {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if strings.Contains(r.RequestURI, "id=stream") {
			w.WriteHeader(http.StatusOK)
			w.Write(stream)
			return
		}
	}))
	c := &Client{Credentials: Credentials{
		Username: "username",
		Host:     ts.URL,
	},
	}

	t.Run("handle_error_code", func(t *testing.T) {
		_, err := c.GetStream("codefail")
		assert.Error(err, "Return error if 200 ok is not returned.")
		assert.Equal("400 Bad Request", err.Error(), "Wrong error message")
	})
	t.Run("get_stream", func(t *testing.T) {
		b, err := c.GetStream("stream")
		assert.NoError(err, "Should not return an error if a stream is returned.")
		actual, _ := ioutil.ReadAll(b)
		b.Close()
		assert.Equal(stream, actual, "Wrong stream returned.")
	})
	t.Run("request_error", func(t *testing.T) {
		c.Credentials.Host = "http://localhost:-8080"
		b, err := c.GetStream("empty")
		assert.Error(err, "Should return an error")
		assert.Nil(b, "No reader should be returned if it's empty")
	})
}

func TestProviderType(t *testing.T) {
	c := &Client{}
	assert.Equal(t, jamsonic.SubSonic, c.GetProvider(), "Wrong provider type returned.")
}

func TestFetchLibrary(t *testing.T) {
	assert := assert.New(t)
	// Mocked data
	a1 := &artist{
		Name: "A1", ID: "A1",
		Albums: []*album{
			&album{ID: "AA1",
				Songs: []*song{
					&song{ID: "AA11"},
					&song{ID: "AA12"},
				},
			},
			&album{ID: "AA2",
				Songs: []*song{
					&song{ID: "AA21"},
					&song{ID: "AA22"},
				},
			},
		},
	}
	a2 := &artist{
		Name: "A2", ID: "A2",
		Albums: []*album{
			&album{ID: "AB1",
				Songs: []*song{
					&song{ID: "AB11"},
					&song{ID: "AB12"},
				},
			},
			&album{ID: "AB2",
				Songs: []*song{
					&song{ID: "AB21"},
					&song{ID: "AB22"},
				},
			},
		},
	}
	a3 := &artist{Name: "A3", ID: "A3"}

	/// Mock server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Get Artists
		if strings.Contains(r.RequestURI, "getArtists.view") {
			writeServerReply(w, &apiData{Response: apiResponse{
				ArtistList: artistList{
					Index: []index{index{Artists: []*artist{a1, a2, a3}}}}}})
			return
		}

		/// Get albums
		if strings.Contains(r.RequestURI, "getArtist.view") &&
			strings.Contains(r.RequestURI, "&id=A1") {
			writeServerReply(w, &apiData{Response: apiResponse{Artist: *a1}})
			return
		}
		if strings.Contains(r.RequestURI, "getArtist.view") &&
			strings.Contains(r.RequestURI, "&id=A2") {
			writeServerReply(w, &apiData{Response: apiResponse{Artist: *a2}})
			return
		}

		// Get tracks
		if strings.Contains(r.RequestURI, "getAlbum.view") &&
			strings.Contains(r.RequestURI, "&id=AA1") {
			writeServerReply(w, &apiData{Response: apiResponse{Album: *a1.Albums[0]}})
			return
		}
		if strings.Contains(r.RequestURI, "getAlbum.view") &&
			strings.Contains(r.RequestURI, "&id=AA2") {
			writeServerReply(w, &apiData{Response: apiResponse{Album: *a1.Albums[1]}})
			return
		}
		if strings.Contains(r.RequestURI, "getAlbum.view") &&
			strings.Contains(r.RequestURI, "&id=AB1") {
			writeServerReply(w, &apiData{Response: apiResponse{Album: *a2.Albums[0]}})
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte{0x0, 0x1})
	}))
	c := &Client{
		Credentials: Credentials{
			Username: "username",
			Host:     ts.URL,
		},
		logger: jamsonic.DefaultLogger(),
	}
	t.Run("get_lib", func(t *testing.T) {
		l, err := c.FetchLibrary()
		assert.NoError(err, "Should not return an error on success.")
		assert.Len(l, 2, "Should return 2 artists")
		for _, v := range l {
			if v.ID == a1.ID {
				assert.Len(v.Albums, 2, "Wrong number of albums returned for A1.")
				assert.Equal(a1.Albums[0].ID, v.Albums[0].ID, "Wrong album ID")
				assert.Equal(a1.Albums[1].ID, v.Albums[1].ID, "Wrong album ID")
				assert.Equal(a1.Albums[0].Songs[0].ID, v.Albums[0].Tracks[0].ID, "Wrong track ID")
				assert.Equal(a1.Albums[0].Songs[1].ID, v.Albums[0].Tracks[1].ID, "Wrong track ID")
				assert.Equal(a1.Albums[1].Songs[0].ID, v.Albums[1].Tracks[0].ID, "Wrong track ID")
				assert.Equal(a1.Albums[1].Songs[1].ID, v.Albums[1].Tracks[1].ID, "Wrong track ID")
			} else if v.ID == a2.ID {
				assert.Len(v.Albums, 2, "Wrong number of albums returned for A2.")
				assert.Equal(a2.Albums[0].ID, v.Albums[0].ID, "Wrong album ID")
				assert.Equal(a2.Albums[1].ID, v.Albums[1].ID, "Wrong album ID")
				assert.Equal(a2.Albums[0].Songs[0].ID, v.Albums[0].Tracks[0].ID, "Wrong track ID")
				assert.Equal(a2.Albums[0].Songs[1].ID, v.Albums[0].Tracks[1].ID, "Wrong track ID")
				assert.Len(v.Albums[1].Tracks, 0, "Wrong number of tracks returned for A2.")
			} else {
				// Should not reach here.
				assert.Fail("unknown artist returned: " + v.ID)
			}
		}

	})
	t.Run("handle_hard_fail", func(t *testing.T) {
		c.Credentials.Host = "http://localhost:-8080"
		l, err := c.FetchLibrary()
		assert.Nil(l, "Should return nil if failed")
		assert.Error(err, "Should return an error on hard fail.")
	})
}

func writeServerReply(w http.ResponseWriter, data *apiData) {
	buf, _ := json.Marshal(&data)
	w.Write(buf)
}

func TestPanics(t *testing.T) {
	assert := assert.New(t)
	c := &Client{}
	assert.PanicsWithValue("should not be called.", func() { _, _ = c.ListTracks() }, "Method should panic.")
	assert.PanicsWithValue("not implemented", func() { _, _ = c.ListPlaylists() }, "Method should panic.")
	assert.PanicsWithValue("should not be called.", func() { _, _ = c.ListPlaylistEntries() }, "Method should panic.")
	assert.PanicsWithValue("should not be called.", func() { _, _ = c.GetTrackInfo("") }, "Method should panic.")
}
