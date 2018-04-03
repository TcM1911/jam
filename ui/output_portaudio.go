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

package ui

import (
	"bytes"
	"encoding/binary"
	"log"

	"github.com/gordonklaus/portaudio"
)

const (
	numInputChans    = 0
	numOutputChans   = 2
	sampleRate       = 44100
	outputBufferSize = 8192 / 2
)

// portaudioOutputStream implements the OutputStream interface and
// the io.WriteCloser interface.
type portaudioOutputStream struct {
	stream *portaudio.Stream
	buf    []int16
}

func makeOutputStream() (OutputStream, error) {
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
	if err := p.stream.Close(); err != nil {
		p.stream = nil
		return err
	}
	p.stream = nil
	return nil
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
