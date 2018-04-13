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

// +build darwin

package native

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPortaudio(t *testing.T) {
	assert := assert.New(t)

	// For some reason calling CloseStream on Travis panics.
	// Skip these tests on Travis.
	if os.Getenv("TRAVIS_GO_VERSION") != "" {
		t.SkipNow()
	}

	t.Run("should create outputstream and close", func(t *testing.T) {
		out, err := makeOutputStream()
		assert.NoError(err, "makeOutputStream should not fail")
		assert.NotNil(out, "Output stream should not be nil")
		err = out.CloseStream()
		assert.NoError(err, "Should close without error")
	})

	t.Run("write bytes", func(t *testing.T) {
		w, err := makeOutputStream()
		assert.NoError(err, "Should not fail")
		buf := make([]byte, inputBufferSize)
		_, err = w.Write(buf)
		assert.NoError(err, "Should write without error")
		w.CloseStream()
	})
}