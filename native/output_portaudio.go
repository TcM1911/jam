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
	"bytes"
	"encoding/binary"
	"io"
	"log"

	"github.com/gordonklaus/portaudio"
)

const (
	numInputChans = 0
)

// portaudioOutputStream implements the OutputStream interface and
// the io.WriteCloser interface.
type portaudioOutputStream struct {
	stream *portaudio.Stream
	buf    []int16
}

var newOutputWriter func(sampleRate int) (io.WriteCloser, error) = func(sampleRate int) (io.WriteCloser, error) {
	out := &portaudioOutputStream{buf: make([]int16, outputBufferSize)}
	if err := portaudio.Initialize(); err != nil {
		return nil, err
	}
	stream, err := portaudio.OpenDefaultStream(
		numInputChans, numOutputChans, float64(sampleRate), len(out.buf), out.buf)
	if err != nil {
		return nil, err
	}
	out.stream = stream
	if err := stream.Start(); err != nil {
		return nil, err
	}
	return out, nil
}

// CloseStream closes the stream.
func (p *portaudioOutputStream) CloseStream() error {
	return p.Close()
}

// Close closes the stream.
func (p *portaudioOutputStream) Close() error {
	defer portaudio.Terminate()
	p.buf = make([]int16, 0)
	p.stream.Stop()
	err := p.stream.Close()
	p.stream = nil
	return err
}

// Write writes the data to portaudio's audio interface.
func (p *portaudioOutputStream) Write(data []byte) (int, error) {
	err := binary.Read(bytes.NewBuffer(data), binary.LittleEndian, p.buf)
	if err != nil {
		log.Println("Read error:", err.Error())
		return 0, err
	}
	err = p.stream.Write()

	// From portaudio docs:
	// On success PaNoError will be returned, or paOutputUnderflowed if
	// additional output data was inserted after the previous call and
	// before this call.
	if err == portaudio.OutputUnderflowed {
		// Handle the error since the player doesn't know what to do with it.
		err = nil
	}
	return outputBufferSize, err
}
