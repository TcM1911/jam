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

package tui

import (
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/TcM1911/jamsonic"
	"github.com/TcM1911/jamsonic/native"
	"github.com/TcM1911/jamsonic/storage"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

// TUI is the struct for the terminal UI.
type TUI struct {
	logger *jamsonic.Logger
	// Holds the id of the current page being displayed by the TUI.
	currentPage int
	// The application struct.
	app *tview.Application
	// Interface to the storage.
	db *storage.BoltDB

	// Provider of music
	provider jamsonic.Provider

	// The music player controller.
	player *jamsonic.Player
	// Current duration of the track being played. This value is updated
	// by the callback function for the player.
	trackDuration time.Duration
	// Current track being played. The value is updated by the callback function.
	currentTrack *jamsonic.Track

	// The window object
	window *tview.Flex
	// The top section of the TUI. Displays current and available pages.
	header *tview.TextView
	// The bottom section of the TUI. Displays track duration and title.
	// The content is updated by the callback function.
	footer *tview.TextView
	// Middle section of the TUI.
	pages *tview.Pages
	// The Library page. This page is split up in two parts. The artistView and tracksView
	libraryView *tview.Flex
	// Displays a selectable list of all artists in the library.
	artistView *tview.List
	// Displays the tracks. List is dependent on which artist is selected in the artistView.
	tracksView *tview.List

	// artists is a sorted list of all artists. This list has the same order as what is displayed
	// in the TUI. This slice can be used to find the artist if only the index is available.
	artists sort.StringSlice
	// artistMap maps the strings in the artists list to an artist struct. This is used to retrive
	// the right struct when an item is selected in the TUI.
	artistMap map[string]*jamsonic.Artist

	// albumListed is used to track the album entries in the tracks list. Using the string line in
	// the TUI list, the correct Album struct can be gotten.
	albumListed map[string]*jamsonic.Album
	// trackListed is used to map the track listing to the correct Track struct in the TUI track list.
	trackListed map[string]*jamsonic.Track

	// settingsList is the menu list with all settings categories.
	settingsList *tview.List
}

// New returns a TUI object. This should only be called once.
func New(db *storage.BoltDB, client jamsonic.Provider, logger *jamsonic.Logger) *TUI {
	tui := &TUI{
		app:    tview.NewApplication(),
		db:     db,
		pages:  tview.NewPages(),
		logger: logger,
	}

	// Header
	header := tview.NewTextView().SetRegions(true).SetWrap(false).SetDynamicColors(true)
	banner := fmt.Sprintf("[ Jamsonic %s ]", jamsonic.Version)
	header.SetBorder(true).SetTitle(banner)
	header.Highlight("0")

	tui.header = header

	pageLists := []string{"Library", "Settings", "Log"}
	for i, page := range pageLists {
		fmt.Fprintf(header, `%d ["%d"][white]%s[white][""]  `, i+1, i, page)
	}

	// Footer
	tui.footer = tview.NewTextView().SetRegions(true).SetWrap(false).SetDynamicColors(true)
	tui.footer.SetBorder(true)
	tui.drawFooter()

	// Layout
	tui.window = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(tui.header, 3, 1, false).
		AddItem(tui.pages, 0, 1, true).
		AddItem(tui.footer, 3, 1, false)

	// Add pages
	logPage := tui.createLogPage()
	tui.pages.AddPage("0", tui.createLibraryPage(), true, true)
	tui.pages.AddPage("1", tui.createSettingsPage(), true, false)
	tui.pages.AddPage("2", logPage, true, false)

	// Set logger
	logger.SetOutput(logPage)
	logger.DebugLog("Switched to log page for logging.")

	// Register global key event handler.
	tui.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		return tui.globalControl(event)
	})

	/// To be moved
	handlerLogger := logger.SubLogger("[Stream handler]")
	streamHandler := native.New(handlerLogger)
	playerLogger := logger.SubLogger("[Player]")
	logger.DebugLog("Starting the player.")
	tui.player = jamsonic.NewPlayer(playerLogger, client, streamHandler, tui.playerCallback, 500)
	go func() {
		err := tui.player.Error
		for {
			select {
			case e := <-err:
				playerLogger.ErrorLog("Player error: " + e.Error())
			}
		}
	}()
	tui.provider = client

	// Hack to redraw the tracks list after the app has started.
	// Otherwise the line is not generated with right width.
	go func() {
		time.Sleep(50 * time.Millisecond)
		curr := tui.artistView.GetCurrentItem()
		tui.populateTracks(tui.artists[curr])
		tui.app.Draw()
	}()
	// End hack.

	return tui
}

// Run starts the TUI application.
func (tui *TUI) Run() error {
	return tui.app.SetRoot(tui.window, true).Run()
}

// drawFooter updates the footer with the latest information.
// This is called by the player's callback function.
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

func (tui *TUI) playerCallback(data *jamsonic.CallbackData) {
	tui.trackDuration = data.Duration
	tui.currentTrack = data.CurrentTrack
	tui.drawFooter()
}

func switchPage(tui *TUI, page int) {
	index := strconv.Itoa(page)
	tui.pages.SwitchToPage(index)
	tui.header.Highlight(index)
	if page == 0 {
		tui.app.SetFocus(tui.libraryView)
	}
}
