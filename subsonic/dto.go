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
