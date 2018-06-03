// Copyright (c) 2018 Joakim Kennedy
// Copyright (c) 2016, 2017 Evgeny Badin
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
