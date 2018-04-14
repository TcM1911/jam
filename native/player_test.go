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
	"bytes"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPlayer(t *testing.T) {
	assert := assert.New(t)
	// Ignore decoding the stream
	newDecoder = func(r *io.Reader) io.Reader { return *r }
	inputBufferSize = 1

	t.Run("play full stream", func(t *testing.T) {
		expectedContent := []byte{0x1, 0x2, 0x3, 0x4}
		recorder := new(bytes.Buffer)
		outStream := &mockOutStream{
			doWrite: func(b []byte) (int, error) { return recorder.Write(b) },
			doClose: func() error { return nil },
		}
		makeOutputStream = func() (OutputStream, error) { return outStream, nil }
		handler := New()
		handler.Play(bytes.NewBuffer([]byte(expectedContent)))
		<-handler.Finished()
		handler.Stop()
		data := recorder.Bytes()
		assert.Equal(expectedContent, data, "Incorrect content returned.")
	})

	t.Run("play two full stream", func(t *testing.T) {
		expectedContent1 := []byte{0x1, 0x2, 0x3, 0x4}
		expectedContent2 := []byte{0x5, 0x6, 0x7, 0x8}
		recorder := new(bytes.Buffer)
		outStream := &mockOutStream{
			doWrite: func(b []byte) (int, error) { return recorder.Write(b) },
			doClose: func() error { return nil },
		}
		makeOutputStream = func() (OutputStream, error) { return outStream, nil }
		handler := New()
		handler.Play(bytes.NewBuffer([]byte(expectedContent1)))
		<-handler.Finished()
		handler.Play(bytes.NewBuffer([]byte(expectedContent2)))
		<-handler.Finished()
		handler.Stop()
		data := recorder.Bytes()
		expectedContent := append(expectedContent1, expectedContent2...)
		assert.Equal(expectedContent, data, "Incorrect content returned.")
	})

	t.Run("handle pause and resume", func(t *testing.T) {
		expectedContent := []byte{0x1, 0x2, 0x3, 0x4}
		wait := make(chan struct{})
		handler := New()
		recorder := new(bytes.Buffer)
		outStream := &mockOutStream{
			doWrite: func(b []byte) (int, error) {
				if b[0] == 0x2 {
					n, err := recorder.Write(b)
					go handler.Pause()
					wait <- struct{}{}
					time.Sleep(time.Millisecond * 200)
					return n, err
				}
				return recorder.Write(b)
			},
			doClose: func() error { return nil }}
		makeOutputStream = func() (OutputStream, error) { return outStream, nil }
		handler.Play(bytes.NewBuffer(expectedContent))
		<-wait
		raw1 := recorder.Bytes()
		handler.Continue()
		<-handler.Finished()
		handler.Stop()
		raw2 := recorder.Bytes()
		assert.Equal(expectedContent[:2], raw1, "Wrong content at the pause point")
		assert.Equal(expectedContent, raw2, "Incorrect content returned.")
	})
	t.Run("handle pause and stop", func(t *testing.T) {
		expectedContent := []byte{0x1, 0x2, 0x3, 0x4}
		wait := make(chan struct{})
		handler := New()
		recorder := new(bytes.Buffer)
		outStream := &mockOutStream{
			doWrite: func(b []byte) (int, error) {
				if b[0] == 0x2 {
					n, err := recorder.Write(b)
					go handler.Pause()
					wait <- struct{}{}
					time.Sleep(time.Millisecond * 200)
					<-wait
					return n, err
				}
				return recorder.Write(b)
			},
			doClose: func() error { return nil }}
		makeOutputStream = func() (OutputStream, error) { return outStream, nil }
		handler.Play(bytes.NewBuffer(expectedContent))
		<-wait
		raw1 := recorder.Bytes()
		wait <- struct{}{}
		time.Sleep(time.Millisecond * 200)
		go handler.Stop()
		time.Sleep(time.Millisecond * 200)
		raw2 := recorder.Bytes()
		assert.Equal(expectedContent[:2], raw1, "Wrong content at the pause point")
		assert.Equal(expectedContent[:2], raw2, "Incorrect content returned.")
	})
	t.Run("handle pause and play new stream", func(t *testing.T) {
		expectedContent1 := []byte{0x1, 0x2, 0x3, 0x4}
		expectedContent2 := []byte{0x5, 0x6, 0x7, 0x8}
		wait := make(chan struct{})
		handler := New()
		recorder := new(bytes.Buffer)
		outStream := &mockOutStream{
			doWrite: func(b []byte) (int, error) {
				if b[0] == 0x2 {
					n, err := recorder.Write(b)
					go handler.Pause()
					wait <- struct{}{}
					time.Sleep(time.Millisecond * 200)
					<-wait
					return n, err
				}
				return recorder.Write(b)
			},
			doClose: func() error { return nil }}
		makeOutputStream = func() (OutputStream, error) { return outStream, nil }
		handler.Play(bytes.NewBuffer(expectedContent1))
		<-wait
		raw1 := recorder.Bytes()
		wait <- struct{}{}
		time.Sleep(time.Millisecond * 200)
		go handler.Play(bytes.NewBuffer(expectedContent2))
		<-handler.Finished()
		raw2 := recorder.Bytes()
		expectedFinalContent := expectedContent1[:2]
		expectedFinalContent = append(expectedFinalContent, expectedContent2...)
		assert.Equal(expectedContent1[:2], raw1, "Wrong content at the pause point")
		assert.Equal(expectedFinalContent, raw2, "Incorrect content returned.")
	})
	t.Run("handle stop", func(t *testing.T) {
		expectedContent := []byte{0x1, 0x2, 0x3, 0x4}
		wait := make(chan struct{})
		handler := New()
		recorder := new(bytes.Buffer)
		outStream := &mockOutStream{
			doWrite: func(b []byte) (int, error) {
				if b[0] == 0x2 {
					n, err := recorder.Write(b)
					wait <- struct{}{}
					<-wait
					time.Sleep(time.Millisecond * 200)
					return n, err
				}
				return recorder.Write(b)
			},
			doClose: func() error { return nil }}
		makeOutputStream = func() (OutputStream, error) { return outStream, nil }
		handler.Play(bytes.NewBuffer(expectedContent))
		<-wait
		raw1 := recorder.Bytes()
		wait <- struct{}{}
		handler.Stop()
		raw2 := recorder.Bytes()
		assert.Equal(expectedContent[:2], raw1, "Wrong content at the pause point")
		assert.True(len(raw2) >= len(raw1), "Second snapshot too short.")
		assert.True(len(raw2) < len(expectedContent), "Second snapshot too Long.")
		assert.Equal(expectedContent[:len(raw2)], raw2, "Wrong content of second snapshot")
	})
	t.Run("handle skip track", func(t *testing.T) {
		expectedContent1 := []byte{0x1, 0x2, 0x3, 0x4}
		expectedContent2 := []byte{0x5, 0x6, 0x7, 0x8}
		wait := make(chan struct{})
		handler := New()
		recorder := new(bytes.Buffer)
		outStream := &mockOutStream{
			doWrite: func(b []byte) (int, error) {
				if b[0] == 0x2 {
					n, err := recorder.Write(b)
					wait <- struct{}{}
					<-wait
					time.Sleep(time.Millisecond * 200)
					return n, err
				}
				return recorder.Write(b)
			},
			doClose: func() error { return nil }}
		makeOutputStream = func() (OutputStream, error) { return outStream, nil }
		handler.Play(bytes.NewBuffer(expectedContent1))
		<-wait
		raw1 := recorder.Bytes()
		go handler.Play(bytes.NewBuffer(expectedContent2))
		wait <- struct{}{}
		time.Sleep(time.Millisecond * 200)
		<-handler.Finished()
		raw2 := recorder.Bytes()
		expectedFinalContent := expectedContent1[:len(raw1)]
		expectedFinalContent = append(expectedFinalContent, expectedContent2...)
		assert.Equal(expectedContent1[:2], raw1, "Wrong content at the pause point")
		assert.Equal(expectedFinalContent, raw2, "Incorrect content returned.")
	})

	t.Run("handle output stream error", func(t *testing.T) {
		content := "some content"
		expectedError := errors.New("exepected error")
		makeOutputStream = func() (OutputStream, error) { return nil, expectedError }
		steam := bytes.NewBuffer([]byte(content))
		handler := New()
		err := handler.Play(steam)
		assert.Equal(expectedError, err, "Incorrect error returned.")
	})

	t.Run("handle write error", func(t *testing.T) {
		content := "some content"
		expectedError := errors.New("exepected error")
		outStream := &mockOutStream{
			doWrite: func(b []byte) (int, error) { return 0, expectedError },
			doClose: func() error { return nil },
		}
		makeOutputStream = func() (OutputStream, error) { return outStream, nil }
		handler := New()
		handler.Play(bytes.NewBuffer([]byte(content)))
		err := <-handler.Errors()
		assert.Equal(expectedError, err, "Incorrect error returned.")
	})

	t.Run("handle read error", func(t *testing.T) {
		content := "some content"
		expectedError := errors.New("exepected error")
		recorder := new(bytes.Buffer)
		outStream := &mockOutStream{
			doWrite: func(b []byte) (int, error) { return recorder.Write(b) },
			doClose: func() error { return nil },
		}
		makeOutputStream = func() (OutputStream, error) { return outStream, nil }

		errReader := &controlledReader{err: expectedError}
		newDecoder = func(r *io.Reader) io.Reader { return errReader }

		handler := New()
		handler.Play(bytes.NewBuffer([]byte(content)))
		err := <-handler.Errors()
		assert.Equal(expectedError, err, "Incorrect error returned.")
	})
}

type mockOutStream struct {
	doWrite func([]byte) (int, error)
	doClose func() error
}

func (m *mockOutStream) Write(b []byte) (int, error) {
	return m.doWrite(b)
}

func (m *mockOutStream) CloseStream() error {
	return m.doClose()
}

type controlledReader struct {
	err error
}

func (c *controlledReader) Read(out []byte) (int, error) {
	return 0, c.err
}
