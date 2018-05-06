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

package tui

import (
	"encoding/json"

	"github.com/TcM1911/jamsonic/subsonic"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

type configPage struct {
	name  string
	panel tview.Primitive
}

const (
	strSave           = "Save"
	strCancel         = "Cancel"
	strHost           = "Host"
	strUsername       = "Username"
	strPassword       = "Password"
	strBlank          = ""
	passwordMask      = '*'
	strDefaultHostStr = "https://"
	fieldWidth        = 0
)

var settingsPages *tview.Pages

func (tui *TUI) createSettingsPage() *tview.Flex {
	configPages := []*configPage{
		&configPage{name: "*sonic", panel: sonicForm(tui)},
	}
	settingsPages = tview.NewPages()
	configList := createConfigList(configPages)
	tui.settingsList = configList
	settingsPages.SetBorder(true).
		SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyTab {
				tui.app.SetFocus(configList)
				return nil
			}
			return event
		})
	configList.SetCurrentItem(0).
		SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyTab {
				tui.app.SetFocus(settingsPages)
				return nil
			}
			return event
		})
	return tview.NewFlex().
		AddItem(configList, 15, 1, true).
		AddItem(settingsPages, 0, 1, false)
}

func createConfigList(configPages []*configPage) *tview.List {
	list := tview.NewList().ShowSecondaryText(false)
	list.SetChangedFunc(func(_ int, entry string, _ string, _ rune) {
		settingsPages.SwitchToPage(entry)
	})
	for _, p := range configPages {
		list.AddItem(p.name, "", 0, nil)
		settingsPages.AddPage(p.name, p.panel, true, false)
	}
	list.SetBorder(true)
	return list
}

// newSettingsForm creates a form for the settings page.
// This should be used to create the initial form so all forms have the same look.
func newSettingsForm() *tview.Form {
	form := tview.NewForm()
	form.SetFieldBackgroundColor(tcell.ColorBlack)
	form.SetFieldTextColor(tcell.ColorWhite)
	form.SetLabelColor(tcell.ColorWhite)
	form.SetButtonTextColor(tcell.ColorBlack)
	form.SetButtonBackgroundColor(tcell.ColorWhiteSmoke)
	return form
}

// sonicForm is the form for changing *sonic credentials.
func sonicForm(tui *TUI) *tview.Form {
	host, username, password := strDefaultHostStr, strBlank, strBlank
	credBuf, err := tui.db.GetCredentials(subsonic.CredentialKey)
	if err == nil {
		var creds subsonic.Credentials
		err := json.Unmarshal(credBuf, &creds)
		if err == nil {
			host, username = creds.Host, creds.Username
		}
	}
	form := newSettingsForm()
	form.AddInputField(strHost, host, fieldWidth, nil, nil).
		AddInputField(strUsername, username, fieldWidth, nil, nil).
		AddPasswordField(strPassword, password, fieldWidth, passwordMask, nil).
		AddButton(strSave, func() {
			h := form.GetFormItemByLabel(strHost).(*tview.InputField).GetText()
			u := form.GetFormItemByLabel(strUsername).(*tview.InputField).GetText()
			p := form.GetFormItemByLabel(strPassword).(*tview.InputField).GetText()
			c, err := subsonic.Login(u, p, h)
			if err != nil {
				tui.player.Error <- err
				return
			}
			buf, err := json.Marshal(&c.Credentials)
			if err != nil {
				tui.player.Error <- err
				return
			}
			err = tui.db.SaveCredentials(subsonic.CredentialKey, buf)
			if err != nil {
				tui.player.Error <- err
				return
			}
			tui.player.UpdateProvider(c)
			tui.app.SetFocus(tui.settingsList)
		}).
		AddButton(strCancel, func() {
			tui.app.SetFocus(tui.settingsList)
		})
	return form
}
