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

package jamsonic

import (
	"bytes"
	"io"
	"sync"
	"time"
)

// State is the state the player can be in.
type State int8

const (
	// Stopped is the state when the player is not playing anything.
	Stopped State = iota
	// Playing is the state when the player is playing a stream.
	Playing
	// Paused is the state when the player has been paused.
	Paused
)

// NewPlayer returns a new Player. The Provider should be a music provider.
// The callback is a function that is called every interval by the Player as long as the state is
// not stopped. If interval is set to 0, the callback will be called every 1000 ms.
// The callback function can be used to update the UI with current play status etc.
func NewPlayer(p Provider, h StreamHandler, callback func(*CallbackData), interval int) *Player {
	if interval == 0 {
		interval = 1000
	}
	player := &Player{
		handler:          h,
		provider:         p,
		Error:            make(chan error),
		callback:         callback,
		callbackInterval: interval,
		closeChan:        make(chan struct{}),
		playChan:         make(chan struct{}),
		pauseChan:        make(chan struct{}),
		nextChan:         make(chan struct{}),
		prevChan:         make(chan struct{}),
		stopChan:         make(chan struct{}),
		played:           &playqueue{array: make([]*Track, 0)},
		buffer:           newBufReadWriter(),
	}
	player.buffer.bufferedMu.Lock()
	player.buffer.buffered = true
	player.buffer.bufferedMu.Unlock()
	go player.playerLoop()
	return player
}

// CallbackData is passed in to the player's callback function.
// It holds current values from right before the callback function was called.
type CallbackData struct {
	// CurrentTrack is the track being played.
	CurrentTrack *Track
	// Duration is how long the current track has been played.
	Duration time.Duration
}

// Player is the mp3 player struct. This struct handles all player actions.
type Player struct {
	// Error returns errors from the handler and the provider.
	Error            chan error
	handler          StreamHandler
	provider         Provider
	callback         func(*CallbackData)
	callbackInterval int
	queue            *playqueue
	played           *playqueue
	queueMu          sync.RWMutex
	currentTrack     *Track
	currentTrackMu   sync.RWMutex
	state            State
	stateMu          sync.RWMutex
	playChan         chan struct{}
	stopChan         chan struct{}
	pauseChan        chan struct{}
	nextChan         chan struct{}
	prevChan         chan struct{}
	closeChan        chan struct{}
	bufMu            sync.Mutex
	buffer           *bufReadWriter
}

// Play starts or resumes playing the track first in the play queue.
func (p *Player) Play() {
	p.playChan <- struct{}{}
}

// Pause pauses or resumes playing a track.
func (p *Player) Pause() {
	state := p.GetCurrentState()
	if state == Playing {
		p.pauseChan <- struct{}{}
	} else if state == Paused {
		p.Play()
	}
}

// Next skips to the next track in the play queue.
func (p *Player) Next() {
	p.nextChan <- struct{}{}
}

// Previous will go back to previous played track.
func (p *Player) Previous() {
	p.prevChan <- struct{}{}
}

// Stop should be called to stop playing the track.
func (p *Player) Stop() {
	p.stopChan <- struct{}{}
}

// GetCurrentState returns the player's current internal state.
func (p *Player) GetCurrentState() State {
	p.stateMu.RLock()
	defer p.stateMu.RUnlock()
	return p.state
}

// CreatePlayQueue creates a new list with queued tracks.
func (p *Player) CreatePlayQueue(tracks []*Track) {
	p.queue = &playqueue{array: tracks}
}

// NextTrack returns the next track in the play queue.
func (p *Player) NextTrack() *Track {
	return p.queue.nextSong()
}

// CurrentTrack returns the current playing or paused track.
func (p *Player) CurrentTrack() *Track {
	p.currentTrackMu.RLock()
	defer p.currentTrackMu.RUnlock()
	return p.currentTrack
}

// Close closes the player.
func (p *Player) Close() {
	p.closeChan <- struct{}{}
}

func (p *Player) updateCurrentTrack(t *Track) {
	p.currentTrackMu.Lock()
	defer p.currentTrackMu.Unlock()
	p.currentTrack = t
}

func (p *Player) playerLoop() {
	finished := p.handler.Finished()
	ticker := time.NewTicker(time.Millisecond * time.Duration(p.callbackInterval))
	// Timers
	pausedDuration := time.Duration(0)
	songDuration := time.Duration(0)
	var pauseTimer time.Time
	var songStart time.Time
	for {
		select {
		case <-p.playChan:
			status := p.changeState(Playing)
			if status == Paused {
				p.handler.Continue()
				pausedDuration = pausedDuration + time.Since(pauseTimer)
				continue
			}
			p.playNextInQueue(p.queue.popSong)
			songStart = time.Now()
			pausedDuration = time.Duration(0)
		case <-p.pauseChan:
			p.handler.Pause()
			p.changeState(Paused)
			pauseTimer = time.Now()
		case <-p.stopChan:
			if p.GetCurrentState() == Stopped {
				continue
			}
			ct := p.CurrentTrack()
			if ct != nil {
				p.queue.pushSong(ct)
			}
			p.stopPlaying()
			pausedDuration = time.Duration(0)
		case <-p.nextChan:
			state := p.GetCurrentState()
			if state == Stopped {
				continue
			}
			if state == Paused {
				p.changeState(Playing)
			}
			ct := p.CurrentTrack()
			if ct != nil {
				p.played.pushSong(ct)
			}
			p.playNextInQueue(p.queue.popSong)
			songStart = time.Now()
			pausedDuration = time.Duration(0)
		case <-p.prevChan:
			state := p.GetCurrentState()
			if state == Stopped {
				continue
			}
			if p.played.nextSong() == nil {
				continue
			}
			p.queue.pushSong(p.CurrentTrack())
			p.playNextInQueue(p.played.popSong)
			songStart = time.Now()
			pausedDuration = time.Duration(0)
		case <-finished:
			ct := p.CurrentTrack()
			if ct != nil {
				p.played.pushSong(ct)
			}
			if p.NextTrack() == nil {
				p.stopPlaying()
				continue
			}
			p.playNextInQueue(p.queue.popSong)
			songStart = time.Now()
			pausedDuration = time.Duration(0)
		case <-p.closeChan:
			return
		case <-ticker.C:
			if p.GetCurrentState() == Stopped {
				continue
			}
			if p.callback != nil {
				if p.GetCurrentState() == Playing {
					songDuration = time.Since(songStart) - pausedDuration
				}
				data := &CallbackData{
					CurrentTrack: p.CurrentTrack(),
					Duration:     songDuration,
				}
				p.callback(data)
			}
		}
	}
}

func (p *Player) playNextInQueue(getTrack func() *Track) {
	p.queueMu.Lock()
	ct := getTrack()
	p.updateCurrentTrack(ct)
	p.queueMu.Unlock()

	stream, err := p.provider.GetStream(ct.ID)
	if err != nil {
		handleStreamError(p, err)
		return
	}
	// If buffered, just reset and reuse the buffer.
	p.buffer.bufferedMu.Lock()
	p.bufMu.Lock()
	tmp := p.buffer
	if p.buffer.buffered {
		p.buffer.Reset()
	} else {
		// Since we can't stop to copy, create a new buffer
		// and let the GC clean up the old buffer.
		p.buffer = newBufReadWriter()
	}
	p.bufMu.Unlock()
	tmp.bufferedMu.Unlock()

	go func() {
		p.bufMu.Lock()
		writer := p.buffer
		p.bufMu.Unlock()
		_, err := io.Copy(writer, stream)
		defer stream.Close()
		if err != nil {
			p.Error <- err
			return
		}
		writer.bufferedMu.Lock()
		writer.buffered = true
		writer.bufferedMu.Unlock()
	}()
	p.bufMu.Lock()
	reader := p.buffer
	p.bufMu.Unlock()
	err = p.handler.Play(reader)
	if err != nil {
		handleStreamError(p, err)
	}
}

func (p *Player) stopPlaying() {
	p.handler.Stop()
	p.updateCurrentTrack(nil)
	p.changeState(Stopped)
}

// changeState changes the state to the new but also returns the previous state.
func (p *Player) changeState(s State) State {
	p.stateMu.Lock()
	defer p.stateMu.Unlock()
	status := p.state
	p.state = s
	return status
}

func handleStreamError(p *Player, err error) {
	go func() {
		p.Error <- err
	}()
	p.changeState(Stopped)
	p.queueMu.Lock()
	p.queue.pushSong(p.CurrentTrack())
	p.queueMu.Unlock()
}

type playqueue struct {
	arrayMu sync.RWMutex
	array   []*Track
}

// nextSong returns the next song in the play queue.
func (q *playqueue) nextSong() *Track {
	q.arrayMu.RLock()
	defer q.arrayMu.RUnlock()
	if len(q.array) >= 1 {
		return q.array[0]
	}
	return nil
}

// popSong returns the next song in the play queue and removes it from the queue.
func (q *playqueue) popSong() *Track {
	track := q.nextSong()
	q.arrayMu.Lock()
	defer q.arrayMu.Unlock()
	tmp := make([]*Track, len(q.array)-1)
	copy(tmp, q.array[1:])
	q.array = tmp
	return track
}

func (q *playqueue) pushSong(t *Track) {
	q.arrayMu.Lock()
	defer q.arrayMu.Unlock()
	tmp := make([]*Track, len(q.array)+1)
	tmp[0] = t
	for i, tr := range q.array {
		tmp[i+1] = tr
	}
	q.array = tmp
}

// StreamHandler should handle streams for the player. It should take an io.Reader
// and do the decoding to play the track.
type StreamHandler interface {
	// Finished should return a chan that is used to signal to the Player when the
	// track has been processed. The Player will call Play with a Reader for the next track
	// in the queue if one exists or call Stop if it was the final track in the playing queue.
	Finished() <-chan struct{}
	// Play is called with an io.Reader for the track. The handler should decode the stream and
	// send it to an output writer.
	Play(io.Reader) error
	// Stop is called by the Player to signal that all processing should stop. It is recommended that
	// output writers is closed when this is called.
	Stop()
	// Pause is called when stream processing should be paused.
	Pause()
	// Continue is called when stream processing should be resumed after it has been paused.
	Continue()
}

type bufReadWriter struct {
	mu         sync.Mutex
	buf        *bytes.Buffer
	bufferedMu sync.Mutex
	buffered   bool
}

func newBufReadWriter() *bufReadWriter {
	buf := new(bytes.Buffer)
	return &bufReadWriter{buf: buf}
}

func (b *bufReadWriter) Read(a []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Read(a)
}

func (b *bufReadWriter) Write(a []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(a)
}

func (b *bufReadWriter) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.buf.Reset()
	b.buffered = false
}
