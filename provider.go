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

package jamsonic

import "io"

// MusicProvider is the provider identifier.
type MusicProvider int

const (
	// GooglePlayMusic is used as an identifier for the GPM provider.
	GooglePlayMusic MusicProvider = iota
	// SubSonic us used as an identifier for the subsonic provider.
	SubSonic
)

// Provider is a music provider.
type Provider interface {
	// ListTracks returns all the tracks from the provider. [DEPRECATED]
	ListTracks() ([]*Track, error)
	// FetchLibrary gets the library from the server. This implementation
	// should be used instead of the old implementations.
	FetchLibrary() ([]*Artist, error)
	// GetTrackInfo returns information abot the track from the provider.
	GetTrackInfo(trackID string) (*Track, error)
	// GetStream returns a ReadCloser stream of the track. The stream
	// has to be a MP3 encoded stream. [DEPRECATED]
	GetStream(songID string) (io.ReadCloser, error)
	// ListPlaylists returns all the playlists from the provider.
	ListPlaylists() ([]*Playlist, error)
	// ListPlaylistEntries returns to entries in the playlist. [DEPRECATED]
	ListPlaylistEntries() ([]*PlaylistEntry, error)
	// GetProvider returns the MusicProvider type.
	GetProvider() MusicProvider
}

// Track is a common structure from a music track.
type Track struct {
	// Artist is the name of the artist.
	Artist string
	// Album is the album.
	Album string
	// AlbumArtist is the album artist.
	AlbumArtist string
	// DiscNumber is the disc number for the track.
	DiscNumber uint8
	// TrackNumber is the track number for the track.
	TrackNumber uint32
	// DurationMillis is the song length in milliseconds.
	DurationMillis string
	// EstimatedSize is the estimated size for the track.
	EstimatedSize string
	// ID is the tracks unique ID.
	ID string
	// PlayCount is how many times to track has been played.
	PlayCount uint32
	// Title is the song title.
	Title string
	// Year is the year the track was released.
	Year uint32
}

// PlaylistEntry represents an entry in a playlist.
type PlaylistEntry struct {
	// PlaylistId is the ID of the playlist.
	PlaylistId string
	// TrackId is the ID of the track.
	TrackId string
}

// Playlist is a playlist structure.
type Playlist struct {
	// ID is the playlist's ID.
	ID string
	// Name is the name of the playlist.
	Name string
}

// Album holds all data for an album.
type Album struct {
	// Name is the album name
	Name string
	// Artist is the name of the artist for this album.
	Artist string
	// Year is the year the album was released.
	Year uint32
	// ID is the ID for the album.
	ID string
	// Tracks is an array of all the tracks.
	Tracks []*Track
}

// Artist holds all the data for an artist.
type Artist struct {
	// Name of the artist.
	Name string
	// ID for the artist.
	ID string
	// Albums is an array of all the artist's albums.
	Albums []*Album
}
