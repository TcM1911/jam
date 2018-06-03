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
