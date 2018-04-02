// Copyright (c) 2018 Joakim Kennedy
// Copyright (c) 2016, 2017 Evgeny Badin

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
