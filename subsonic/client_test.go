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
	expectedLogger := jamsonic.DefaultLogger()
	expectedClient := Client{Credentials: expectedCreds, logger: expectedLogger}

	// Tests
	t.Run("stored_credentials", func(t *testing.T) {
		buf, _ := json.Marshal(expectedCreds)
		s := &mockCredstore{
			doGetCreds: func(k []byte) ([]byte, error) {
				return buf, nil
			},
		}
		actual, err := New(s, jamsonic.DefaultCredentialRequest, expectedLogger)
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
		_, err := New(s, reqer, jamsonic.DefaultLogger())
		assert.NoError(err, "Should create without an error")
	})
	t.Run("wrong_credentials", func(t *testing.T) {
		reqer := &mockRequester{doServer: ts.URL, doUsername: "baduser", doPassword: "password"}
		s := &mockCredstore{
			doGetCreds: func(k []byte) ([]byte, error) {
				return nil, jamsonic.ErrNoCredentialsStored
			},
		}
		_, err := New(s, reqer, jamsonic.DefaultLogger())
		assert.Equal(ErrAuthenticationFailed, err, "Should return an error if auth fails")
	})
	t.Run("error_from_GetCredentials", func(t *testing.T) {
		expectedErr := errors.New("expected")
		s := &mockCredstore{
			doGetCreds: func(k []byte) ([]byte, error) {
				return nil, expectedErr
			},
		}
		_, err := New(s, jamsonic.DefaultCredentialRequest, jamsonic.DefaultLogger())
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
		_, err := New(s, reqer, jamsonic.DefaultLogger())
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
