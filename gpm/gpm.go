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

package gpm

/*
import (
	"io"

	"github.com/TcM1911/jamsonic"
	"github.com/budkin/gmusic"
)

// Client is an API client for Google Play Music.
type Client struct {
	*gmusic.GMusic
}

// GetProvider returns Google Play Music as the provider.
func (c *Client) GetProvider() jamsonic.MusicProvider {
	return jamsonic.GooglePlayMusic
}

func (c *Client) FetchLibrary() ([]*jamsonic.Artist, error) {
	panic("this should not be called. Use old method.")
}

// GetStream returns a ReadCloser stream of the track. The stream
// has to be a MP3 encoded stream.
func (c *Client) GetStream(songID string) (io.ReadCloser, error) {
	r, err := c.GMusic.GetStream(songID)
	return r.Body, err
}

// GetTrackInfo returns information abot the track from the provider.
func (c *Client) GetTrackInfo(songID string) (*jamsonic.Track, error) {
	t, err := c.GMusic.GetTrackInfo(songID)
	if err != nil {
		return nil, err
	}
	return convertTrack(t), nil
}

// ListPlaylistEntries returns to entries in the playlist.
func (c *Client) ListPlaylistEntries() ([]*jamsonic.PlaylistEntry, error) {
	entries, err := c.GMusic.ListPlaylistEntries()
	if err != nil {
		return make([]*jamsonic.PlaylistEntry, 0), err
	}
	return convertPlaylistEntries(entries), nil
}

// ListPlaylists returns all the playlists from the provider.
func (c *Client) ListPlaylists() ([]*jamsonic.Playlist, error) {
	ps, err := c.GMusic.ListPlaylists()
	if err != nil {
		return make([]*jamsonic.Playlist, 0), err
	}
	return convertPlaylists(ps), nil
}

// ListTracks returns all the tracks from the provider.
func (c *Client) ListTracks() ([]*jamsonic.Track, error) {
	ts, err := c.GMusic.ListTracks()
	if err != nil {
		return make([]*jamsonic.Track, 0), err
	}
	return convertTracks(ts), nil
}

func convertTrack(t *gmusic.Track) *jamsonic.Track {
	return &jamsonic.Track{
		Artist:         t.Artist,
		Album:          t.Album,
		AlbumArtist:    t.AlbumArtist,
		DiscNumber:     t.DiscNumber,
		TrackNumber:    t.TrackNumber,
		DurationMillis: t.DurationMillis,
		EstimatedSize:  t.EstimatedSize,
		ID:             t.ID,
		PlayCount:      t.PlayCount,
		Title:          t.Title,
		Year:           t.Year,
	}
}

func convertTracks(ts []*gmusic.Track) []*jamsonic.Track {
	a := make([]*jamsonic.Track, len(ts), len(ts))
	for i := 0; i < len(ts); i++ {
		a[i] = convertTrack(ts[i])
	}
	return a
}

func convertPlaylistEntries(es []*gmusic.PlaylistEntry) []*jamsonic.PlaylistEntry {
	a := make([]*jamsonic.PlaylistEntry, len(es), len(es))
	for i := 0; i < len(es); i++ {
		a[i] = &jamsonic.PlaylistEntry{
			PlaylistId: es[i].PlaylistId,
			TrackId:    es[i].TrackId,
		}
	}
	return a
}

func convertPlaylists(ps []*gmusic.Playlist) []*jamsonic.Playlist {
	a := make([]*jamsonic.Playlist, len(ps), len(ps))
	for i := 0; i < len(ps); i++ {
		a[i] = &jamsonic.Playlist{
			ID:   ps[i].ID,
			Name: ps[i].Name,
		}
	}
	return a
}
*/
