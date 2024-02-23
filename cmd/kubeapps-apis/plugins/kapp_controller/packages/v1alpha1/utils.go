// Copyright 2024 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	new "carvel.dev/vendir/pkg/vendir/versions/v1alpha1"
	old "github.com/vmware-tanzu/carvel-vendir/pkg/vendir/versions/v1alpha1"
)

// This is a temporary function to convert from the new vendir version to the old
// vendir version. This can be removed once the new vendir version is released.
// More details at https://github.com/vmware-tanzu/kubeapps/pull/7515

func toOldVendirVSS(ver *new.VersionSelectionSemver) *old.VersionSelectionSemver {
	return &old.VersionSelectionSemver{
		Constraints: ver.Constraints,
		Prereleases: &old.VersionSelectionSemverPrereleases{
			Identifiers: ver.Prereleases.Identifiers,
		},
	}
}

func toNewVendirVS(ver *old.VersionSelection) *new.VersionSelection {
	return &new.VersionSelection{
		Semver: toNewVendirVSS(ver.Semver),
	}
}

func toNewVendirVSS(ver *old.VersionSelectionSemver) *new.VersionSelectionSemver {
	return &new.VersionSelectionSemver{
		Constraints: ver.Constraints,
		Prereleases: &new.VersionSelectionSemverPrereleases{
			Identifiers: ver.Prereleases.Identifiers,
		},
	}
}

func toOldVendirVS(ver *new.VersionSelection) *old.VersionSelection {
	return &old.VersionSelection{
		Semver: toOldVendirVSS(ver.Semver),
	}
}
