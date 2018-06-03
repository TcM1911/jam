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
