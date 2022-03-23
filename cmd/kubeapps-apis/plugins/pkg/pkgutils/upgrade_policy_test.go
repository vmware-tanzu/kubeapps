// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package pkgutils

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestVersionConstraintWithUpgradePolicy(t *testing.T) {
	tests := []struct {
		name          string
		version       string
		upgradePolicy UpgradePolicy
		expected      string
	}{
		{"get constraints with upgradePolicy 'major'", "1.2.3", UpgradePolicyMajor, ">=1.2.3"},
		{"get constraints with upgradePolicy 'minor'", "1.2.3", UpgradePolicyMinor, ">=1.2.3 <2.0.0"},
		{"get constraints with upgradePolicy 'patch'", "1.2.3", UpgradePolicyPatch, ">=1.2.3 <1.3.0"},
		{"get constraints with upgradePolicy 'none'", "1.2.3", UpgradePolicyNone, "1.2.3"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values, err := VersionConstraintWithUpgradePolicy(tt.version, tt.upgradePolicy)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !cmp.Equal(tt.expected, values) {
				t.Errorf("mismatch in '%s': %s", tt.name, cmp.Diff(tt.expected, values))
			}
		})
	}
}
