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
	// CredentialKey is the database key for the credential structure.
	CredentialKey = []byte("subsonicCredsKey")
	// ErrAuthenticationFailed is returned if authentication with the server failed.
	ErrAuthenticationFailed = errors.New("authentication failed")
)

// Client is the Subsonic client which talks to the Subsonic server.
type Client struct {
	Credentials
	lib    []*jamsonic.Artist
	logger *jamsonic.Logger
}

// Credentials is structure for subsonic credentials.
type Credentials struct {
	// Username for the account.
	Username string
	// Authentication token.
	Token string
	// Salt for the authentication token.
	Salt string
	// Host is the URL to the server.
	Host string
}

// New returns a new instance of the Subsonic client. If credentials are stored
// in the storage, the client will use the stored credentials. Otherwise, it will
// request the user to enter the server url, username, and password on the command
// line.
func New(db jamsonic.AuthStore, credRequester jamsonic.CredentialRequester, logger *jamsonic.Logger) (*Client, error) {
	logger.DebugLog("Getting stored credentials from the database.")
	credBuf, err := db.GetCredentials(CredentialKey)
	if err == jamsonic.ErrNoCredentialsStored {
		logger.DebugLog("No stored credentials")
		host := credRequester.GetServer()
		username := credRequester.GetUsername()
		password := credRequester.GetPassword()
		logger.InfoLog("Testing credentials.")
		client, err := Login(username, password, host)
		if err != nil {
			logger.DebugLog("Testing credentials failed.")
			return nil, err
		}
		client.logger = logger
		logger.InfoLog("Credentials valid.")
		creds := client.Credentials
		buf, err := json.Marshal(creds)
		if err != nil {
			return nil, err
		}
		logger.DebugLog("Saving the credentials to the database.")
		err = db.SaveCredentials(CredentialKey, buf)
		if err != nil {
			return nil, err
		}
		return client, nil
	}
	if err != nil {
		return nil, err
	}
	var creds Credentials
	err = json.Unmarshal(credBuf, &creds)
	if err != nil {
		return nil, err
	}
	client := Client{Credentials: creds, logger: logger}
	return &client, nil
}

func Login(username, password, host string) (*Client, error) {
	hasher := md5.New()
	randomUUID, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}
	salt := randomUUID.String()
	hasher.Write([]byte(password + salt))
	c := &Client{
		Credentials: Credentials{
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
		return nil, ErrAuthenticationFailed
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
		c.Credentials.Host,
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
	return c.Credentials.Host
}
