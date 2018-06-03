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
