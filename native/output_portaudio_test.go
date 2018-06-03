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

package native

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testSampleRate = 44100

func TestPortaudio(t *testing.T) {
	assert := assert.New(t)

	// For some reason calling CloseStream on Travis panics.
	// Skip these tests on Travis.
	if os.Getenv("TRAVIS_GO_VERSION") != "" {
		t.SkipNow()
	}

	t.Run("should create outputstream and close", func(t *testing.T) {
		out, err := newOutputWriter(testSampleRate)
		require.NoError(t, err, "makeOutputStream should not fail")
		assert.NotNil(out, "Output stream should not be nil")
		err = out.Close()
		assert.NoError(err, "Should close without error")
	})

	t.Run("write bytes", func(t *testing.T) {
		w, err := newOutputWriter(testSampleRate)
		require.NoError(t, err, "Should not fail")
		buf := make([]byte, inputBufferSize)
		_, err = w.Write(buf)
		assert.NoError(err, "Should write without error")
		w.Close()
	})
}
