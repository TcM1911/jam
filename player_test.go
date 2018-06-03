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

package jamsonic

import (
	"errors"
	"io"
	"io/ioutil"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

const (
	track1Content = "Song content 1"
	track2Content = "Song content 2"
	track3Content = "Song content 3"
	track4Content = "Song content 4"
)

var tracks = []*Track{
	&Track{ID: "1"},
	&Track{ID: "2"},
	&Track{ID: "3"},
	&Track{ID: "4"},
}

func init() {
	// Set to 0 since it's not needed for testing.
	BufferingWait = time.Duration(0)
}

func TestPlayControl(t *testing.T) {
	assert := assert.New(t)
	player, _, _, _ := getPlayer()
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
		// Catch error
		var err error
		go func() {
			err = <-p.Error
		}()
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

		assert.NoError(err)
		p.Close()
	})

	t.Run("Playing", func(t *testing.T) {
		p, finished, provider, handler := getPlayer()
		// Catch error
		var err error
		go func() {
			err = <-p.Error
		}()
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

		assert.NoError(err)
		p.Close()
	})

	t.Run("Pausing", func(t *testing.T) {
		p, _, provider, handler := getPlayer()
		// Catch error
		var err error
		go func() {
			err = <-p.Error
		}()
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

		// Pause during the first track.
		p.Pause()
		time.Sleep(time.Millisecond * 200)
		assert.Equal(Paused, p.GetCurrentState(), "State should be paused.")
		calledMu.RLock()
		assert.Equal(1, handler.calledPlay, "Wrong number of calls")
		assert.Equal(2, handler.calledPause, "Wrong number of calls")
		assert.Equal(1, handler.calledContrinue, "Wrong number of calls")
		calledMu.RUnlock()
		assert.Equal(tracks[0], p.CurrentTrack(), "First track not playing")
		assert.Equal(tracks[1], p.NextTrack(), "2nd track should be marked as next")

		// Continue with a second call to pause.
		p.Pause()
		time.Sleep(time.Millisecond * 200)
		assert.Equal(Playing, p.GetCurrentState(), "State should be paused.")
		calledMu.RLock()
		assert.Equal(1, handler.calledPlay, "Wrong number of calls")
		assert.Equal(2, handler.calledPause, "Wrong number of calls")
		assert.Equal(2, handler.calledContrinue, "Wrong number of calls")
		calledMu.RUnlock()
		assert.Equal(tracks[0], p.CurrentTrack(), "First track not playing")
		assert.Equal(tracks[1], p.NextTrack(), "2nd track should be marked as next")

		assert.NoError(err)
		p.Close()
	})

	t.Run("Song skipping", func(t *testing.T) {
		p, _, provider, handler := getPlayer()
		// Catch error
		var err error
		go func() {
			err = <-p.Error
		}()
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

		assert.NoError(err)
		p.Close()
	})

	t.Run("Song reverse skipping", func(t *testing.T) {
		p, _, provider, handler := getPlayer()
		// Catch error
		var err error
		go func() {
			err = <-p.Error
		}()
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

		assert.NoError(err)
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
		p = NewPlayer(DefaultLogger(), provider, handler, callback, 1000)
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
		p := NewPlayer(DefaultLogger(), provider, handler, nil, 0)
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
		p := NewPlayer(DefaultLogger(), provider, handler, nil, 0)
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

func TestStreamBuffering(t *testing.T) {
	assert := assert.New(t)

	finishedChan := make(chan struct{})
	provider := &mockProvider{
		doGetStream: func(id string) (io.ReadCloser, error) {
			return &recorder{streamID: id}, nil
		},
	}

	t.Run("Can_read", func(t *testing.T) {
		var trackStreamMu sync.Mutex
		var trackStream io.Reader
		handler := &mockStreaHandler{
			doFinished: func() <-chan struct{} { return finishedChan },
			doPlay: func(r io.Reader) error {
				trackStreamMu.Lock()
				defer trackStreamMu.Unlock()
				trackStream = r
				return nil
			},
			doStop:     func() {},
			doPause:    func() {},
			doContinue: func() {},
		}
		p := NewPlayer(DefaultLogger(), provider, handler, nil, 0)

		// Catch error
		var err error
		go func() {
			err = <-p.Error
		}()
		p.CreatePlayQueue(tracks)
		p.Play()
		time.Sleep(time.Millisecond * 70)
		trackStreamMu.Lock()
		content, err := ioutil.ReadAll(trackStream)
		trackStreamMu.Unlock()

		assert.NoError(err, "Should read without error")
		assert.Equal(track1Content, string(content), "Wrong track content")

		p.Close()
	})

	t.Run("Create_new_buffer_if_still_copying", func(t *testing.T) {
		var trackStreamMu sync.Mutex
		var trackStream io.Reader
		handler := &mockStreaHandler{
			doFinished: func() <-chan struct{} { return finishedChan },
			doPlay: func(r io.Reader) error {
				trackStreamMu.Lock()
				defer trackStreamMu.Unlock()
				trackStream = r
				return nil
			},
			doStop:     func() {},
			doPause:    func() {},
			doContinue: func() {},
		}
		p := NewPlayer(DefaultLogger(), provider, handler, nil, 0)

		// Catch error
		var err error
		go func() {
			err = <-p.Error
		}()
		p.CreatePlayQueue(tracks)
		p.Play()
		time.Sleep(time.Millisecond * 10)
		p.Next()
		time.Sleep(time.Millisecond * 100)
		trackStreamMu.Lock()
		content, err := ioutil.ReadAll(trackStream)
		trackStreamMu.Unlock()

		assert.NoError(err, "Should read without error")
		assert.Equal(track2Content, string(content), "Wrong track content")

		p.Close()
	})

	t.Run("Handle errors when stream copying", func(t *testing.T) {
		finishedChan := make(chan struct{})
		expectedError := errors.New("expected error")
		provider := &mockProvider{
			doGetStream: func(id string) (io.ReadCloser, error) {
				return &recorder{streamID: expectedError.Error()}, nil
			},
		}
		handler := &mockStreaHandler{
			doFinished: func() <-chan struct{} { return finishedChan },
			doPlay:     func(io.Reader) error { return nil },
			doStop:     func() {},
			doPause:    func() {},
			doContinue: func() {},
		}
		p := NewPlayer(DefaultLogger(), provider, handler, nil, 0)
		p.CreatePlayQueue(tracks)

		p.Play()
		err := <-p.Error
		assert.Equal(expectedError.Error(), err.Error(), "Returned wrong error")
		assert.Equal(tracks[0], p.CurrentTrack(), "Wrong track set as next")

		p.Close()
	})

	t.Run("Handle errors when closing stream", func(t *testing.T) {
		finishedChan := make(chan struct{})
		expectedError := errors.New("expected error")
		provider := &mockProvider{
			doGetStream: func(id string) (io.ReadCloser, error) {
				return &recorder{streamID: id, closeErr: expectedError}, nil
			},
		}
		handler := &mockStreaHandler{
			doFinished: func() <-chan struct{} { return finishedChan },
			doPlay:     func(io.Reader) error { return nil },
			doStop:     func() {},
			doPause:    func() {},
			doContinue: func() {},
		}
		p := NewPlayer(DefaultLogger(), provider, handler, nil, 0)
		p.CreatePlayQueue(tracks)

		p.Play()
		err := <-p.Error
		assert.Equal(expectedError.Error(), err.Error(), "Returned wrong error")
		assert.Equal(tracks[0], p.CurrentTrack(), "Wrong track set as next")

		p.Close()
	})
}

func TestQueue(t *testing.T) {
	assert := assert.New(t)

	t.Run("no panic if skipping past end return error", func(t *testing.T) {
		p, _, _, _ := getPlayer()
		p.CreatePlayQueue(tracks)
		p.Play()
		p.Next()
		p.Next()
		p.Next()
		p.Next()
		err := <-p.Error
		assert.Equal(ErrNoNextTrack, err, "Wrong error returned")
	})
}

func TestChangeProvider(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)
	p, _, oldProvider, _ := getPlayer()
	p.CreatePlayQueue(tracks)
	p.Play()
	require.Equal(oldProvider, p.provider, "Wrong initial provider")
	require.Equal(Playing, p.GetCurrentState(), "Should be playing")

	expected := &mockProvider{}
	p.UpdateProvider(expected)

	assert.Equal(expected, p.provider, "New provider not set")
	assert.Equal(Stopped, p.GetCurrentState(), "Should stop any track playing")
}

func TestErrorHandle(t *testing.T) {
	p, _, _, handler := getPlayer()
	errChan := handler.errChan
	p.CreatePlayQueue(tracks)
	p.Play()

	// Should not block on error.
	errChan <- errors.New("test error")
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
		errChan:    make(chan error),
	}
	return NewPlayer(DefaultLogger(), provider, handler, nil, 0), finishedChan, provider, handler
}

type recorder struct {
	streamID string
	read1    bool
	read2    bool
	read3    bool
	read4    bool
	closeErr error
}

func (r *recorder) Read(b []byte) (int, error) {
	// Delay read so we can simulate a read from a socket.
	time.Sleep(time.Millisecond * 50)
	var reader *strings.Reader
	if r.streamID == "1" {
		if r.read1 {
			return 0, io.EOF
		}
		r.read1 = true
		reader = strings.NewReader(track1Content)
	}
	if r.streamID == "2" {
		if r.read2 {
			return 0, io.EOF
		}
		r.read2 = true
		reader = strings.NewReader(track2Content)
	}
	if r.streamID == "3" {
		if r.read3 {
			return 0, io.EOF
		}
		r.read3 = true
		reader = strings.NewReader(track3Content)
	}
	if r.streamID == "4" {
		if r.read4 {
			return 0, io.EOF
		}
		r.read4 = true
		reader = strings.NewReader(track4Content)
	}
	if r.streamID == "expected error" {
		return 0, errors.New("expected error")
	}
	return reader.Read(b)
}

func (r *recorder) Close() error {
	if r.closeErr != nil {
		return r.closeErr
	}
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
	errChan         chan error
}

var calledMu sync.RWMutex

func (m *mockStreaHandler) Finished() <-chan struct{} {
	return m.doFinished()
}

func (m *mockStreaHandler) Play(r io.Reader) error {
	calledMu.Lock()
	defer calledMu.Unlock()
	m.calledPlay++
	return m.doPlay(r)
}

func (m *mockStreaHandler) Stop() {
	calledMu.Lock()
	defer calledMu.Unlock()
	m.calledStopped++
	m.doStop()
}

func (m *mockStreaHandler) Pause() {
	calledMu.Lock()
	defer calledMu.Unlock()
	m.calledPause++
	m.doPause()
}

func (m *mockStreaHandler) Continue() {
	calledMu.Lock()
	defer calledMu.Unlock()
	m.calledContrinue++
	m.doContinue()
}

func (m *mockStreaHandler) Errors() <-chan error {
	return m.errChan
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
