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
	"fmt"
	"io"
	"sync"

	"github.com/TcM1911/jamsonic"
)

type StreamHandler struct {
	logger       *jamsonic.Logger
	writer       io.WriteCloser
	writerMu     sync.Mutex
	reader       io.Reader
	finishedChan chan struct{}
	stopChan     chan struct{}
	newTrackChan chan *switchStream
	errChan      chan error
	pauseChan    chan struct{}
	continueChan chan struct{}
}

// New returns a new stream handler.
func New(logger *jamsonic.Logger) *StreamHandler {
	return &StreamHandler{
		logger:       logger,
		finishedChan: make(chan struct{}),
		stopChan:     make(chan struct{}),
		newTrackChan: make(chan *switchStream),
		errChan:      make(chan error),
		pauseChan:    make(chan struct{}),
		continueChan: make(chan struct{}),
	}
}

func (p *StreamHandler) closeOutput() {
	p.writerMu.Lock()
	defer p.writerMu.Unlock()
	if p.writer != nil {
		p.writer.Close()
	}
	p.writer = nil
}

// Errors gives a channel where all errors are sent to.
func (p *StreamHandler) Errors() <-chan error {
	return p.errChan
}

// Play starts processing the stream.
func (p *StreamHandler) Play(iostream io.Reader) error {
	s, err := newDecoder(&stream{reader: iostream})
	if err != nil {
		return err
	}
	p.logger.DebugLog(fmt.Sprintf("Sample Rate: %d", s.SampleRate()))
	// Nothing playing
	if p.writer == nil {
		p.logger.DebugLog("Nothing playing, starting the main loop.")
		writer, err := newOutputWriter(s.SampleRate())
		if err != nil {
			return err
		}
		p.writerMu.Lock()
		p.writer = writer
		p.writerMu.Unlock()
		go mainLoop(p, s)
	} else {
		p.logger.DebugLog("Switching track.")
		// Already playing a track, telling to switch stream.
		p.newTrackChan <- &switchStream{stream: s, sampleRate: s.SampleRate()}
	}
	return nil
}

func mainLoop(p *StreamHandler, stream io.Reader) {
	defer p.closeOutput()
	p.reader = stream
	buf := make([]byte, inputBufferSize)
	for {
		select {
		// Stop playing
		case <-p.stopChan:
			return
		// Pause
		case <-p.pauseChan:
			p.logger.DebugLog("Stream processing paused.")
			select {
			case <-p.continueChan:
				p.logger.DebugLog("Resuming stream processing.")
				continue
			case <-p.stopChan:
				return
			}
		// Play
		default:
			_, err := p.reader.Read(buf)
			if err == io.EOF {
				p.logger.DebugLog("Finished reading the stream.")
				// Finished with this Track. Tell controller we are done.
				p.finishedChan <- struct{}{}
				select {
				// New track
				case s := <-p.newTrackChan:
					err := p.switchStreams(s)
					if err != nil {
						p.logger.ErrorLog(fmt.Sprintf("Error when switching stream: %s\n", err.Error()))
						return
					}
					p.logger.DebugLog("Processing new stream.")
					continue
				case <-p.stopChan:
					p.logger.DebugLog("Closing handler.")
					return
				}
				// Ignoring io.ErrUnexpectedEOF. Write what we have in the buffer
				// and allow another read that returns an io.EOF.
				// Other errors are reported to the controller.
			} else if err != nil && err != io.ErrUnexpectedEOF {
				p.errChan <- err
			}
			_, err = p.writer.Write(buf)
			if err != nil {
				p.errChan <- err
			}
		}
	}
}

// Stop tells the handler to stop processing the current stream.
func (p *StreamHandler) Stop() {
	p.logger.DebugLog("Sending stop signal the main loop.")
	p.stopChan <- struct{}{}
}

// Pause tells the handler to pause the processing of the current stream.
func (p *StreamHandler) Pause() {
	p.logger.DebugLog("Sending pause signal the main loop.")
	p.pauseChan <- struct{}{}
}

// Continue tells the handler to resume the processing of the current stream.
func (p *StreamHandler) Continue() {
	p.logger.DebugLog("Sending resume signal the main loop.")
	p.continueChan <- struct{}{}
}

// Finished returns a channel that is used to signal that the handler is done
// processing the current stream.
func (p *StreamHandler) Finished() <-chan struct{} {
	return p.finishedChan
}

func (p *StreamHandler) switchStreams(s *switchStream) error {
	p.closeOutput()
	p.reader = s.stream
	return p.newWriter(s.sampleRate)
}

func (p *StreamHandler) newWriter(sampleRate int) error {
	p.writerMu.Lock()
	defer p.writerMu.Unlock()
	w, err := newOutputWriter(sampleRate)
	if err != nil {
		return err
	}
	p.writer = w
	return nil
}

type switchStream struct {
	stream     io.Reader
	sampleRate int
}
