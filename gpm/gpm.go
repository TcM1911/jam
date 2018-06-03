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
