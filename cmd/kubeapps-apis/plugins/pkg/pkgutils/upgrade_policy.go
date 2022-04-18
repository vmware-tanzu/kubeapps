// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package pkgutils

import (
	"fmt"

	"github.com/Masterminds/semver/v3"
)

// Create a UpgradePolicy enum-alike
type UpgradePolicy int

const (
	UpgradePolicyNone UpgradePolicy = iota
	UpgradePolicyPatch
	UpgradePolicyMinor
	UpgradePolicyMajor
)

func UpgradePolicyFromString(policyStr string) (UpgradePolicy, error) {
	switch policyStr {
	case "", "none":
		return UpgradePolicyNone, nil
	case "major":
		return UpgradePolicyMajor, nil
	case "minor":
		return UpgradePolicyMinor, nil
	case "patch":
		return UpgradePolicyPatch, nil
	default:
		return UpgradePolicyNone, fmt.Errorf("unsupported upgrade policy: [%s]", policyStr)
	}
}

func (s UpgradePolicy) String() string {
	switch s {
	case UpgradePolicyMajor:
		return "major"
	case UpgradePolicyMinor:
		return "minor"
	case UpgradePolicyPatch:
		return "patch"
	default:
		return "none"
	}
}

func VersionConstraintWithUpgradePolicy(pkgVersion string, policy UpgradePolicy) (string, error) {
	version, err := semver.NewVersion(pkgVersion)
	if err != nil {
		// if the pkgVersion looks like this "< 5", .NewVersion() will return an error
		// "Invalid Semantic Version", but .NewConstraint() will work fine
		if _, err2 := semver.NewConstraint(pkgVersion); err2 == nil {
			// this is a constraint-based semver expression
			// per https://github.com/vmware-tanzu/kubeapps/issues/4424#issuecomment-1068776980
			// return as is, ignoring the upgrade policy
			return pkgVersion, nil
		} else {
			// neither a version nor a constraint, raise an error
			return "", err
		}
	}

	// Example: 1.2.3
	switch policy {
	case UpgradePolicyMajor:
		// >= 1.2.3 (1.2.4 and 1.3.0 and 2.0.0 are valid)
		return fmt.Sprintf(">=%s", version.String()), nil
	case UpgradePolicyMinor:
		// >= 1.2.3 <2.0.0 (1.2.4 and 1.3.0 are valid, but 2.0.0 is not)
		return fmt.Sprintf(">=%s <%s", version.String(), version.IncMajor().String()), nil
	case UpgradePolicyPatch:
		// >= 1.2.3 <2.0.0 (1.2.4 is valid, but 1.3.0 and 2.0.0 are not)
		return fmt.Sprintf(">=%s <%s", version.String(), version.IncMinor().String()), nil
	case UpgradePolicyNone:
		// 1.2.3 (only 1.2.3 is valid)
		return version.String(), nil
	}
	// Default: 1.2.3 (only 1.2.3 is valid)
	return version.String(), nil
}
