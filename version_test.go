package ma_test

import (
	"testing"

	"github.com/bzimmer/ma"
)

func TestVersion(t *testing.T) {
	tests := []harness{
		{
			name: "version",
			args: []string{"version"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			run(t, &tt, nil, ma.CommandVersion)
		})
	}
}
