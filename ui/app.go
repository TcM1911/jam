// Copyright (c) 2018 Joakim Kennedy
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
	// "encoding/json"
	// "fmt"
	"log"
	"math/rand"
	"sort"
	// "strconv"
	"strings"

	// "time"

	"github.com/boltdb/bolt"
	"github.com/gdamore/tcell"
	// runewidth "github.com/mattn/go-runewidth"

	"github.com/TcM1911/jamsonic"
	"github.com/TcM1911/jamsonic/lastfm"
	"github.com/TcM1911/jamsonic/music"
)

const (
	play = iota
	stop
	pause
	next
	prev
)

// type Database struct {
// 	DB         *bolt.DB
// 	ArtistsMap map[string]bool
// 	Artists    sort.StringSlice
// 	Songs      map[string][]string
// 	Albums     map[string][]string
// 	LastAlbum  string
// }

type Status struct {
	ScrOffset map[bool]int
	Offset    int
	CurPos    map[bool]int
	CurView   int
	NumAlbum  map[bool]int
	InTracks  bool
	InSearch  bool
	LastFM    bool
	NumTrack  int
	Queue     [][]*music.BTrack // playlist, updated on each movement of cursor in artists view
	Query     []rune            // search query

	State       chan int // player's state: play, pause, stop, etc
	RepeatTrack bool
}

// App define the UI application
type App struct {
	Screen tcell.Screen
	Width  int
	Height int

	Provider jamsonic.Provider
	LastFM   *lastfm.Client

	// Better:
	// Database *Database
	DB         *bolt.DB
	ArtistsMap map[string]bool
	Artists    sort.StringSlice
	Playlists  sort.StringSlice
	Songs      map[string][]string
	Albums     map[string][]string

	LastAlbum string
	Status    *Status
}

// New creates a new UI
func New(provider jamsonic.Provider, lmclient *lastfm.Client, lastfm string, db *bolt.DB) (*App, error) {
	var lastfmStatus bool
	screen, err := tcell.NewScreen()
	if err != nil {
		return nil, err
	}
	err = screen.Init()
	if err != nil {
		return nil, err
	}
	width, height := screen.Size()
	if lastfm != "None" {
		lastfmStatus = true
	} else {
		lmclient = nil
	}
	return &App{
		Screen:     screen,
		Width:      width,
		Height:     height,
		Provider:   provider,
		LastFM:     lmclient,
		DB:         db,
		ArtistsMap: map[string]bool{},
		Artists:    sort.StringSlice{},
		Songs:      map[string][]string{},
		Albums:     map[string][]string{},
		Status: &Status{
			ScrOffset: map[bool]int{
				false: 0, // in artists view
				true:  0, // in tracks view
			},
			Offset: 0,
			CurPos: map[bool]int{
				false: 1, // same as in scrOffset. -1 is because the artist is unfolded (yet)
				true:  2,
			},
			CurView: 0,
			NumAlbum: map[bool]int{
				false: -1, // same as in scrOffset. -1 is because the artist is unfolded (yet)
				true:  0,
			},
			InTracks: false,
			InSearch: false,
			NumTrack: 0,
			Queue:    make([][]*music.BTrack, 0),
			State:    make(chan int),
			LastFM:   lastfmStatus,
		},
	}, nil
}

func (app *App) Run() {
	defer app.Screen.Fini()
	app.populateArtists()
	app.populatePlaylists()
	// log.Printf("Artists done")
	go app.player()
	app.mainLoop()
}

func (app *App) populatePlaylists() {
	app.Playlists = sort.StringSlice{}
	app.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Playlists"))
		c := b.Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			app.Playlists = append(app.Playlists, string(k))
		}

		return nil
	})
}

func (app *App) populateArtists() {
	app.Artists = sort.StringSlice{}
	app.DB.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("Library"))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			if !app.ArtistsMap[string(k)] {
				app.ArtistsMap[string(k)] = false
			}
			if v == nil {
				if err := b.Bucket(k).ForEach(func(kk []byte, vv []byte) error {
					app.Albums[string(k)] = append(app.Albums[string(k)], string(kk))

					return nil
				}); err != nil {
					log.Fatalf("Can't populate artists: %s", err)
				}
			}

		}
		for k := range app.ArtistsMap {
			app.Artists = append(app.Artists, k)
		}
		app.Artists.Sort()
		// log.Printf("Artists: %s", app.Artists)
		return nil
	})
}

func (app *App) populateSongs(what []string) {
	app.Songs = map[string][]string{}
	if err := app.DB.View(func(tx *bolt.Tx) error {
		var b *bolt.Bucket
		var c *bolt.Cursor
		if app.Status.CurView == 0 {
			i := app.Status.CurPos[false] - 1 + app.Status.ScrOffset[false]
			b = tx.Bucket([]byte("Library")).Bucket([]byte(what[i-app.numAlb(i)]))
		} else if app.Status.CurView == 1 {
			b = tx.Bucket([]byte("Playlists"))
		}
		c = b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			if v == nil {
				cc := b.Bucket(k).Cursor()
				for kk, vv := cc.First(); kk != nil; kk, vv = cc.Next() {
					app.Songs[string(k)] = append(app.Songs[string(k)], string(vv))
				}
			}
		}
		return nil
	}); err != nil {
		log.Fatalf("Can't populate songs: %s", err)
	}

}

func (app *App) search(what []string) {
	app.Status.InTracks = false
	app.Status.InSearch = true
	app.Status.NumTrack = 0
	app.Status.CurPos[true] = 2
	app.Status.ScrOffset[true] = 0
	app.Status.Query = []rune{}
	for {
		app.printStatus()
		app.Screen.Show()
		ev := app.Screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyRune:
				app.Status.Query = append(app.Status.Query, ev.Rune())
			case tcell.KeyBackspace, tcell.KeyBackspace2:
				if len(app.Status.Query) > 0 {
					app.Screen.SetContent(len(app.Status.Query), app.Height-1, ' ', nil, dfStyle)
					app.Status.Query = app.Status.Query[:len(app.Status.Query)-1]
				} else {
					app.Status.InSearch = false
					return
				}
			case tcell.KeyEnter:
				app.Status.InSearch = false
				return
			}
		}
		app.searchQuery(what)
	}
}

func (app *App) searchQuery(what []string) {
	app.Status.NumAlbum[false] = -1
	var i int
	if !app.Status.InSearch {
		i = app.Status.ScrOffset[false] + app.Status.CurPos[false]
	}
	if len(app.Status.Query) > 0 {
		for i < len(app.Artists) {
			if strings.HasPrefix(strings.ToLower(what[i]), strings.ToLower(string(app.Status.Query))) {
				if i > 2 {
					app.Status.ScrOffset[false] = i - 2
					app.Status.CurPos[false] = 3
				} else {
					app.Status.ScrOffset[false] = 0
					app.Status.CurPos[false] = i + 1
				}
				app.updateUI(what)
				return
			}
			i++
		}
	}

}

func (app *App) randomizeArtists() {
	app.Status.NumTrack = 0
	app.Status.NumAlbum[true] = 0
	app.Status.ScrOffset[true] = 0
	app.Status.CurPos[true] = 2
	var numAlbums int
	for i, art := range app.Artists {

		if art == "" {
			numAlbums++
			app.ArtistsMap[app.Artists[i-app.numAlb(i)]] = false

		}
	}

	var temp = make(sort.StringSlice, len(app.Artists)-numAlbums)

	perm := rand.Perm(len(app.Artists))
	var index int
	for _, v := range perm {
		if app.Artists[v] == "" {

			continue
		}
		temp[index] = app.Artists[v]
		index++

	}

	app.Artists = temp
	app.updateUI(app.Artists)

}

func (app *App) mainLoop() {
	var what sort.StringSlice
	for {
		app.Screen.Show()
		ev := app.Screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventResize:
			app.Width, app.Height = app.Screen.Size()
			app.updateUI(what)
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyEscape:
				return
			case tcell.KeyPgDn:
				app.pageDn()
			case tcell.KeyPgUp:
				app.pageUp()
			case tcell.KeyEnd:
				app.scrollDown()
			case tcell.KeyHome:
				app.scrollUp()
			case tcell.KeyTab:
				app.toggleView()
			case tcell.KeyUp:
				app.upEntry(what)
			case tcell.KeyDown:
				app.downEntry(what)
			case tcell.KeyEnter:
				app.Status.State <- play
			case tcell.KeyCtrlSpace:
				app.toggleTab()
			}
			switch ev.Rune() {
			case '/':
				app.search(what)
			case 'q':
				return
			case 'j':
				app.downEntry(what)
			case 'k':
				app.upEntry(what)
			case ' ':
				app.toggleAlbums()
			case 'u':
				err := music.RefreshLibrary(app.DB, app.Provider)
				if err != nil {
					log.Fatalf("Can't refresh library: %s", err)
				}
				app.populateArtists()
			case 'x':
				app.Status.State <- play
			case 'v':
				app.Status.State <- stop
			case 'c':
				app.Status.State <- pause
			case 'b':
				app.Status.State <- next
			case 'z':
				app.Status.State <- prev
			case 'n':
				app.searchQuery(what)
			case 'R':
				app.randomizeArtists()
			case 'r':
				app.Status.RepeatTrack = !app.Status.RepeatTrack
			case 'G':
				app.scrollDown()
			case 'g':
				app.scrollUp()
			}
		}
		if app.Status.CurView == 0 {
			what = app.Artists
		} else if app.Status.CurView == 1 {
			what = app.Playlists
		}
		app.updateUI(what)
	}
}
