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

	"github.com/korandiz/mpa"
)

var (
	inputBufferSize = 1024 * 8
)

var makeOutputStream func() (OutputStream, error)

// OutputStream define an output stream
type OutputStream interface {
	CloseStream() error
	Write(data []byte) (int, error)
}

type StreamHandler struct {
	writer       OutputStream
	writerMu     sync.Mutex
	reader       io.Reader
	finishedChan chan struct{}
	stopChan     chan struct{}
	newTrackChan chan *io.Reader
	errChan      chan error
	pauseChan    chan struct{}
	continueChan chan struct{}
}

func New() *StreamHandler {
	return &StreamHandler{
		finishedChan: make(chan struct{}),
		stopChan:     make(chan struct{}),
		newTrackChan: make(chan *io.Reader),
		errChan:      make(chan error),
		pauseChan:    make(chan struct{}),
		continueChan: make(chan struct{}),
	}
}

func (p *StreamHandler) closeOutput() {
	p.writerMu.Lock()
	defer p.writerMu.Unlock()
	p.writer.CloseStream()
	p.writer = nil
}

func (p *StreamHandler) Errors() <-chan error {
	return p.errChan
}

func (p *StreamHandler) Play(stream io.Reader) error {
	p.writerMu.Lock()
	defer p.writerMu.Unlock()
	// Nothing playing
	if p.writer == nil {
		writer, err := makeOutputStream()
		if err != nil {
			return err
		}
		p.writer = writer
		go mainLoop(p, stream)
	} else {
		// Already playing a track, telling to switch stream.
		p.newTrackChan <- &stream
	}
	return nil
}

func mainLoop(p *StreamHandler, stream io.Reader) {
	defer p.closeOutput()
	p.reader = newDecoder(&stream)
	buf := make([]byte, inputBufferSize)
	for {
		select {
		// Stop playing
		case <-p.stopChan:
			return
		// Pause
		case <-p.pauseChan:
			select {
			case <-p.continueChan:
				continue
			case <-p.stopChan:
				return
			case r := <-p.newTrackChan:
				p.reader = newDecoder(r)
				continue
			}
		// New stream
		case r := <-p.newTrackChan:
			p.reader = newDecoder(r)
		// Play
		default:
			_, err := p.reader.Read(buf)
			if err == io.EOF {
				// Finished with this Track. Tell controller we are done.
				p.finishedChan <- struct{}{}
				select {
				// New track
				case r := <-p.newTrackChan:
					p.reader = newDecoder(r)
					continue
				case <-p.stopChan:
					return
				}
			} else if err != nil {
				p.errChan <- err
				return
			}
			_, err = p.writer.Write(buf)
			if err != nil {
				p.errChan <- err
				return
			}
		}
	}
}

func (p *StreamHandler) Stop() {
	p.stopChan <- struct{}{}
}

func (p *StreamHandler) Pause() {
	p.pauseChan <- struct{}{}
}

func (p *StreamHandler) Continue() {
	p.continueChan <- struct{}{}
}

func (p *StreamHandler) Finished() <-chan struct{} {
	return p.finishedChan
}

var newDecoder func(*io.Reader) io.Reader = func(r *io.Reader) io.Reader {
	return &mpa.Reader{Decoder: &mpa.Decoder{Input: *r}}
}
