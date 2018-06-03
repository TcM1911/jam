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

type apiData struct {
	Response apiResponse `json:"subsonic-response"`
}

type apiResponse struct {
	Status     string     `json:"status"`
	Version    string     `json:"version"`
	ArtistList artistList `json:"artists"`
	Artist     artist     `json:"artist"`
	Album      album      `json:"album"`
}

type artistList struct {
	Index []index `json:"index"`
}

type index struct {
	Name    string    `json:"name"`
	Artists []*artist `json:"artist"`
}

type artist struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	CoverArt   string   `json:"coverArt"`
	AlbumCount int      `json:"albumCount"`
	Albums     []*album `json:"album"`
}

type album struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Artist    string  `json:"artist"`
	ArtistID  string  `json:"artistId"`
	CoverArt  string  `json:"coverArt"`
	SongCount int     `json:"songCount"`
	Duration  int     `json:"duration"`
	Year      int     `json:"year"`
	Genre     string  `json:"genre"`
	Songs     []*song `json:"song"`
}

type song struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	Track      int    `json:"track"`
	Year       int    `json:"year"`
	Size       int    `json:"size"`
	Duration   int    `json:"duration"`
	DiscNumber int    `json:"discNumber"`
}
