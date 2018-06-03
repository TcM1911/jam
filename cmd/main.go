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
	"os"

	"github.com/TcM1911/jamsonic"
	"github.com/TcM1911/jamsonic/storage"
	"github.com/TcM1911/jamsonic/subsonic"
	"github.com/TcM1911/jamsonic/tui"
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
	legacy       bool
)

func init() {
	// parse flags
	flag.BoolVar(&vers, "version", false, "print version and exit")
	flag.BoolVar(&debug, "debug", false, "debug")

	flag.Usage = func() {
		fmt.Fprint(os.Stderr, fmt.Sprintf(BANNER, jamsonic.Version))
		flag.PrintDefaults()
	}

	flag.Parse()

	if vers {
		fmt.Printf("%s\n", jamsonic.Version)
		os.Exit(0)
	}
}

func main() {
	logger := jamsonic.DefaultLogger()
	if debug {
		logger.SetLevel(jamsonic.DebugLevel)
	}
	storageLogger := logger.SubLogger("[Storage]")
	db, err := storage.Open(storageLogger)
	defer db.Bolt.Close()
	if err != nil {
		logger.ErrorLog("Can't open database: " + err.Error())
		return
	}
	subsonicLogger := logger.SubLogger("[Subsonic client]")
	client, err := subsonic.New(db, jamsonic.DefaultCredentialRequest, subsonicLogger)
	if err != nil {
		logger.ErrorLog("Can't connect to SubSonic server: " + err.Error())
		return
	}
	db.LibName = []byte(client.Host())
	if err != nil {
		logger.ErrorLog("Failed to sync the library with the SubSonic server: " + err.Error())
		return
	}
	ui := tui.New(db, client, logger)
	if err := ui.Run(); err != nil {
		logger.ErrorLog(err.Error())
	}
}
