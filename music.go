// Copyright (c) 2018 Joakim Kennedy
// Copyright (c) 2016, 2017 Evgeny Badin
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

func RefreshLibrary(db MusicStore, provider Provider) error {
	var err error
	if provider.GetProvider() == GooglePlayMusic {
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
		err = db.AddTracks(tracks)
		if err != nil {
			return err
		}
		err = db.AddPlaylists(provider, playlists, entries)
		if err != nil {
			return err
		}
	} else {
		albums, err := provider.FetchLibrary()
		if err != nil {
			return err
		}
		err = db.SaveArtists(albums)
	}
	return err
}
