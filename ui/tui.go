// Copyright (c) 2016, 2017 Evgeny Badin

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package ui

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"time"

	runewidth "github.com/mattn/go-runewidth"

	"github.com/budkin/jam/music"
)

var (
	defDur   = time.Duration(0) // this and below are used for when nothing is playing
	defTrack = &music.BTrack{}
)

func (app *App) updateUI(what []string) {
	app.Screen.Clear()
	fill(app.Screen, app.Width/3, 1, 1, app.Height-3, '│', dfStyle)
	app.printHeader()
	if len(what) > 0 {
		app.printArtists(app.Status.ScrOffset[false], app.Height+app.Status.ScrOffset[false]-3, what)
		app.printSongs(app.Status.ScrOffset[true], app.Height+app.Status.ScrOffset[true]-3, what)
		for app.Status.CurPos[false] > app.Height-3 {
			app.Status.CurPos[false]--
			app.Status.ScrOffset[false]++
		}
		for app.Status.CurPos[true] > app.Height-3 {
			app.Status.CurPos[true] = 2
			app.Status.NumTrack = 0
			app.Status.NumAlbum[true] = 0
		}
		app.hlEntry(what)
	}
	app.printStatus()
	app.printBar(defDur, defTrack)
}

func (app *App) hlEntry(what []string) {
	var song string
	i := app.Status.CurPos[false] - 1 + app.Status.ScrOffset[false]
	if what[i] != "" {
		printSingleItem(app.Screen, 0, app.Status.CurPos[false], hlStyle, what[i], 1, true, app.Width)
	} else {
		printSingleItem(app.Screen, 0, app.Status.CurPos[false], hlStyle, app.Albums[what[i-app.numAlb(i)]][app.numAlb(i)-1], 3, true, app.Width)

	}

	if app.Status.InTracks {
		js := new(music.BTrack)
		if app.Status.CurView == 0 {
			song = app.Songs[app.Albums[what[i-app.numAlb(i)]][app.Status.NumAlbum[true]]][app.Status.NumTrack]
		} else if app.Status.CurView == 1 {
			song = app.Songs[what[i]][app.Status.NumTrack]
		}
		err := json.Unmarshal([]byte(song), js)
		if err != nil {
			panic(err)
		}
		printSingleItem(app.Screen, app.Width/3+2, app.Status.CurPos[app.Status.InTracks], hlStyle, app.makeSongLine(js), 0, false, app.Width)
	}
}

func (app *App) printHeader() {
	fill(app.Screen, 0, 0, app.Width, 1, ' ', hlStyle)
	if app.Status.CurView == 0 {
		print(app.Screen, 1, 0, hlStyle, "Artist / Album")
	} else if app.Status.CurView == 1 {
		print(app.Screen, 1, 0, hlStyle, "Playlist")
	}
	print(app.Screen, app.Width/3+2, 0, hlStyle, "Track")
	print(app.Screen, app.Width-8, 0, hlStyle, "Library")
	fill(app.Screen, 0, app.Height-2, app.Width, 1, ' ', hlStyle)
}

func (app *App) printStatus() {
	if app.Status.InSearch {
		app.Screen.SetContent(0, app.Height-1, '/', nil, dfStyle)
		for i, r := range app.Status.Query {
			app.Screen.SetContent(i+1, app.Height-1, r, nil, dfStyle)
		}
	}
}

func (app *App) printBar(dur time.Duration, track *music.BTrack) {
	strdur := ""
	str := fmt.Sprintf(" %02v:%02v %v - %v ", int(dur.Minutes()), int(dur.Seconds())%60,
		track.Artist, track.Title)
	lenstr := 0
	for _, r := range str {
		lenstr += runewidth.RuneWidth(r)
	}
	print(app.Screen, 0, app.Height-2, barStyle, str)
	leng := app.Width - lenstr
	durat, _ := strconv.Atoi(track.DurationMillis)
	dura := time.Duration(durat) * time.Millisecond
	for i := 0.0; i < float64(leng)/dura.Seconds()*dur.Seconds(); i += 1.0 {
		strdur += "—"
	}
	for len(strdur) < leng {
		strdur += " "
	}
	print(app.Screen, lenstr, app.Height-2, barStyle, strdur)
	app.Screen.Show()
}

func (app *App) printArtists(beg, end int, what []string) {
	if len(what) < end {
		end = len(what)
	}
	j := 1
	for beg < end {
		if what[beg] != "" {
			printSingleItem(app.Screen, 0, j, dfStyle, what[beg], 1, true, app.Width)

		} else {
			printSingleItem(app.Screen, 0, j, dfStyle, app.Albums[what[beg-app.numAlb(beg)]][app.numAlb(beg)-1], 3, true, app.Width)
		}
		j++
		beg++

	}
}

func (app *App) printAlbum(y int, alb string) {
	makeAlbumLine(&alb, app.Width)
	print(app.Screen, app.Width/3+2, y, alStyle, alb)

}

func (app *App) printSongs(beg, end int, what []string) {
	app.Status.Queue = [][]*music.BTrack{}
	var js *music.BTrack
	app.populateSongs(what)
	i, k := 0, 1
	if app.Status.CurView == 1 {
		que := []*music.BTrack{}
		j := app.Status.CurPos[false] - 1 + app.Status.ScrOffset[false]
		if len(app.Songs[what[j]]) < end {
			end = len(app.Songs[what[j]])
		}
		for _, song := range app.Songs[what[j]] {
			js = new(music.BTrack)
			json.Unmarshal([]byte(song), js)
			if i >= beg && i < end {
				printSingleItem(app.Screen, app.Width/3+2, k, dfStyle, app.makeSongLine(js), 0, false, app.Width)
				k++
			}
			que = append(que, js)
			i++
		}
		for l := app.numAlb(j); l > 1; l-- {
			app.Status.Queue = append(app.Status.Queue, []*music.BTrack{})
		}
		app.Status.Queue = append(app.Status.Queue, que)

	} else if app.Status.NumAlbum[false] == -1 {
		j := app.numSongs()
		if j < end {
			end = j
		}
		keys := []string{}
		for a := range app.Songs {
			keys = append(keys, a)
		}
		sort.Strings(keys)
		app.LastAlbum = keys[len(keys)-1]
		for _, key := range keys {
			que := []*music.BTrack{}
			if i >= beg && i < end {
				app.printAlbum(k, key)
				k++
			}
			i++
			for _, song := range app.Songs[key] {
				js = new(music.BTrack)
				json.Unmarshal([]byte(song), js)
				if i >= beg && i < end {
					printSingleItem(app.Screen, app.Width/3+2, k, dfStyle, app.makeSongLine(js), 0, false, app.Width)
					que = append(que, js)
					k++
				}
				i++
			}
			app.Status.Queue = append(app.Status.Queue, que)

		}
	} else {
		que := []*music.BTrack{}
		j := app.Status.CurPos[false] - 1 + app.Status.ScrOffset[false]
		if len(app.Songs[app.Albums[app.Artists[j-app.numAlb(j)]][app.Status.NumAlbum[false]]]) < end {
			end = len(app.Songs[app.Albums[app.Artists[j-app.numAlb(j)]][app.Status.NumAlbum[false]]])
		}

		for _, song := range app.Songs[app.Albums[app.Artists[j-app.numAlb(j)]][app.Status.NumAlbum[false]]] {
			js = new(music.BTrack)
			json.Unmarshal([]byte(song), js)
			if i >= beg && i < end {
				printSingleItem(app.Screen, app.Width/3+2, k, dfStyle, app.makeSongLine(js), 0, false, app.Width)
				k++
			}
			que = append(que, js)
			i++
		}
		for l := app.numAlb(j); l > 1; l-- {
			app.Status.Queue = append(app.Status.Queue, []*music.BTrack{})
		}
		app.Status.Queue = append(app.Status.Queue, que)
	}
}
