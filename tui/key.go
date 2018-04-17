package tui

import (
	"github.com/TcM1911/jamsonic"
	"github.com/gdamore/tcell"
)

// All pages that handles music control events should pass the event to this
// function as part of SetInputCapture.
func (tui *TUI) musicControl(event *tcell.EventKey) *tcell.EventKey {
	switch event.Rune() {
	// Music control events.
	case 'b':
		tui.player.Next()
		return nil
	case 'v':
		tui.player.Stop()
		return nil
	case 'c':
		tui.player.Pause()
		return nil
	case 'x':
		if tui.player.GetCurrentState() == jamsonic.Paused {
			tui.player.Play()
			return nil
		}
	case 'z':
		tui.player.Previous()
		return nil
	}
	return event
}

// Global key controls which is handled by the application. Should use
// Ctr combinations so typing in input boxes are not treated as events.
func (tui *TUI) globalControl(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	// Switch to next page.
	case tcell.KeyCtrlN:
		tui.currentPage = (tui.currentPage + 1) % 2
		switchPage(tui, tui.currentPage)
	}
	return event
}
