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
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogger(t *testing.T) {
	assert := assert.New(t)
	infoTestLogLine := "Should log at info level"
	debugTestLogLine := "Should log debug level"
	notDebugTestLogLine := "This debug level should not be logged"
	errTestLogLine := "Should log errors"
	buf := new(bytes.Buffer)
	logger := NewLogger(buf)
	logger.InfoLog(infoTestLogLine)
	logger.DebugLog(notDebugTestLogLine)
	logger.SetLevel(DebugLevel)
	logger.DebugLog(debugTestLogLine)
	logger.ErrorLog(errTestLogLine)
	logged := buf.String()

	t.Run("info logging", func(t *testing.T) {
		assert.Contains(logged, defaultInfoPrefix)
		assert.Contains(logged, infoTestLogLine, infoTestLogLine)
	})

	t.Run("debug logging", func(t *testing.T) {
		assert.Contains(logged, defaultDebugPrefix)
		assert.Contains(logged, debugTestLogLine, debugTestLogLine)
	})

	t.Run("not debug logging", func(t *testing.T) {
		assert.NotContains(logged, notDebugTestLogLine, notDebugTestLogLine)
	})

	t.Run("error logging", func(t *testing.T) {
		assert.Contains(logged, defaultErrorPrefix)
		assert.Contains(logged, errTestLogLine, errTestLogLine)
	})
}

func TestSubLogs(t *testing.T) {
	assert := assert.New(t)
	logger, prefixes, loggers := getLoggers()

	t.Run("correct sub prefix", func(t *testing.T) {
		for i, p := range prefixes {
			expected := fmt.Sprintf("%s%s ", defaultInfoPrefix, p)
			assert.Equal(expected, loggers[i].info.Prefix())
		}
	})

	t.Run("children are added to array", func(t *testing.T) {
		logger.childrenmu.RLock()
		defer logger.childrenmu.RUnlock()
		children := logger.children
		for _, child := range loggers {
			assert.Contains(children, child)
		}
	})

	t.Run("set log level", func(t *testing.T) {
		for _, log := range loggers {
			log.levelmu.RLock()
			require.Equal(t, InfoLevel, log.level)
			log.levelmu.RUnlock()
		}

		logger.SetLevel(DebugLevel)

		for _, log := range loggers {
			log.levelmu.RLock()
			assert.Equal(DebugLevel, log.level)
			log.levelmu.RUnlock()
		}
	})
}

func TestRemoveChildren(t *testing.T) {
	assert := assert.New(t)
	t.Run("remove_first", func(t *testing.T) {
		logger, _, loggers := getLoggers()
		logger.RemoveChild(loggers[0])
		logger.childrenmu.RLock()
		defer logger.childrenmu.RUnlock()
		children := logger.children
		assert.Contains(children, loggers[1])
		assert.Contains(children, loggers[2])
		assert.NotContains(children, loggers[0])
	})

	t.Run("remove_middle", func(t *testing.T) {
		logger, _, loggers := getLoggers()
		logger.RemoveChild(loggers[1])
		logger.childrenmu.RLock()
		defer logger.childrenmu.RUnlock()
		children := logger.children
		assert.Contains(children, loggers[0])
		assert.Contains(children, loggers[2])
		assert.NotContains(children, loggers[1])
	})

	t.Run("remove_last", func(t *testing.T) {
		logger, _, loggers := getLoggers()
		logger.RemoveChild(loggers[2])
		logger.childrenmu.RLock()
		defer logger.childrenmu.RUnlock()
		children := logger.children
		assert.Contains(children, loggers[0])
		assert.Contains(children, loggers[1])
		assert.NotContains(children, loggers[2])
	})
}

func TestChangeOutput(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	t.Run("change default logger", func(t *testing.T) {
		logger := DefaultLogger()
		buf := new(bytes.Buffer)
		require.Equal(os.Stderr, logger.writer)

		logger.SetOutput(buf)
		assert.Equal(buf, logger.writer)
	})

	t.Run("change childrens' writer", func(t *testing.T) {
		logger, _, loggers := getLoggers()
		buf := new(bytes.Buffer)
		logger.SetOutput(buf)
		for _, c := range loggers {
			assert.Equal(buf, c.writer)
		}
	})
}

func getLoggers() (*Logger, []string, []*Logger) {
	prefixes := []string{"[Logger 1]", "[Logger 2]", "[Logger 3]"}
	logger := NewLogger(new(bytes.Buffer))
	return logger,
		prefixes,
		[]*Logger{
			logger.SubLogger(prefixes[0]),
			logger.SubLogger(prefixes[1]),
			logger.SubLogger(prefixes[2]),
		}
}
