package tui

import (
	"github.com/TcM1911/jamsonic"
	"github.com/gdamore/tcell"
)

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

func (tui *TUI) globalControl(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	// Pages
	case tcell.KeyCtrlN:
		tui.currentPage = (tui.currentPage + 1) % 2
		switchPage(tui, tui.currentPage)
	}
	return event
}
