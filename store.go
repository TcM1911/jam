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

import "errors"

var (
	// ErrNoCredentialsStored is returned if the backend does not have any
	// credentials stored.
	ErrNoCredentialsStored = errors.New("No credentials stored")
)

// MusicStore is the interface for databases which stores library caches.
type MusicStore interface {
	// AddTracks stores the tracks to the database. This methods is
	// deprecated and should not be used by new implementations.
	AddTracks([]*Track) error
	// AddPlaylists stores the playlists to the database. This methods is
	// deprecated and should not be used by new implementations.
	AddPlaylists(Provider, []*Playlist, []*PlaylistEntry) error
	// Artists returns the stored artists from the database.
	Artists() ([]*Artist, error)
	// SaveArtists saves the artists to the database.
	SaveArtists(artists []*Artist) error
}

// AuthStore is the interface for databases which handles credentials.
type AuthStore interface {
	// GetCredentials gets the credentials from the database.
	GetCredentials(key []byte) ([]byte, error)
	// SaveCredentials saves the credentials to the database.
	SaveCredentials(key []byte, credStruct []byte) error
}
