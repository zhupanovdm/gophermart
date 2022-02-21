package user

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/zhupanovdm/gophermart/pkg/hash"
)

func TestCredentials_HashPassword(t *testing.T) {
	tests := []struct {
		name         string
		cred         Credentials
		hash         hash.StringFunc
		wantPassword string
	}{
		{
			name: "Basic test",
			cred: Credentials{
				Login:    "Vasily@Pupkin.com",
				Password: "123",
			},
			hash: func(s string) string {
				return s + "321"
			},
			wantPassword: "123321",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.cred.HashPassword(tt.hash)
			assert.Equal(t, tt.wantPassword, tt.cred.Password)
		})
	}
}
