package handlers

import (
	"github.com/stretchr/testify/assert"
	"github.com/zhupanovdm/gophermart/model/user"
	"net/http"
	"testing"
)

func TestTokenBearerHeader(t *testing.T) {
	tests := []struct {
		name       string
		sample     user.Token
		skipSample bool
		want       user.Token
	}{
		{
			name:   "Basic test",
			sample: user.Token("1234567890"),
			want:   user.Token("1234567890"),
		},
		{
			name:   "Void token",
			sample: user.VoidToken,
			want:   user.VoidToken,
		},
		{
			name:       "Has no token",
			skipSample: true,
			want:       user.VoidToken,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var header = make(http.Header)
			bearer := TokenBearerHeader(header)
			if !tt.skipSample {
				bearer.Set(tt.sample)
			}
			assert.Equal(t, tt.want, bearer.Get())
		})
	}
}
