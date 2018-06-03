// Copyright (c) 2018 Joakim Kennedy
// Copyright (c) 2016, 2017 Evgeny Badin
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
	"fmt"
	"log"

	"github.com/howeyc/gopass"
)

// CredentialRequester is an interface for getting credentials from the user.
type CredentialRequester interface {
	// GetServer should return the server url.
	GetServer() string
	// GetUsername should return the username gotten from the user.
	GetUsername() string
	// GetPassword should return the password gotten from the user.
	GetPassword() string
}

type CredentialRequest struct{}

// DefaultCredentialRequest is a default credential requester. It
// will read the info from stdin.
var DefaultCredentialRequest = &CredentialRequest{}

// GetServer asks the user to enter the server url via stdin.
func (c *CredentialRequest) GetServer() string {
	return AskForServer()
}

// GetUsername asks the user to enter the username via stdin.
func (c *CredentialRequest) GetUsername() string {
	return AskForUsername()
}

// GetPassword asks the user to enter the password via stdin.
func (c *CredentialRequest) GetPassword() string {
	pass, err := AskForPassword()
	if err != nil {
		log.Println("Error when asking for the password:", err.Error())
		return ""
	}
	return string(pass)
}

func AskForServer() string {
	var host string
	fmt.Print("Url to Subsonic host: ")
	fmt.Scanln(&host)
	return host
}

func AskForUsername() string {
	var email string
	fmt.Print("Email/Username: ")
	fmt.Scanln(&email)
	return email
}

func AskForPassword() ([]byte, error) {
	var password []byte
	fmt.Print("Password: ")
	password, err := gopass.GetPasswdMasked()
	if err != nil {
		return nil, err
	}
	return password, nil
}
