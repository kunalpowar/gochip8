package emulator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCicularShiftLeftUint8(t *testing.T) {
	tests := []struct {
		desc     string
		input    uint8
		shiftBy  int
		expected uint8
	}{
		{
			desc:     "shift ones by 2",
			input:    0b00000011,
			expected: 0b00001100,
			shiftBy:  2,
		},
		{
			desc:     "shift ones by 2 with leading zero",
			input:    0b00000110,
			expected: 0b00011000,
			shiftBy:  2,
		},
		{
			desc:     "shift ones by 8 should not change anything",
			input:    0b00000110,
			expected: 0b00000110,
			shiftBy:  8,
		},
		{
			desc:     "shift ones by 9 should shift by 1",
			input:    0b00000110,
			expected: 0b00001100,
			shiftBy:  9,
		},
		{
			desc:     "shift ones by 18 should shift by 2",
			input:    0b00000110,
			expected: 0b00011000,
			shiftBy:  18,
		},
		{
			desc:     "shift ones by 2 with cicular",
			input:    0b11100000,
			expected: 0b10000011,
			shiftBy:  2,
		},
		{
			desc:     "shift mix by 2",
			input:    0b00000101,
			expected: 0b00010100,
			shiftBy:  2,
		},
		{
			desc:     "shift mix by 2 with leading zero",
			input:    0b00001010,
			expected: 0b00101000,
			shiftBy:  2,
		},
		{
			desc:     "shift mix by 8 should not change anything",
			input:    0b00001010,
			expected: 0b00001010,
			shiftBy:  8,
		},
		{
			desc:     "shift mix by 9 should shift by 1",
			input:    0b00001010,
			expected: 0b00010100,
			shiftBy:  9,
		},
		{
			desc:     "shift mix by 18 should shift by 2",
			input:    0b00001010,
			expected: 0b00101000,
			shiftBy:  18,
		},
		{
			desc:     "shift mix by 2 with cicular",
			input:    0b10010000,
			expected: 0b01000010,
			shiftBy:  2,
		},
		{
			desc:     "shift mix by 3 with cicular",
			input:    0b10010001,
			expected: 0b10001100,
			shiftBy:  3,
		},
	}

	for _, tt := range tests {
		t.Logf("case: %s", tt.desc)
		actual := cicularShiftLeftUint8(tt.input, tt.shiftBy)
		assert.Equal(t, tt.expected, actual, "input:    %08b\nexpected: %08b\nactual:   %08b", tt.input, tt.expected, actual)
	}
}
