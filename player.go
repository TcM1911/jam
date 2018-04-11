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

type playqueue struct {
	arrayMu sync.RWMutex
	array   []*Track
}

/*func newQueue(tracks []*Track) *playqueue {
	return &playqueue{array: tracks}
}*/

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

func NewPlayer(p Provider, h StreamHandler, cb func(*CallbackData), cd int) *Player {
	if cd == 0 {
		cd = 1
	}
	player := &Player{
		Handler:          h,
		Provider:         p,
		Error:            make(chan error),
		Callback:         cb,
		CallbackInterval: cd,
		closeChan:        make(chan struct{}),
		playChan:         make(chan struct{}),
		pauseChan:        make(chan struct{}),
		nextChan:         make(chan struct{}),
		prevChan:         make(chan struct{}),
		stopChan:         make(chan struct{}),
		played:           &playqueue{array: make([]*Track, 0)},
	}
	go player.playerLoop()
	return player
}

type CallbackData struct {
	CurrentTrack *Track
	Duration     time.Duration
}

type Player struct {
	Handler          StreamHandler
	Provider         Provider
	Error            chan error
	Callback         func(*CallbackData)
	CallbackInterval int
	queue            *playqueue
	played           *playqueue
	queueMu          sync.RWMutex
	currentTrack     *Track
	currentTrackMu   sync.RWMutex
	currentStream    io.ReadCloser
	state            State
	stateMu          sync.RWMutex
	playChan         chan struct{}
	stopChan         chan struct{}
	pauseChan        chan struct{}
	nextChan         chan struct{}
	prevChan         chan struct{}
	closeChan        chan struct{}
}

func (p *Player) Play() {
	p.playChan <- struct{}{}
}

func (p *Player) Pause() {
	state := p.GetCurrentState()
	if state == Playing {
		p.pauseChan <- struct{}{}
	} else if state == Paused {
		p.Play()
	}
}

func (p *Player) Next() {
	p.nextChan <- struct{}{}
}

func (p *Player) Previous() {
	p.prevChan <- struct{}{}
}

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

func (p *Player) updateCurrentTrack(t *Track) {
	p.currentTrackMu.Lock()
	defer p.currentTrackMu.Unlock()
	p.currentTrack = t
}

func (p *Player) playerLoop() {
	finished := p.Handler.Finished()
	ticker := time.NewTicker(time.Second * time.Duration(p.CallbackInterval))
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
				p.Handler.Continue()
				pausedDuration = pausedDuration + time.Since(pauseTimer)
				continue
			}
			p.playNextInQueue(p.queue.popSong)
			songStart = time.Now()
		case <-p.pauseChan:
			p.Handler.Pause()
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
			pausedDuration = time.Duration(0)
		case <-p.closeChan:
			return
		case <-ticker.C:
			if p.GetCurrentState() == Stopped {
				continue
			}
			if p.Callback != nil {
				if p.GetCurrentState() == Playing {
					songDuration = time.Since(songStart) - pausedDuration
				}
				data := &CallbackData{
					CurrentTrack: p.CurrentTrack(),
					Duration:     songDuration,
				}
				p.Callback(data)
			}
		}
	}
}

func (p *Player) Close() {
	p.closeChan <- struct{}{}
}

func (p *Player) playNextInQueue(getTrack func() *Track) {
	p.queueMu.Lock()
	ct := getTrack()
	p.updateCurrentTrack(ct)
	p.queueMu.Unlock()

	// Save old stream so we can close it.
	oldStream := p.currentStream

	stream, err := p.Provider.GetStream(ct.ID)
	p.currentStream = stream
	if err != nil {
		handleStreamError(p, err)
	}
	err = p.Handler.Play(stream)
	if err != nil {
		handleStreamError(p, err)
	}
	if oldStream != nil {
		oldStream.Close()
	}
}

func (p *Player) stopPlaying() {
	p.Handler.Stop()
	p.currentStream.Close()
	p.currentStream = nil
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
	p.currentStream.Close()
}

type StreamHandler interface {
	Finished() <-chan struct{}
	Play(io.Reader) error
	Stop()
	Pause()
	Continue()
}
