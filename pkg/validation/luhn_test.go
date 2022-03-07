package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLuhn(t *testing.T) {
	tests := []struct {
		name   string
		number string
		want   bool
	}{
		{
			name:   "Correct number",
			number: "4561261212345467",
			want:   true,
		},
		{
			name:   "Wrong number",
			number: "4561261212345464",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, Luhn(tt.number))
		})
	}
}
