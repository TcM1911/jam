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
	// AddPlaylists stores the playlists to teh database. This methods is
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
