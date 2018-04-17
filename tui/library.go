// Copyright (c) 2018 Joakim Kennedy

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

package tui

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/TcM1911/jamsonic"
	"github.com/gdamore/tcell"
	"github.com/mattn/go-runewidth"
	"github.com/rivo/tview"
)

const (
	libraryTitle = "Library"
)

func (tui *TUI) populateArtists() {
	as, err := tui.db.Artists()
	if err != nil {
		log.Fatalln(err)
	}
	tui.artists = sort.StringSlice{}
	tui.artistMap = make(map[string]*jamsonic.Artist)

	for _, artist := range as {
		tui.artists = append(tui.artists, artist.Name)
		tui.artistMap[artist.Name] = artist
	}
	tui.artists.Sort()

	tui.artistView.Clear()
	for _, a := range tui.artists {
		tui.artistView.AddItem(a, "", 0, nil)
	}
}

func createArtistList(t *TUI) *tview.List {
	artistList := tview.NewList().ShowSecondaryText(false)
	artistList.SetBorder(true).SetTitle("Artists")

	// Called every time an entry in the list is changed to.
	artistList.SetChangedFunc(func(index int, a string, _ string, _ rune) {
		t.populateTracks(a)
	})

	// Some key overrides.
	artistList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			t.app.SetFocus(t.tracksView)
			return nil
		}
		return t.musicControl(event)
	})

	return artistList
}

func (t *TUI) populateTracks(a string) {
	t.tracksView.Clear()
	t.albumListed = make(map[string]*jamsonic.Album)
	t.trackListed = make(map[string]*jamsonic.Track)
	_, _, lineWidth, _ := t.tracksView.GetInnerRect()
	artist := t.artistMap[a]
	for _, album := range artist.Albums {
		albumLine := fmt.Sprintf("~~~%s~~~", album.Name)

		if album.Year != uint32(0) {
			albumLine += "{" + strconv.Itoa(int(album.Year)) + "}"
		}
		// Fill the rest of the line with "~"
		width := runewidth.StringWidth(albumLine)
		if (lineWidth - width) > 0 {
			albumLine += getFillString(lineWidth, width, "~")
		}
		t.albumListed[albumLine] = album
		t.tracksView.AddItem(albumLine, "", 0, nil)
		for i, tr := range album.Tracks {
			// Ensure the track has a matching artist name.
			tr.Artist = artist.Name
			// Ensure the track has a matching album name.
			tr.Album = album.Name
			// Ensure the track has a track number.
			if tr.TrackNumber == uint32(0) {
				tr.TrackNumber = uint32(i + 1)
			}
			entry := fmt.Sprintf("%d. %s", tr.TrackNumber, tr.Title)
			d, err := strconv.Atoi(tr.DurationMillis)
			// If we parsed it we add duration to the entry line
			if err == nil {
				durration := durationString(time.Millisecond * time.Duration(d))
				width := runewidth.StringWidth(entry + durration)
				entry += getFillString(lineWidth, width, " ") + durration
			}
			// durationStr :=
			t.tracksView.AddItem(entry, "", 0, nil)
			t.albumListed[entry] = album
			t.trackListed[entry] = tr
		}
	}
}

func durationString(d time.Duration) string {
	min := int(d.Minutes())
	secs := int(d.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d", min, secs)
}

func getFillString(lineLength, stringLength int, fillStr string) string {
	diff := lineLength - stringLength
	if diff < 0 {
		return ""
	}
	count := diff / runewidth.StringWidth(fillStr)
	return strings.Repeat(fillStr, count)
}

func createTrackList(tui *TUI) *tview.List {
	tracks := tview.NewList().ShowSecondaryText(false)
	tracks.SetBorder(true).SetTitle("Tracks")

	tui.albumListed = make(map[string]*jamsonic.Album)
	tui.trackListed = make(map[string]*jamsonic.Track)

	// When track or album is selected via Enter.
	tracks.SetSelectedFunc(func(index int, a string, _ string, _ rune) {
		tui.playTracks(index, a)
	})

	tracks.SetChangedFunc(func(index int, line string, _ string, _ rune) {
		if line[0] == '~' {
			tui.lastAlbum = line
		}
	})

	tracks.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			tui.app.SetFocus(tui.artistView)
			return nil
		}
		return tui.musicControl(event)
	})
	return tracks
}

func (tui *TUI) createLibraryPage() *tview.Flex {

	tui.artistView = createArtistList(tui)
	tui.tracksView = createTrackList(tui)

	tui.libraryView = tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(tui.artistView, 0, 1, true).
		AddItem(tui.tracksView, 0, 2, false)

	tui.populateArtists()

	tui.artistView.SetCurrentItem(0)
	return tui.libraryView
}

func (tui *TUI) playTracks(index int, entry string) {
	if entry[0] == '~' {
		tracks := tui.albumListed[entry].Tracks
		tui.player.CreatePlayQueue(tracks)
		tui.player.Play()
	} else {
		tr := tui.trackListed[entry]
		artist := tui.artistMap[tr.Artist]
		for _, alb := range artist.Albums {
			if alb.Name != tr.Album {
				continue
			}
			tracks := alb.Tracks[int(tr.TrackNumber)-1:]
			tui.player.CreatePlayQueue(tracks)
			tui.player.Play()
			return
		}
	}
}
