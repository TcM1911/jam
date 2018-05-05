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

package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/TcM1911/jamsonic"
	"github.com/TcM1911/jamsonic/gpm"
	"github.com/TcM1911/jamsonic/lastfm"
	"github.com/TcM1911/jamsonic/storage"
	"github.com/TcM1911/jamsonic/subsonic"
	"github.com/TcM1911/jamsonic/ui"
)

const (
	// BANNER is what is printed for help/info output.
	BANNER = "Jamsonic - %s\n"
)

var (
	vers         bool
	debug        bool
	lastFM       bool
	useGPM       bool
	experimental bool
)

func init() {
	// parse flags
	flag.BoolVar(&vers, "version", false, "print version and exit")
	flag.BoolVar(&debug, "debug", false, "debug")
	flag.BoolVar(&lastFM, "lastfm", false, "Enable LastFM scrobbler")
	flag.BoolVar(&useGPM, "googlemusic", false, "Use Google Play Music")
	flag.BoolVar(&experimental, "experimental", false, "Use experimental features")

	flag.Usage = func() {
		fmt.Fprint(os.Stderr, fmt.Sprintf(BANNER, jamsonic.Version))
		flag.PrintDefaults()
	}

	flag.Parse()

	if experimental {
		jamsonic.Experimental = true
	}
	if debug {
		jamsonic.Debug = true
	}

	if vers {
		fmt.Printf("%s\n", jamsonic.Version)
		os.Exit(0)
	}
}

func main() {
	db, err := storage.Open()
	defer db.Bolt.Close()
	if err != nil {
		log.Fatalf("Can't open database: %s", err)
	}
	if useGPM {
		provider, lmclient, lastfm, err := gpm.CheckCreds(db, &lastFM)
		if err != nil {
			log.Fatalf("Can't connect to Google Music: %s", err)
		}
		if err = doUI(provider, lmclient, lastfm, db); err != nil {
			log.Fatalf("Can't start UI: %s", err)
		}
	} else {
		client, err := subsonic.New(db, jamsonic.DefaultCredentialRequest)
		if err != nil {
			log.Fatalln("Can't connect to SubSonic server:", err.Error())
		}
		db.LibName = []byte(client.Host())
		if err != nil {
			log.Fatalln("Failed to sync the library with the SubSonic server:", err.Error())
		}
		doUI(client, &lastfm.Client{}, "None", db)
	}
}

func doUI(provider jamsonic.Provider, lmclient *lastfm.Client, lastfm string, db *storage.BoltDB) error {
	app, err := ui.New(provider, lmclient, lastfm, db)
	if err != nil {
		return err
	}
	app.Run()
	return nil
}
