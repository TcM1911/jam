/*
Copyright (c) 2018 Joakim Kennedy

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package subsonic

import (
	"errors"
	"io"
	"net/http"
	"strconv"
	"sync"

	"github.com/TcM1911/jamsonic"
)

// ListTracks is an old API and is not implemeted for this provider.
// Instead, FetchLibrary should be used.
func (c *Client) ListTracks() ([]*jamsonic.Track, error) {
	panic("should not be called.")
}

// ListPlaylists should return the playlists. Currently this feature
// is not implemented for this provider.
func (c *Client) ListPlaylists() ([]*jamsonic.Playlist, error) {
	panic("not implemented")
}

// ListPlaylistEntries is an old API and is not implemented for this provider.
func (c *Client) ListPlaylistEntries() ([]*jamsonic.PlaylistEntry, error) {
	panic("should not be called.")
}

// GetTrackInfo is an old API and is not implemented for this provider.
func (c *Client) GetTrackInfo(trackID string) (*jamsonic.Track, error) {
	panic("should not be called.")
}

// GetStream returns a ReadCloser stream of the track. The audio is encoded
// as a MP3.
func (c *Client) GetStream(songID string) (io.ReadCloser, error) {
	url := c.makeRequestURL("stream") + "&format=mp3&id=" + songID
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}
	return resp.Body, nil
}

// GetProvider returns the provider identifier.
func (c *Client) GetProvider() jamsonic.MusicProvider {
	return jamsonic.SubSonic
}

// FetchLibrary gets the library from the server.
// Library structure:
//		Artist1{
//			Album1{
//				Track1, Track2,...
//			},
//			Album2{...}
//		}
//		Artist2{...}
func (c *Client) FetchLibrary() ([]*jamsonic.Artist, error) {
	artists, err := getAllArtists(c)
	if err != nil {
		return nil, err
	}

	jobs := make(chan *artist)
	results := make(chan *jamsonic.Artist)
	var wgroup sync.WaitGroup

	for i := 0; i < 10; i++ {
		wgroup.Add(1)
		go getArtistSongs(c, jobs, results, &wgroup)
	}

	// Observer ensures result channels is closed when done processing.
	go func() {
		wgroup.Wait()
		close(results)
	}()

	go func() {
		for _, a := range artists {
			jobs <- a
		}
		close(jobs)
	}()

	as := make([]*jamsonic.Artist, 0)
	for a := range results {
		as = append(as, a)
	}
	return as, nil
}

func getArtistSongs(c *Client, ajob <-chan *artist, result chan<- *jamsonic.Artist, wg *sync.WaitGroup) {
	defer wg.Done()
	logger := c.logger
	for a := range ajob {
		logger.InfoLog("Downloading tracks for " + a.Name)
		albumRes, err := getArtistAlbums(c, a.ID)
		albums := make([]*jamsonic.Album, len(albumRes))
		if err != nil {
			logger.ErrorLog("Failed to process " + a.Name)
			continue
		}
		for k, album := range albumRes {
			logger.DebugLog("Processing " + album.Name)
			songs, err := getAlbumSongs(c, album.ID)
			if err != nil {
				logger.ErrorLog("Failed to process " + album.Name)
			}
			tracks := make([]*jamsonic.Track, len(songs))
			for i, v := range songs {
				tracks[i] = &jamsonic.Track{
					Title:          v.Title,
					ID:             v.ID,
					TrackNumber:    uint32(v.Track),
					DiscNumber:     uint8(v.DiscNumber),
					Year:           uint32(v.Year),
					DurationMillis: strconv.Itoa(v.Duration * 1000),
				}
			}
			albums[k] = &jamsonic.Album{
				Artist: a.Name,
				ID:     album.ID,
				Name:   album.Name,
				Tracks: tracks,
				Year:   uint32(album.Year),
			}
		}
		result <- &jamsonic.Artist{
			Name:   a.Name,
			ID:     a.ID,
			Albums: albums,
		}
	}
}

func getAllArtists(c *Client) ([]*artist, error) {
	data, err := sendRequest(c.makeRequestURL("getArtists"))
	if err != nil {
		return nil, err
	}
	artists := make([]*artist, 0)
	for _, v := range data.ArtistList.Index {
		artists = append(artists, v.Artists...)
	}
	return artists, nil
}

func getArtistAlbums(c *Client, artistID string) ([]*album, error) {
	data, err := sendRequest(c.makeRequestURL("getArtist") + "&id=" + artistID)
	if err != nil {
		return nil, err
	}
	return data.Artist.Albums, nil
}

func getAlbumSongs(c *Client, albumID string) ([]*song, error) {
	data, err := sendRequest(c.makeRequestURL("getAlbum") + "&id=" + albumID)
	if err != nil {
		return nil, err
	}
	return data.Album.Songs, nil
}
