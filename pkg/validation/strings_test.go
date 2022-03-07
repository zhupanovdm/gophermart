package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOnlyDigitsValidator(t *testing.T) {
	tests := []struct {
		name   string
		sample string
		want   bool
	}{
		{
			name:   "Basic test",
			sample: "12345",
			want:   true,
		},
		{
			name:   "Non digits present",
			sample: "12.345",
		},
		{
			name: "Empty string",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, OnlyDigits(tt.sample))
		})
	}
}
