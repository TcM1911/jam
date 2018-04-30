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
	// Get the cached library from the database.
	as, err := tui.db.Artists()
	if err != nil {
		log.Fatalln(err)
	}

	// Sort the list.
	tui.artists = sort.StringSlice{}
	tui.artistMap = make(map[string]*jamsonic.Artist)
	for _, artist := range as {
		tui.artists = append(tui.artists, artist.Name)
		tui.artistMap[artist.Name] = artist
	}
	tui.artists.Sort()

	// Clear the list and poulate the TUI list.
	tui.artistView.Clear()
	for _, a := range tui.artists {
		tui.artistView.AddItem(a, "", 0, nil)
	}
}

// Creates a TUI list with all the artists in the library.
func createArtistList(t *TUI) *tview.List {
	// Simple list.
	artistList := tview.NewList().ShowSecondaryText(false)
	artistList.SetBorder(true).SetTitle("Artists")

	// Called every time an entry in the list is changed to.
	// When a artist is highlighted, the artist's tracks are added to the track list.
	artistList.SetChangedFunc(func(index int, a string, _ string, _ rune) {
		t.populateTracks(a)
	})

	// Some key overrides.
	artistList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// When tab is pressed, switch focus to the trackView.
		if event.Key() == tcell.KeyTab {
			t.app.SetFocus(t.tracksView)
			return nil
		}
		// Handle when Enter is pressed on an artist
		if event.Key() == tcell.KeyEnter {
			t.playArtist(t.artists[t.artistView.GetCurrentItem()])
			return nil
		}
		// Also handle music control and VIM bindings.
		return t.vimBindings(t.musicControl(event))
	})

	return artistList
}

// Adds the tracks based on the filter to the tracksView.
func (t *TUI) populateTracks(artistStr string) {
	// First ensure we start with an empty tracksView.
	t.tracksView.Clear()
	// Used to track album listings in the tracksView.
	t.albumListed = make(map[string]*jamsonic.Album)
	// Used to track the songs listed in the tracksView.
	t.trackListed = make(map[string]*jamsonic.Track)
	// Get the line length so we can calculate how many runes are needed to fill the line.
	_, _, lineWidth, _ := t.tracksView.GetInnerRect()

	// Get the Artist struct for the selected artist.
	artist := t.artistMap[artistStr]
	for _, album := range artist.Albums {
		// Create an entry for the albums in the tracks list. Include the year.
		albumLine := fmt.Sprintf("~~~%s~~~", album.Name)
		if album.Year != uint32(0) {
			albumLine += "{" + strconv.Itoa(int(album.Year)) + "}"
		}
		// Fill the rest of the line with "~"
		width := runewidth.StringWidth(albumLine)
		if (lineWidth - width) > 0 {
			albumLine += getFillString(lineWidth, width, "~")
		}
		// Store a mapping to this album so we can know this line is for this album.
		t.albumListed[albumLine] = album
		// Add the line to the tracksView list.
		t.tracksView.AddItem(albumLine, "", 0, nil)

		// Add entries for all the tracks in the album too.
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

			// Add the track duration to the end of the line if we have it.
			d, err := strconv.Atoi(tr.DurationMillis)
			// If we parsed it we add duration to the entry line
			if err == nil {
				// Generate a duration string "mm:ss".
				durration := durationString(time.Millisecond * time.Duration(d))
				// Calculate how much padding is needed between the track name
				// and the track durration.
				width := runewidth.StringWidth(entry + durration)
				// Create the final line.
				entry += getFillString(lineWidth, width, " ") + durration
			}
			// Add the line to the tracksView
			t.tracksView.AddItem(entry, "", 0, nil)
			// Save entries for this line to corresponding album and track struct.
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
	})

	tracks.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Switch to the artistView if Tab is pressed.
		if event.Key() == tcell.KeyTab {
			tui.app.SetFocus(tui.artistView)
			return nil
		}
		// Handle music control input and VIM bindings.
		return tui.vimBindings(tui.musicControl(event))
	})
	return tracks
}

func (tui *TUI) createLibraryPage() *tview.Flex {
	tui.artistView = createArtistList(tui)
	tui.tracksView = createTrackList(tui)

	// Page split in two for the artists and tracks view. 2/3 of the space
	// is given to the tracks and 1/3 is given the the artists.
	tui.libraryView = tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(tui.artistView, 0, 1, true).
		AddItem(tui.tracksView, 0, 2, false)

	// Populate the artists list and select the first item.
	tui.populateArtists()
	tui.artistView.SetCurrentItem(0)
	return tui.libraryView
}

// playTracks figures out which tracks to play from the tracksView line selected.
func (tui *TUI) playTracks(index int, entry string) {
	var tracks []*jamsonic.Track
	// Check if it's an album entry. If it is get the album from albumListed map.
	if entry[0] == '~' {
		tracks = tui.albumListed[entry].Tracks
	} else {
		// Otherwise get the track and album structs from the maps
		tr := tui.trackListed[entry]
		alb := tui.albumListed[entry]
		// Select the track and the tracks after in the album.
		tracks = alb.Tracks[int(tr.TrackNumber)-1:]
	}
	tui.player.CreatePlayQueue(tracks)
	tui.player.Play()
}

// playArtist plays all songs for an artist.
func (tui *TUI) playArtist(entry string) {
	var tracks []*jamsonic.Track
	for _, album := range tui.artistMap[entry].Albums {
		tracks = append(tracks, album.Tracks...)
	}
	tui.player.CreatePlayQueue(tracks)
	tui.player.Play()
}
