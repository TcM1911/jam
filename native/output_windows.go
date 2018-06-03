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

// +build windows

package native

import (
	"fmt"
	"io"

	"github.com/koron/go-waveout"
)

type windowsOutputStream struct {
	Player *waveout.Player
}

func init() {
	// Changing to Windows specific function.
	newOutputWriter = func(sampleRate int) (io.WriteCloser, error) {
		player, err := waveout.NewWithBuffers(numOutputChans, sampleRate, 16, 8, outputBufferSize)
		if err != nil {
			return nil, fmt.Errorf("failed to create waveout: %s", err)
		}
		return &windowsOutputStream{
			Player: player,
		}, nil
	}
}

func (wos *windowsOutputStream) Close() error {
	return wos.Player.Close()
}

func (wos *windowsOutputStream) Write(data []byte) (int, error) {
	return wos.Player.Write(data)
}
