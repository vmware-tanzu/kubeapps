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
	old := &old.VersionSelectionSemver{
		Constraints: ver.Constraints,
		Prereleases: &old.VersionSelectionSemverPrereleases{},
	}

	if ver.Prereleases != nil {
		old.Prereleases.Identifiers = ver.Prereleases.Identifiers
	}

	return old
}

func toNewVendirVS(ver *old.VersionSelection) *new.VersionSelection {
	return &new.VersionSelection{
		Semver: toNewVendirVSS(ver.Semver),
	}
}

func toNewVendirVSS(ver *old.VersionSelectionSemver) *new.VersionSelectionSemver {
	new := &new.VersionSelectionSemver{
		Constraints: ver.Constraints,
		Prereleases: &new.VersionSelectionSemverPrereleases{},
	}

	if ver.Prereleases != nil {
		new.Prereleases.Identifiers = ver.Prereleases.Identifiers
	}

	return new
}

func toOldVendirVS(ver *new.VersionSelection) *old.VersionSelection {
	return &old.VersionSelection{
		Semver: toOldVendirVSS(ver.Semver),
	}
}
