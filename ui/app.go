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
	"encoding/json"
	"log"
	"math/rand"
	"sort"
	"strings"

	"github.com/TcM1911/jamsonic/native"

	"github.com/TcM1911/jamsonic"
	"github.com/TcM1911/jamsonic/lastfm"
	"github.com/TcM1911/jamsonic/storage"
	"github.com/gdamore/tcell"
)

const (
	play = iota
	stop
	pause
	next
	prev
)

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
	Queue     [][]*jamsonic.Track // playlist, updated on each movement of cursor in artists view
	Query     []rune              // search query

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

	DB         *storage.BoltDB
	Library    map[string]*jamsonic.Artist
	ArtistsMap map[string]bool
	Artists    sort.StringSlice
	Playlists  sort.StringSlice
	Songs      map[string][]string
	Albums     map[string][]string

	LastAlbum string
	Status    *Status
}

// New creates a new UI
func New(provider jamsonic.Provider, lmclient *lastfm.Client, lastfm string, db *storage.BoltDB) (*App, error) {
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
		Library:    map[string]*jamsonic.Artist{},
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
			Queue:    make([][]*jamsonic.Track, 0),
			State:    make(chan int),
			LastFM:   lastfmStatus,
		},
	}, nil
}

var musicPlayer *jamsonic.Player

func (app *App) Run() {
	defer app.Screen.Fini()
	app.populateArtists()
	app.populatePlaylists()
	// log.Printf("Artists done")
	if jamsonic.Experimental {
		streamHandler := native.New(jamsonic.DefaultLogger())
		callback := func(data *jamsonic.CallbackData) {
			app.printBar(data.Duration, data.CurrentTrack)
		}
		musicPlayer = jamsonic.NewPlayer(jamsonic.DefaultLogger(), app.Provider, streamHandler, callback, 1)
		go func() {
			err := musicPlayer.Error
			for {
				select {
				case <-err:
					// if jamsonic.Debug {
					// 	log.Println("Player error:", e.Error())
					// }
				}
			}
		}()

	}
	go app.player()
	app.mainLoop()
}

func (app *App) populatePlaylists() {
	if app.Provider.GetProvider() == jamsonic.GooglePlayMusic {
		app.Playlists = storage.GetPlaylists(app.DB)
	}
}

func (app *App) populateArtists() {
	if app.Provider.GetProvider() == jamsonic.GooglePlayMusic {
		app.Artists, app.ArtistsMap, app.Albums = storage.GetArtistsAndAlbums(app.DB)
		return
	}
	as, err := app.DB.Artists()
	if err != nil {
		app.Screen.Fini()
		log.Fatalln("Error when populating artists:", err.Error())
	}
	artists := sort.StringSlice{}
	artistsMap := make(map[string]bool)
	albums := make(map[string][]string)
	for _, artist := range as {
		tmpAlbums := make([]string, len(artist.Albums))
		for i, v := range artist.Albums {
			tmpAlbums[i] = v.Name
		}
		albums[artist.Name] = tmpAlbums
		artists = append(artists, artist.Name)
		artistsMap[artist.Name] = false
		app.Library[artist.Name] = artist
	}
	artists.Sort()
	app.Artists = artists
	app.ArtistsMap = artistsMap
	app.Albums = albums
}

func (app *App) populateSongs(what []string) {
	if app.Provider.GetProvider() == jamsonic.GooglePlayMusic {
		app.Songs = storage.GetTracks(app.DB, what, app.Status.CurView, app.Status.CurPos, app.Status.ScrOffset, app.numAlb)
		return
	}
	app.Songs = make(map[string][]string)
	// Artist index
	i := app.Status.CurPos[false] - 1 + app.Status.ScrOffset[false]
	// Album index
	j := app.numAlb(i) - 1
	tmp := what[i]
	for tmp == "" {
		i--
		tmp = what[i]
	}
	artist := app.Library[what[i]]
	if j < 0 {
		for _, album := range artist.Albums {
			trackList := make([]string, 0)
			for _, track := range album.Tracks {
				buf, _ := json.Marshal(track)
				trackList = append(trackList, string(buf))
			}
			app.Songs[album.Name] = trackList
		}
	} else {
		trackList := make([]string, 0)
		for _, track := range artist.Albums[j].Tracks {
			buf, _ := json.Marshal(track)
			trackList = append(trackList, string(buf))
		}
		app.Songs[artist.Albums[j].Name] = trackList
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
				if jamsonic.Experimental {
					// Create playlist
					var queue []*jamsonic.Track
					for i := app.Status.NumAlbum[true]; i < len(app.Status.Queue); i++ {
						album := app.Status.Queue[i]
						for j := app.Status.NumTrack; j < len(album); j++ {
							queue = append(queue, album[j])
						}
					}
					musicPlayer.CreatePlayQueue(queue)
					musicPlayer.Play()
				} else {
					app.Status.State <- play
				}
			case tcell.KeyCtrlSpace:
				app.toggleTab()
			}
			switch ev.Rune() {
			case '/':
				app.search(what)
			case 'q':
				if jamsonic.Experimental {
					musicPlayer.Stop()
				}
				return
			case 'j':
				app.downEntry(what)
			case 'k':
				app.upEntry(what)
			case ' ':
				app.toggleAlbums()
			case 'u':
				err := jamsonic.RefreshLibrary(app.DB, app.Provider)
				if err != nil {
					log.Fatalf("Can't refresh library: %s", err)
				}
				app.populateArtists()
			case 'x':
				if jamsonic.Experimental {
					// Create playlist
					var queue []*jamsonic.Track
					for i := app.Status.NumAlbum[true]; i < len(app.Status.Queue); i++ {
						album := app.Status.Queue[i]
						for j := app.Status.NumTrack; j < len(album); j++ {
							queue = append(queue, album[j])
						}
					}
					musicPlayer.CreatePlayQueue(queue)
					musicPlayer.Play()
				} else {
					app.Status.State <- play
				}
			case 'v':
				if jamsonic.Experimental {
					musicPlayer.Stop()
				} else {
					app.Status.State <- stop
				}
			case 'c':
				if jamsonic.Experimental {
					musicPlayer.Pause()
				} else {
					app.Status.State <- pause
				}
			case 'b':
				if jamsonic.Experimental {
					musicPlayer.Next()
				} else {
					app.Status.State <- next
				}
			case 'z':
				if jamsonic.Experimental {
					musicPlayer.Previous()
				} else {
					app.Status.State <- prev
				}
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
