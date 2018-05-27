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
	"fmt"
	"io"
	"log"
	"os"
	"sync"
)

// LogLevel is used to determine which logs are written to the writer.
// The logger will write log types of the set level or lower.
// For example if the level is set at InfoLevel, info and error logs will
// be written to the writer but debug logs will not.
type LogLevel int8

const (
	// ErrorLevel will only log errors.
	ErrorLevel LogLevel = iota
	// InfoLevel will log error and info logs.
	InfoLevel
	// DebugLevel is the highest level of logging.
	DebugLevel
)

const (
	defaultInfoPrefix  = "[Info] "
	defaultDebugPrefix = "[Debug] "
	defaultErrorPrefix = "[Error] "
)

// NewLogger creates a new Logger instance that will log to the given io.Writer.
func NewLogger(w io.Writer) *Logger {
	return newLogger(w, defaultInfoPrefix, defaultDebugPrefix, defaultErrorPrefix)
}

// DefaultLogger returns a Logger instance that logs to STDERR.
func DefaultLogger() *Logger {
	return NewLogger(os.Stderr)
}

func newLogger(w io.Writer, infoPrefix, debugPrefix, errorPrefix string) *Logger {
	return &Logger{
		writer: w,
		info:   log.New(w, infoPrefix, log.LstdFlags),
		debug:  log.New(w, debugPrefix, log.LstdFlags),
		err:    log.New(w, errorPrefix, log.LstdFlags),
		level:  InfoLevel,
	}
}

// Logger instance can be used to log messages to the user.
// Each parent logger keep tracks of its children logs.
type Logger struct {
	level      LogLevel
	levelmu    sync.RWMutex
	writer     io.Writer
	info       *log.Logger
	debug      *log.Logger
	err        *log.Logger
	children   []*Logger
	childrenmu sync.RWMutex
}

// ErrorLog writes an error the log.
func (l *Logger) ErrorLog(logLine string) {
	l.err.Println(logLine)
}

// InfoLog writes an info message to the log.
func (l *Logger) InfoLog(logLine string) {
	l.info.Println(logLine)
}

// DebugLog writes a debug message to the log.
func (l *Logger) DebugLog(logLine string) {
	l.levelmu.RLock()
	defer l.levelmu.RUnlock()
	if l.level >= DebugLevel {
		l.debug.Println(logLine)
	}
}

// SetLevel changes the logging level to the new level.
func (l *Logger) SetLevel(level LogLevel) {
	l.levelmu.Lock()
	defer l.levelmu.Unlock()
	l.level = level
	l.childrenmu.Lock()
	defer l.childrenmu.Unlock()
	for _, c := range l.children {
		c.SetLevel(level)
	}
}

// SetOutput changes the writer used by the logger.
func (l *Logger) SetOutput(w io.Writer) {
	l.writer = w
	l.info.SetOutput(w)
	l.debug.SetOutput(w)
	l.err.SetOutput(w)

	l.childrenmu.Lock()
	defer l.childrenmu.Unlock()
	for _, c := range l.children {
		c.SetOutput(w)
	}
}

// SubLogger creates a new child log. The prefix for the child will be
// The given prefix appended to the parents prefix.
func (l *Logger) SubLogger(prefix string) *Logger {
	l.levelmu.RLock()
	level := l.level
	l.levelmu.RUnlock()
	child := newLogger(
		l.writer,
		fmt.Sprintf("%s%s ", l.info.Prefix(), prefix),
		fmt.Sprintf("%s%s ", l.debug.Prefix(), prefix),
		fmt.Sprintf("%s%s ", l.err.Prefix(), prefix))
	l.childrenmu.Lock()
	l.children = append(l.children, child)
	l.childrenmu.Unlock()
	child.SetLevel(level)
	return child
}

// RemoveChild tells the Logger to stop tracking the child logger.
func (l *Logger) RemoveChild(child *Logger) {
	l.childrenmu.Lock()
	defer l.childrenmu.Unlock()
	numChildren := len(l.children)
	for i, c := range l.children {
		if c == child {
			if i != numChildren-1 {
				l.children = append(l.children[:i], l.children[i+1:]...)
			} else {
				l.children = l.children[:i]
			}
			return
		}
	}
}
