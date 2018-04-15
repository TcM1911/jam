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
	"time"

	"github.com/TcM1911/jamsonic"
	"github.com/TcM1911/jamsonic/native"
	"github.com/TcM1911/jamsonic/storage"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

type TUI struct {
	currentPage int
	app         *tview.Application
	db          *storage.BoltDB

	player        *jamsonic.Player
	trackDuration time.Duration
	currentTrack  *jamsonic.Track

	header      *tview.TextView
	footer      *tview.TextView
	pages       *tview.Pages
	libraryView *tview.Flex
	artistView  *tview.List
	tracksView  *tview.List

	artists   sort.StringSlice
	artistMap map[string]*jamsonic.Artist

	albumListed map[string]*jamsonic.Album
	trackListed map[string]*jamsonic.Track
}

func New(db *storage.BoltDB, client jamsonic.Provider) {
	tui := &TUI{
		app:   tview.NewApplication(),
		db:    db,
		pages: tview.NewPages(),
	}

	pageLists := []string{"Library", "Settings"}

	header := tview.NewTextView().SetRegions(true).SetWrap(false).SetDynamicColors(true)
	header.SetBorder(true).SetTitle("Jamsonic")
	header.Highlight("0")

	tui.header = header

	for i, page := range pageLists {
		fmt.Fprintf(header, `%d ["%d"][white]%s[white][""]  `, i+1, i, page)
	}

	tui.footer = tview.NewTextView().SetRegions(true).SetWrap(false).SetDynamicColors(true)
	tui.footer.SetBorder(true)
	tui.drawFooter()

	// Layout
	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(tui.header, 3, 1, false).
		AddItem(tui.pages, 0, 1, true).
		AddItem(tui.footer, 3, 1, false)

	// Add pages
	tui.pages.AddPage("0", tui.createLibraryPage(), true, true)
	tui.pages.AddPage("1", tui.createSettingsPage(), true, false)

	// Register global key event handler.
	tui.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		// Music control events.
		case 'b':
			tui.player.Next()
			return nil
		case 'v':
			tui.player.Stop()
			return nil
		case 'c':
			tui.player.Pause()
			return nil
		case 'x':
			if tui.player.GetCurrentState() == jamsonic.Paused {
				tui.player.Play()
				return nil
			}
		case 'z':
			tui.player.Previous()
			return nil
		}
		switch event.Key() {
		// Pages
		case tcell.KeyCtrlN:
			tui.currentPage = (tui.currentPage + 1) % 2
			switchPage(tui, tui.currentPage)
		}
		return event
	})

	/// To be moved
	streamHandler := native.New()
	callback := func(data *jamsonic.CallbackData) {
		tui.trackDuration = data.Duration
		tui.currentTrack = data.CurrentTrack
		tui.drawFooter()
	}
	tui.player = jamsonic.NewPlayer(client, streamHandler, callback, 500)
	go func() {
		err := tui.player.Error
		for {
			select {
			case e := <-err:
				if jamsonic.Debug {
					log.Println("Player error:", e.Error())
				}
			}
		}
	}()

	tui.app.SetRoot(flex, true).Run()
}

func (tui *TUI) drawFooter() {
	min := int(tui.trackDuration.Minutes())
	secs := int(tui.trackDuration.Seconds()) % 60
	var title string
	if tui.currentTrack != nil {
		title = tui.currentTrack.Title
	}
	tui.footer.Clear()
	fmt.Fprintf(tui.footer, "%02d:%02d / %s", min, secs, title)
	tui.app.Draw()
}

func switchPage(tui *TUI, page int) {
	index := strconv.Itoa(page)
	tui.pages.SwitchToPage(index)
	tui.header.Highlight(index)
	if page == 0 {
		tui.app.SetFocus(tui.libraryView)
	}
}
