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
	"errors"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPlayControl(t *testing.T) {
	assert := assert.New(t)
	player, _, _, _ := getPlayer()
	tracks := []*Track{
		&Track{ID: "1"},
		&Track{ID: "2"},
		&Track{ID: "3"},
		&Track{ID: "4"},
	}
	player.CreatePlayQueue(tracks)

	t.Run("get state", func(t *testing.T) {
		assert.Equal(Stopped, player.GetCurrentState())
	})

	t.Run("change current state to playing", func(t *testing.T) {
		player.state = Playing
		assert.Equal(Playing, player.GetCurrentState(), "Should be set to playing")
	})

	t.Run("change current state to paused", func(t *testing.T) {
		player.state = Paused
		assert.Equal(Paused, player.GetCurrentState(), "Should be set to paused")
	})

	t.Run("change current state to stopped", func(t *testing.T) {
		player.state = Stopped
		assert.Equal(Stopped, player.GetCurrentState(), "Should be set to stopped")
	})

	t.Run("current when not playing", func(t *testing.T) {
		track := player.CurrentTrack()
		assert.Nil(track, "Track should be nil")
	})

	t.Run("next track before playing", func(t *testing.T) {
		track := player.NextTrack()
		assert.Equal(tracks[0], track, "Should point to first track before playing")
	})

	player.Close()

	t.Run("Stopping", func(t *testing.T) {
		p, _, provider, handler := getPlayer()
		p.CreatePlayQueue(tracks)

		// At stop nothing should happen.
		p.Stop()
		time.Sleep(time.Millisecond * 200)
		calledMu.RLock()
		assert.Equal(0, handler.calledPlay, "Wrong number of calls")
		assert.Equal(0, handler.calledStopped, "Wrong number of calls")
		calledMu.RUnlock()
		assert.Equal(Stopped, p.GetCurrentState(), "State should be playing.")
		assert.Nil(p.CurrentTrack(), "Current track should be nil.")
		assert.Equal(tracks[0], p.NextTrack(), "1st track should be marked as next")

		p.Play()
		time.Sleep(time.Millisecond * 200)
		calledMu.RLock()
		assert.Equal(1, handler.calledPlay, "Wrong number of calls")
		calledMu.RUnlock()
		provider.streamIDMu.RLock()
		assert.Equal(tracks[0].ID, provider.streamID, "Wrong track id called")
		provider.streamIDMu.RUnlock()
		assert.Equal(Playing, p.GetCurrentState(), "State should be playing.")
		assert.Equal(tracks[0], p.CurrentTrack(), "First track not playing")
		assert.Equal(tracks[1], p.NextTrack(), "2nd track should be marked as next")

		// Stop playing
		p.Stop()
		time.Sleep(time.Millisecond * 200)
		calledMu.RLock()
		assert.Equal(1, handler.calledPlay, "Wrong number of calls")
		assert.Equal(1, handler.calledStopped, "Wrong number of calls")
		calledMu.RUnlock()
		assert.Equal(Stopped, p.GetCurrentState(), "State should be playing.")
		assert.Nil(p.CurrentTrack(), "Current track should be nil.")
		assert.Equal(tracks[0], p.NextTrack(), "1st track should be marked as next")

		p.Close()
	})

	t.Run("Playing", func(t *testing.T) {
		p, finished, provider, handler := getPlayer()
		p.CreatePlayQueue(tracks)
		p.Play()
		time.Sleep(time.Millisecond * 200)
		calledMu.RLock()
		assert.Equal(1, handler.calledPlay, "Wrong number of calls")
		calledMu.RUnlock()
		provider.streamIDMu.RLock()
		assert.Equal(tracks[0].ID, provider.streamID, "Wrong track id called")
		provider.streamIDMu.RUnlock()
		assert.Equal(Playing, p.GetCurrentState(), "State should be playing.")
		assert.Equal(tracks[0], p.CurrentTrack(), "First track not playing")
		assert.Equal(tracks[1], p.NextTrack(), "2nd track should be marked as next")

		// Second track.
		time.Sleep(time.Millisecond * 200)
		finished <- struct{}{}
		time.Sleep(time.Millisecond * 200)
		assert.Equal(Playing, p.GetCurrentState(), "State should be playing.")
		calledMu.RLock()
		assert.Equal(2, handler.calledPlay, "Wrong number of calls")
		calledMu.RUnlock()
		provider.streamIDMu.RLock()
		assert.Equal(tracks[1].ID, provider.streamID, "Wrong track id called")
		provider.streamIDMu.RUnlock()
		assert.Equal(tracks[1], p.CurrentTrack(), "2nd track not playing")
		assert.Equal(tracks[2], p.NextTrack(), "3rd track should be marked as next")
		assert.Equal(tracks[0], p.played.nextSong(), "1st track should be first in the played list")

		// Third track
		time.Sleep(time.Millisecond * 200)
		finished <- struct{}{}
		time.Sleep(time.Millisecond * 200)
		assert.Equal(Playing, p.GetCurrentState(), "State should be playing.")
		calledMu.RLock()
		assert.Equal(3, handler.calledPlay, "Wrong number of calls")
		calledMu.RUnlock()
		provider.streamIDMu.RLock()
		assert.Equal(tracks[2].ID, provider.streamID, "Wrong track id called")
		provider.streamIDMu.RUnlock()
		assert.Equal(tracks[2], p.CurrentTrack(), "3rd track not playing")
		assert.Equal(tracks[3], p.NextTrack(), "4th track should be marked as next")
		assert.Equal(tracks[1], p.played.nextSong(), "2nd track should be first in the played list")

		// 4th track
		time.Sleep(time.Millisecond * 200)
		finished <- struct{}{}
		time.Sleep(time.Millisecond * 200)
		assert.Equal(Playing, p.GetCurrentState(), "State should be playing.")
		calledMu.RLock()
		assert.Equal(4, handler.calledPlay, "Wrong number of calls")
		calledMu.RUnlock()
		provider.streamIDMu.RLock()
		assert.Equal(tracks[3].ID, provider.streamID, "Wrong track id called")
		provider.streamIDMu.RUnlock()
		assert.Equal(tracks[3], p.CurrentTrack(), "4th track not playing")
		assert.Nil(p.NextTrack(), "Nil should be returned for next track")
		assert.Equal(tracks[2], p.played.nextSong(), "3rd track should be first in the played list")

		// Stop when queue is empty
		time.Sleep(time.Millisecond * 200)
		finished <- struct{}{}
		time.Sleep(time.Millisecond * 200)
		assert.Equal(Stopped, p.GetCurrentState(), "State should be stopped.")
		assert.Nil(p.CurrentTrack(), "Current track should be nil")
		assert.Nil(p.NextTrack(), "Nil should be returned for next track")
		assert.Equal(tracks[3], p.played.nextSong(), "4th track should be first in the played list")

		p.Close()
	})

	t.Run("Pausing", func(t *testing.T) {
		p, _, provider, handler := getPlayer()
		p.CreatePlayQueue(tracks)
		p.Play()

		// Ensure initial state is correct.
		time.Sleep(time.Millisecond * 200)
		provider.streamIDMu.RLock()
		assert.Equal(tracks[0].ID, provider.streamID, "Wrong track id called")
		provider.streamIDMu.RUnlock()
		calledMu.RLock()
		assert.Equal(1, handler.calledPlay, "Wrong number of calls")
		assert.Equal(0, handler.calledPause, "Wrong number of calls")
		assert.Equal(0, handler.calledContrinue, "Wrong number of calls")
		calledMu.RUnlock()
		assert.Equal(Playing, p.GetCurrentState(), "State should be playing.")
		assert.Equal(tracks[0], p.CurrentTrack(), "First track not playing")
		assert.Equal(tracks[1], p.NextTrack(), "2nd track should be marked as next")

		// Pause during the first track.
		p.Pause()
		time.Sleep(time.Millisecond * 200)
		assert.Equal(Paused, p.GetCurrentState(), "State should be paused.")
		calledMu.RLock()
		assert.Equal(1, handler.calledPlay, "Wrong number of calls")
		assert.Equal(1, handler.calledPause, "Wrong number of calls")
		assert.Equal(0, handler.calledContrinue, "Wrong number of calls")
		calledMu.RUnlock()
		assert.Equal(tracks[0], p.CurrentTrack(), "First track not playing")
		assert.Equal(tracks[1], p.NextTrack(), "2nd track should be marked as next")

		// Continue
		p.Play()
		time.Sleep(time.Millisecond * 200)
		assert.Equal(Playing, p.GetCurrentState(), "State should be paused.")
		calledMu.RLock()
		assert.Equal(1, handler.calledPlay, "Wrong number of calls")
		assert.Equal(1, handler.calledPause, "Wrong number of calls")
		assert.Equal(1, handler.calledContrinue, "Wrong number of calls")
		calledMu.RUnlock()
		assert.Equal(tracks[0], p.CurrentTrack(), "First track not playing")
		assert.Equal(tracks[1], p.NextTrack(), "2nd track should be marked as next")

		p.Close()
	})

	t.Run("Song skipping", func(t *testing.T) {
		p, _, provider, handler := getPlayer()
		p.CreatePlayQueue(tracks)

		// Skipping during stopped shouldn't do anything.
		assert.Equal(Stopped, p.GetCurrentState(), "Wrong state")
		p.Next()
		time.Sleep(time.Millisecond * 200)
		assert.Equal(Stopped, p.GetCurrentState(), "Wrong state")
		assert.Nil(p.CurrentTrack(), "Wrong track")
		assert.Equal(tracks[0], p.NextTrack(), "1st track should be marked as next")

		// Ensure initial state is correct.
		p.Play()
		time.Sleep(time.Millisecond * 200)
		provider.streamIDMu.RLock()
		assert.Equal(tracks[0].ID, provider.streamID, "Wrong track id called")
		provider.streamIDMu.RUnlock()
		calledMu.RLock()
		assert.Equal(1, handler.calledPlay, "Wrong number of calls")
		assert.Equal(0, handler.calledPause, "Wrong number of calls")
		assert.Equal(0, handler.calledContrinue, "Wrong number of calls")
		calledMu.RUnlock()
		assert.Equal(Playing, p.GetCurrentState(), "State should be playing.")
		assert.Equal(tracks[0], p.CurrentTrack(), "First track not playing")
		assert.Equal(tracks[1], p.NextTrack(), "2nd track should be marked as next")

		// Skip to next track.
		p.Next()
		time.Sleep(time.Millisecond * 200)
		assert.Equal(Playing, p.GetCurrentState(), "State should be playing.")
		calledMu.RLock()
		assert.Equal(2, handler.calledPlay, "Wrong number of calls")
		assert.Equal(0, handler.calledPause, "Wrong number of calls")
		assert.Equal(0, handler.calledContrinue, "Wrong number of calls")
		calledMu.RUnlock()
		provider.streamIDMu.RLock()
		assert.Equal(tracks[1].ID, provider.streamID, "Wrong track id called")
		provider.streamIDMu.RUnlock()
		assert.Equal(tracks[1], p.CurrentTrack(), "2nd track not playing")
		assert.Equal(tracks[2], p.NextTrack(), "3rd track should be marked as next")
		assert.Equal(tracks[0], p.played.nextSong(), "1st track should be first in the played list")

		// Skip from paused state.
		p.Pause()
		time.Sleep(time.Millisecond * 200)
		assert.Equal(Paused, p.GetCurrentState(), "Wrong state")
		p.Next()
		time.Sleep(time.Millisecond * 200)
		assert.Equal(Playing, p.GetCurrentState(), "State should be playing.")
		calledMu.RLock()
		assert.Equal(3, handler.calledPlay, "Wrong number of calls")
		assert.Equal(1, handler.calledPause, "Wrong number of calls")
		assert.Equal(0, handler.calledContrinue, "Wrong number of calls")
		calledMu.RUnlock()
		provider.streamIDMu.RLock()
		assert.Equal(tracks[2].ID, provider.streamID, "Wrong track id called")
		provider.streamIDMu.RUnlock()
		assert.Equal(tracks[2], p.CurrentTrack(), "3rd track not playing")
		assert.Equal(tracks[3], p.NextTrack(), "4th track should be marked as next")
		assert.Equal(tracks[1], p.played.nextSong(), "2nd track should be first in the played list")

		p.Close()
	})

	t.Run("Song reverse skipping", func(t *testing.T) {
		p, _, provider, handler := getPlayer()
		p.CreatePlayQueue(tracks)

		// Skipping back during stopped shouldn't do anything.
		assert.Equal(Stopped, p.GetCurrentState(), "Wrong state")
		p.Previous()
		time.Sleep(time.Millisecond * 200)
		assert.Equal(Stopped, p.GetCurrentState(), "Wrong state")
		assert.Nil(p.CurrentTrack(), "Wrong track")
		assert.Equal(tracks[0], p.NextTrack(), "1st track should be marked as next")

		// Skipping if no songs have been played shouldn't do anything.
		p.Play()
		p.Previous()
		time.Sleep(time.Millisecond * 200)
		assert.Equal(Playing, p.GetCurrentState(), "Wrong state")
		provider.streamIDMu.RLock()
		assert.Equal(tracks[0].ID, provider.streamID, "Wrong track id called")
		provider.streamIDMu.RUnlock()
		assert.Equal(tracks[0], p.CurrentTrack(), "1st track not playing")
		assert.Equal(tracks[1], p.NextTrack(), "2nd track should be marked as next")
		assert.Nil(p.played.nextSong(), "Nil should be returned for played.")

		// Ensure initial state is correct. First track in played list.
		p.Next()
		time.Sleep(time.Millisecond * 200)
		provider.streamIDMu.RLock()
		assert.Equal(tracks[1].ID, provider.streamID, "Wrong track id called")
		provider.streamIDMu.RUnlock()
		calledMu.RLock()
		assert.Equal(2, handler.calledPlay, "Wrong number of calls")
		assert.Equal(0, handler.calledPause, "Wrong number of calls")
		assert.Equal(0, handler.calledContrinue, "Wrong number of calls")
		calledMu.RUnlock()
		assert.Equal(Playing, p.GetCurrentState(), "State should be playing.")
		assert.Equal(tracks[1], p.CurrentTrack(), "2nd track not playing")
		assert.Equal(tracks[2], p.NextTrack(), "3rd track should be marked as next")
		assert.Equal(tracks[0], p.played.nextSong(), "1st track should be first in the played list")

		// Back to prev track.
		p.Previous()
		time.Sleep(time.Millisecond * 200)
		assert.Equal(Playing, p.GetCurrentState(), "State should be playing.")
		calledMu.RLock()
		assert.Equal(3, handler.calledPlay, "Wrong number of calls")
		assert.Equal(0, handler.calledPause, "Wrong number of calls")
		assert.Equal(0, handler.calledContrinue, "Wrong number of calls")
		calledMu.RUnlock()
		provider.streamIDMu.RLock()
		assert.Equal(tracks[0].ID, provider.streamID, "Wrong track id called")
		provider.streamIDMu.RUnlock()
		assert.Equal(tracks[0], p.CurrentTrack(), "1st track not playing")
		assert.Equal(tracks[1], p.NextTrack(), "2nd track should be marked as next")
		assert.Nil(p.played.nextSong(), "Nil should be returned for played.")

		p.Close()
	})

	t.Run("Callback", func(t *testing.T) {
		finishedChan := make(chan struct{})
		provider := &mockProvider{
			doGetStream: func(id string) (io.ReadCloser, error) {
				return &recorder{streamID: id}, nil
			},
		}
		handler := &mockStreaHandler{
			doFinished: func() <-chan struct{} { return finishedChan },
			doPlay:     func(io.Reader) error { return nil },
			doStop:     func() {},
			doPause:    func() {},
			doContinue: func() {},
		}
		var callbackTrack *Track
		var duration time.Duration
		var p *Player
		callback := func(data *CallbackData) {
			callbackTrack = data.CurrentTrack
			duration = data.Duration
			go p.Stop()
		}
		p = NewPlayer(provider, handler, callback, 1000)
		p.CreatePlayQueue(tracks)

		// Should not run the callback when the state is stopped.
		p.Stop()
		time.Sleep(time.Millisecond * 1100)
		assert.Equal(Stopped, p.GetCurrentState(), "State should be playing.")
		assert.Equal(time.Duration(0), duration, "Duration should be 0")
		assert.Nil(callbackTrack, "Track should be nil")

		// Should run callback when state is playing.
		p.Play()
		time.Sleep(time.Millisecond * 1100)
		// Callback stops the player
		assert.Equal(Stopped, p.GetCurrentState(), "State should be playing.")
		assert.True(duration >= time.Duration(800*time.Millisecond), "Too short duration")
		assert.Equal(tracks[0], callbackTrack, "Wrong track in the callback")

		p.Close()
	})

	t.Run("Handle errors from provider", func(t *testing.T) {
		finishedChan := make(chan struct{})
		expectedError := errors.New("expected error")
		provider := &mockProvider{
			doGetStream: func(id string) (io.ReadCloser, error) {
				return &recorder{streamID: id}, nil
			},
		}
		handler := &mockStreaHandler{
			doFinished: func() <-chan struct{} { return finishedChan },
			doPlay:     func(io.Reader) error { return expectedError },
			doStop:     func() {},
			doPause:    func() {},
			doContinue: func() {},
		}
		p := NewPlayer(provider, handler, nil, 0)
		p.CreatePlayQueue(tracks)

		p.Play()
		err := <-p.Error
		time.Sleep(time.Millisecond * 200)
		assert.Equal(Stopped, p.GetCurrentState(), "State should be playing.")
		assert.Equal(expectedError, err, "Returned wrong error")
		assert.Equal(tracks[0], p.CurrentTrack(), "Wrong track set as next")

		p.Close()
	})

	t.Run("Handle errors from handler", func(t *testing.T) {
		finishedChan := make(chan struct{})
		expectedError := errors.New("expected error")
		provider := &mockProvider{
			doGetStream: func(id string) (io.ReadCloser, error) {
				return nil, expectedError
			},
		}
		handler := &mockStreaHandler{
			doFinished: func() <-chan struct{} { return finishedChan },
			doPlay:     func(io.Reader) error { return nil },
			doStop:     func() {},
			doPause:    func() {},
			doContinue: func() {},
		}
		p := NewPlayer(provider, handler, nil, 0)
		p.CreatePlayQueue(tracks)

		p.Play()
		err := <-p.Error
		time.Sleep(time.Millisecond * 200)
		assert.Equal(Stopped, p.GetCurrentState(), "State should be playing.")
		assert.Equal(expectedError, err, "Returned wrong error")
		assert.Equal(tracks[0], p.CurrentTrack(), "Wrong track set as next")

		p.Close()
	})
}

func getPlayer() (*Player, chan struct{}, *mockProvider, *mockStreaHandler) {
	finishedChan := make(chan struct{})

	provider := &mockProvider{
		doGetStream: func(id string) (io.ReadCloser, error) {
			return &recorder{streamID: id}, nil
		},
	}
	handler := &mockStreaHandler{
		doFinished: func() <-chan struct{} { return finishedChan },
		doPlay:     func(io.Reader) error { return nil },
		doStop:     func() {},
		doPause:    func() {},
		doContinue: func() {},
	}
	return NewPlayer(provider, handler, nil, 0), finishedChan, provider, handler
}

type recorder struct {
	streamID string
}

func (r *recorder) Read([]byte) (int, error) {
	return 0, nil
}

func (r *recorder) Close() error {
	return nil
}

type mockStreaHandler struct {
	doFinished      func() <-chan struct{}
	doPlay          func(io.Reader) error
	calledPlay      int
	doStop          func()
	calledStopped   int
	doPause         func()
	calledPause     int
	doContinue      func()
	calledContrinue int
}

var calledMu sync.RWMutex

func (m *mockStreaHandler) Finished() <-chan struct{} {
	return m.doFinished()
}

func (m *mockStreaHandler) Play(r io.Reader) error {
	calledMu.Lock()
	defer calledMu.Unlock()
	m.calledPlay += 1
	return m.doPlay(r)
}

func (m *mockStreaHandler) Stop() {
	calledMu.Lock()
	defer calledMu.Unlock()
	m.calledStopped += 1
	m.doStop()
}

func (m *mockStreaHandler) Pause() {
	calledMu.Lock()
	defer calledMu.Unlock()
	m.calledPause += 1
	m.doPause()
}

func (m *mockStreaHandler) Continue() {
	calledMu.Lock()
	defer calledMu.Unlock()
	m.calledContrinue += 1
	m.doContinue()
}

type mockProvider struct {
	streamID              string
	streamIDMu            sync.RWMutex
	doListTracks          func() ([]*Track, error)
	doFetchLibrary        func() ([]*Artist, error)
	doGetTrackInfo        func(trackID string) (*Track, error)
	doGetStream           func(songID string) (io.ReadCloser, error)
	doListPlaylists       func() ([]*Playlist, error)
	doListPlaylistEntries func() ([]*PlaylistEntry, error)
	doGetProvider         func() MusicProvider
}

func (m *mockProvider) ListTracks() ([]*Track, error) {
	return m.doListTracks()
}

func (m *mockProvider) FetchLibrary() ([]*Artist, error) {
	return m.doFetchLibrary()
}

func (m *mockProvider) GetTrackInfo(trackID string) (*Track, error) {
	return m.doGetTrackInfo(trackID)
}

func (m *mockProvider) GetStream(songID string) (io.ReadCloser, error) {
	m.streamIDMu.Lock()
	defer m.streamIDMu.Unlock()
	m.streamID = songID
	return m.doGetStream(songID)
}

func (m *mockProvider) ListPlaylists() ([]*Playlist, error) {
	return m.doListPlaylists()
}

func (m *mockProvider) ListPlaylistEntries() ([]*PlaylistEntry, error) {
	return m.ListPlaylistEntries()
}

func (m *mockProvider) GetProvider() MusicProvider {
	return m.doGetProvider()
}
