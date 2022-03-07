package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_authenticationHandler_Login(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantStatus int
		wantHeader http.Header
	}{
		{
			name:       "Basic test",
			body:       `{"login":"Vasily.Pupkin@mail.ru","password":"secret"}`,
			wantStatus: http.StatusOK,
			wantHeader: http.Header{authorizationHeader: []string{tokenPrefix + SampleFakeString}},
		},
		{
			name:       "Bad request",
			body:       `{"login":"Vasily.Pupkin@mail.ru"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Bad credentials",
			body:       `{"login":"wrong","password":"wrong"}`,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "Service fail",
			body:       `{"login":"crash","password":"secret"}`,
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/login", bytes.NewBufferString(tt.body))
			resp := httptest.NewRecorder()

			(&authenticationHandler{
				Auth: NewAuthServiceMock(),
			}).Login(resp, req)

			result := resp.Result()
			//goland:noinspection GoUnhandledErrorResult
			defer result.Body.Close()

			if assert.Equal(t, tt.wantStatus, result.StatusCode, "wrong status code returned") {
				for key := range tt.wantHeader {
					assert.Equalf(t, tt.wantHeader.Get(key), result.Header.Get(key), "wrong header %s", key)
				}
			}
		})
	}
}

func Test_authenticationHandler_Register(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{
			name:       "Basic test",
			body:       `{"login":"Vasily.Pupkin@mail.ru","password":"secret"}`,
			wantStatus: http.StatusOK,
		},
		{
			name:       "Bad request",
			body:       `{"login":"Vasily.Pupkin@mail.ru"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Register existing user",
			body:       `{"login":"exists","password":"secret"}`,
			wantStatus: http.StatusConflict,
		},
		{
			name:       "Service fail",
			body:       `{"login":"crash","password":"secret"}`,
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/register", bytes.NewBufferString(tt.body))
			resp := httptest.NewRecorder()

			(&authenticationHandler{
				Auth: NewAuthServiceMock(),
			}).Register(resp, req)

			result := resp.Result()
			//goland:noinspection GoUnhandledErrorResult
			defer result.Body.Close()

			assert.Equal(t, tt.wantStatus, result.StatusCode, "wrong status code returned")
		})
	}
}
