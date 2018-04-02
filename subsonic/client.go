/*
Copyright (c) 2018 Joakim Kennedy

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package subsonic

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/TcM1911/jamsonic"
	"github.com/satori/go.uuid"
)

const (
	apiVersion = "1.13.0"
	clientName = "Jamsonic"
)

var (
	credentialKey = []byte("subsonicCredsKey")
)

// Client is the Subsonic client which talks to the Subsonic server.
type Client struct {
	credentials
	lib []*jamsonic.Artist
}

type credentials struct {
	Username string
	Token    string
	Salt     string
	Host     string
}

// New returns a new instance of the Subsonic client. If credentials are stored
// in the storage, the client will use the stored credentials. Otherwise, it will
// request the user to enter the server url, username, and password on the command
// line.
func New(db jamsonic.AuthStore) (*Client, error) {
	credBuf, err := db.GetCredentials(credentialKey)
	if err == jamsonic.ErrNoCredentialsStored || credBuf == nil {
		host := jamsonic.AskForServer()
		username := jamsonic.AskForUsername()
		password, err := jamsonic.AskForPassword()
		if err != nil {
			return nil, err
		}
		client, err := login(username, string(password), host)
		if err != nil {
			return nil, err
		}
		creds := client.credentials
		buf, err := json.Marshal(creds)
		if err != nil {
			return nil, err
		}
		err = db.SaveCredentials(credentialKey, buf)
		if err != nil {
			return nil, err
		}
		return client, nil
	}
	if err != nil {
		return nil, err
	}
	var creds credentials
	err = json.Unmarshal(credBuf, &creds)
	if err != nil {
		return nil, err
	}
	client := Client{credentials: creds}
	return &client, nil
}

func login(username, password, host string) (*Client, error) {
	hasher := md5.New()
	randomUUID, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}
	salt := randomUUID.String()
	hasher.Write([]byte(password + salt))
	c := &Client{
		credentials: credentials{
			Username: username,
			Host:     host,
			Token:    hex.EncodeToString(hasher.Sum(nil)),
			Salt:     salt,
		},
	}
	url := c.makeRequestURL("ping")
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	var data apiData
	json.NewDecoder(res.Body).Decode(&data)
	if data.Response.Status != "ok" {
		return nil, errors.New("authentication failed")
	}
	return c, nil
}

func sendRequest(url string) (*apiResponse, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	var data apiData
	err = json.NewDecoder(res.Body).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data.Response, nil
}

func (c *Client) makeRequestURL(method string) string {
	return fmt.Sprintf("%s/rest/%s.view?u=%s&t=%s&s=%s&v=%s&c=%s&f=json",
		c.credentials.Host,
		method,
		c.Username,
		c.Token,
		c.Salt,
		apiVersion,
		clientName,
	)
}

// Host returns the server host.
func (c *Client) Host() string {
	return c.credentials.Host
}
