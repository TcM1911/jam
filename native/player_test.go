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
		<-handler.FinishedChan
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
		<-handler.FinishedChan
		handler.Play(bytes.NewBuffer([]byte(expectedContent2)))
		<-handler.FinishedChan
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
					time.Sleep(time.Microsecond * 10)
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
		<-handler.FinishedChan
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
					time.Sleep(time.Microsecond * 10)
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
		time.Sleep(time.Microsecond * 10)
		go handler.Stop()
		time.Sleep(time.Microsecond * 10)
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
					time.Sleep(time.Microsecond * 10)
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
		time.Sleep(time.Microsecond * 10)
		go handler.Play(bytes.NewBuffer(expectedContent2))
		<-handler.FinishedChan
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
					time.Sleep(time.Microsecond * 10)
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
					time.Sleep(time.Microsecond * 10)
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
		time.Sleep(time.Microsecond * 10)
		<-handler.FinishedChan
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
		err := <-handler.ErrChan
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
		err := <-handler.ErrChan
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
