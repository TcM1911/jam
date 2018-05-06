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
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/TcM1911/jamsonic"
	"github.com/stretchr/testify/assert"
)

func TestHost(t *testing.T) {
	expectedString := "expectedString"
	c := &Client{Credentials: Credentials{Host: expectedString}}
	assert.Equal(t, expectedString, c.Host(), "Wrong value returned.")
}

func TestNewClient(t *testing.T) {
	assert := assert.New(t)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var data apiData
		if strings.Contains(r.RequestURI, "baduser") {
			data = apiData{Response: apiResponse{Status: "error"}}
		} else {
			data = apiData{Response: apiResponse{Status: "ok"}}
		}
		buf, _ := json.Marshal(&data)
		w.Write(buf)
	}))
	expectedCreds := Credentials{Host: ts.URL, Salt: "salt", Token: "token", Username: "un"}
	expectedClient := Client{Credentials: expectedCreds}

	// Tests
	t.Run("stored_credentials", func(t *testing.T) {
		buf, _ := json.Marshal(expectedCreds)
		s := &mockCredstore{
			doGetCreds: func(k []byte) ([]byte, error) {
				return buf, nil
			},
		}
		actual, err := New(s, jamsonic.DefaultCredentialRequest)
		assert.NoError(err, "Should create without an error")
		assert.Equal(expectedClient, *actual, "Wrong client data returned.")
	})
	t.Run("new_credentials", func(t *testing.T) {
		reqer := &mockRequester{doServer: ts.URL, doUsername: "username", doPassword: "password"}
		s := &mockCredstore{
			doGetCreds: func(k []byte) ([]byte, error) {
				return nil, jamsonic.ErrNoCredentialsStored
			},
			doSaveCreds: func(k []byte, v []byte) error {
				return nil
			},
		}
		_, err := New(s, reqer)
		assert.NoError(err, "Should create without an error")
	})
	t.Run("wrong_credentials", func(t *testing.T) {
		reqer := &mockRequester{doServer: ts.URL, doUsername: "baduser", doPassword: "password"}
		s := &mockCredstore{
			doGetCreds: func(k []byte) ([]byte, error) {
				return nil, jamsonic.ErrNoCredentialsStored
			},
		}
		_, err := New(s, reqer)
		assert.Equal(ErrAuthenticationFailed, err, "Should return an error if auth fails")
	})
	t.Run("error_from_GetCredentials", func(t *testing.T) {
		expectedErr := errors.New("expected")
		s := &mockCredstore{
			doGetCreds: func(k []byte) ([]byte, error) {
				return nil, expectedErr
			},
		}
		_, err := New(s, jamsonic.DefaultCredentialRequest)
		assert.Equal(expectedErr, err, "Wrong error returned")
	})
	t.Run("error_from_SaveCredentials", func(t *testing.T) {
		expectedErr := errors.New("expected")
		reqer := &mockRequester{doServer: ts.URL, doUsername: "username", doPassword: "password"}
		s := &mockCredstore{
			doGetCreds: func(k []byte) ([]byte, error) {
				return nil, jamsonic.ErrNoCredentialsStored
			},
			doSaveCreds: func(k []byte, v []byte) error {
				return expectedErr
			},
		}
		_, err := New(s, reqer)
		assert.Equal(expectedErr, err, "Wrong error returned")
	})
}

type mockRequester struct {
	doUsername string
	doPassword string
	doServer   string
}

func (r *mockRequester) GetServer() string {
	return r.doServer
}

func (r *mockRequester) GetUsername() string {
	return r.doUsername
}

func (r *mockRequester) GetPassword() string {
	return r.doPassword
}

type mockCredstore struct {
	doGetCreds  func([]byte) ([]byte, error)
	doSaveCreds func([]byte, []byte) error
}

func (m *mockCredstore) GetCredentials(k []byte) ([]byte, error) {
	return m.doGetCreds(k)
}

func (m *mockCredstore) SaveCredentials(k []byte, v []byte) error {
	return m.doSaveCreds(k, v)
}
