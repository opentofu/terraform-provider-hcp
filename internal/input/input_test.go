// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package input_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-hcp/internal/input"
	"github.com/stretchr/testify/require"
)

func Test_NormalizeVersion(t *testing.T) {
	tcs := map[string]struct {
		expected string
		input    string
	}{
		"with a prefixed v": {
			input:    "v1.9.0",
			expected: "v1.9.0",
		},
		"without a prefixed v": {
			input:    "1.9.0",
			expected: "v1.9.0",
		},
	}

	for n, tc := range tcs {
		t.Run(n, func(t *testing.T) {
			r := require.New(t)

			result := input.NormalizeVersion(tc.input)
			r.Equal(tc.expected, result)
		})
	}
}
