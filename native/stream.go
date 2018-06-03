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

package native

import (
	"io"
	"sync"
	"time"

	"github.com/TcM1911/jamsonic"
	mp3 "github.com/hajimehoshi/go-mp3"
)

var (
	numOutputChans  = 2
	inputBufferSize = 1024 / 2
	// Since input uses int8 and output uses int16, we need a buffer of half the size.
	outputBufferSize = inputBufferSize / 2
)

type stream struct {
	reader io.Reader
	once   sync.Once
}

func (s *stream) Close() error {
	// The stream closing is handled by the player.
	return nil
}

func (s *stream) Read(b []byte) (int, error) {
	var n int
	var err error
	// First read with back-off.
	s.once.Do(func() {
		n, err = s.reader.Read(b)
		// If we get an EOF we start to retry with backoffs.
		if err == io.EOF {
			n, err = s.retryWithBackoff(b, jamsonic.BufferingWait, 1)
		}
	})
	// Return if n or err was set by the Do function.
	// n != 0 is true if read worked.
	// err != nil is true if read failed.
	// n == 0 and err == nil if Do wasn't executed.
	if n != 0 || err != nil {
		return n, err
	}
	return s.reader.Read(b)
}

func (s *stream) retryWithBackoff(b []byte, wait time.Duration, attempt int) (int, error) {
	if attempt >= jamsonic.MaxReadRetryAttempts {
		return 0, io.EOF
	}
	time.Sleep(wait)
	n, err := s.reader.Read(b)
	if err == io.EOF {
		// Double the wait time and retry.
		newWait := wait << 1
		attempt += 1
		return s.retryWithBackoff(b, newWait, attempt)
	}
	return n, err
}

type mp3Stream interface {
	io.Reader
	SampleRate() int
}

var newDecoder func(io.ReadCloser) (mp3Stream, error) = func(r io.ReadCloser) (mp3Stream, error) {
	return mp3.NewDecoder(r)
}
