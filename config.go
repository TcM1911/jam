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

package jamsonic

import (
	"time"
)

// BufferingWait is the time in milliseconds to wait on reading from the network socket
// before playing the track. If this is to low, an EOF can be returned when reading
// from the memory buffer causing the track from being skipped.
// Default value is 200 ms.
var BufferingWait = time.Duration(200 * time.Millisecond)

// MaxReadRetryAttempts is the number of retries when a EOF is returned during the first read.
var MaxReadRetryAttempts = 5
