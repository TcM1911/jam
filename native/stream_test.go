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
	"strings"
	"testing"
	"time"

	"github.com/TcM1911/jamsonic"

	"github.com/stretchr/testify/assert"
)

func TestNormalStreamRead(t *testing.T) {
	assert := assert.New(t)

	readerContent := "test reader content!"
	contentLength := len(readerContent)

	t.Run("normal read", func(t *testing.T) {
		testReader := strings.NewReader(readerContent)
		s := &stream{reader: testReader}
		buf := make([]byte, contentLength)

		n, err := s.Read(buf)

		assert.NoError(err)
		assert.Equal(n, contentLength)
		assert.Equal(readerContent, string(buf))
	})

	t.Run("multiple reads", func(t *testing.T) {
		testReader := strings.NewReader(readerContent)
		s := &stream{reader: testReader}
		buf1 := make([]byte, contentLength/2)
		buf2 := make([]byte, contentLength/2)

		n1, err1 := s.Read(buf1)
		n2, err2 := s.Read(buf2)
		readBuf := append(buf1, buf2...)

		assert.NoError(err1)
		assert.NoError(err2)
		assert.Equal(n1+n2, contentLength)
		assert.Equal(readerContent, string(readBuf))
	})
}

func TestRetryReads(t *testing.T) {
	assert := assert.New(t)
	readerContent := "test reader content!"

	// Save the default so we can reset it after these tests
	oldWaitTime := jamsonic.BufferingWait
	jamsonic.BufferingWait = time.Duration(20 * time.Millisecond)

	var newStream func(n int) *stream = func(n int) *stream {
		return &stream{
			reader: &failingReader{
				numbFails: n,
				reader:    strings.NewReader(readerContent),
			},
		}
	}

	t.Run("fail first read", func(t *testing.T) {
		s := newStream(1)
		buf := make([]byte, len(readerContent))
		max := time.Duration(int64(float64(jamsonic.BufferingWait) * 1.5))
		min := time.Duration(int64(float64(jamsonic.BufferingWait) * 0.5))

		start := time.Now()
		n, err := s.Read(buf)
		readTime := time.Since(start)

		assert.NoError(err)
		assert.Equal(len(readerContent), n)
		assert.Equal(readerContent, string(buf))
		assert.True(readTime < max)
		assert.True(readTime > min)
	})

	t.Run("fail first two reads", func(t *testing.T) {
		s := newStream(2)
		buf := make([]byte, len(readerContent))
		expectedRuntime := jamsonic.BufferingWait + jamsonic.BufferingWait<<1
		max := time.Duration(int64(float64(expectedRuntime) * 1.2))
		min := time.Duration(int64(float64(expectedRuntime) * 0.8))

		start := time.Now()
		n, err := s.Read(buf)
		readTime := time.Since(start)

		assert.NoError(err)
		assert.Equal(len(readerContent), n)
		assert.Equal(readerContent, string(buf))
		assert.True(readTime < max)
		assert.True(readTime > min)
	})

	t.Run("EOF on more than max", func(t *testing.T) {
		s := newStream(jamsonic.MaxReadRetryAttempts + 1)
		buf := make([]byte, len(readerContent))
		expectedRuntime := jamsonic.BufferingWait
		for i := 1; i < jamsonic.MaxReadRetryAttempts-1; i++ {
			expectedRuntime += (jamsonic.BufferingWait << uint(i))
		}
		max := time.Duration(int64(float64(expectedRuntime) * 1.1))
		min := time.Duration(int64(float64(expectedRuntime) * 0.9))

		start := time.Now()
		n, err := s.Read(buf)
		readTime := time.Since(start)

		assert.Error(err)
		assert.Equal(0, n)
		assert.Equal(io.EOF, err)
		assert.True(readTime < max)
		assert.True(readTime > min)
	})

	jamsonic.BufferingWait = oldWaitTime
}

type failingReader struct {
	attempts  int
	numbFails int
	reader    io.Reader
}

func (r *failingReader) Read(p []byte) (int, error) {
	if r.attempts < r.numbFails {
		r.attempts += 1
		return 0, io.EOF
	}
	return r.reader.Read(p)
}
