package tui

import (
	"github.com/TcM1911/jamsonic"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

// All pages that handles music control events should pass the event to this
// function as part of SetInputCapture.
func (tui *TUI) musicControl(event *tcell.EventKey) *tcell.EventKey {
	if event == nil {
		return nil
	}
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
	if event == nil {
		return nil
	}
	switch event.Key() {
	// Switch to next page.
	case tcell.KeyCtrlN:
		tui.currentPage = (tui.currentPage + 1) % 2
		switchPage(tui, tui.currentPage)
	case tcell.KeyEsc:
		tui.player.Stop()
		tui.player.Close()
		tui.app.Stop()
	}
	return event
}

// vimBindings provides similar navigations to Vim.
func (tui *TUI) vimBindings(event *tcell.EventKey) *tcell.EventKey {
	if event == nil {
		return nil
	}
	current := tui.app.GetFocus()
	list, ok := current.(*tview.List)
	if !ok {
		return event
	}
	switch event.Rune() {
	case 'j':
		return tcell.NewEventKey(tcell.KeyDown, 'j', tcell.ModNone)
	case 'k':
		return tcell.NewEventKey(tcell.KeyUp, 'k', tcell.ModNone)
	case 'G':
		list.SetCurrentItem(list.GetItemCount() - 1)
		return nil
	case 'g':
		list.SetCurrentItem(0)
		return nil
	}
	return event
}
