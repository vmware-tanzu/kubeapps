// Copyright 2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package helm

import "fmt"

// SecretNameForRepo returns the name of a secret for an apprepo
func SecretNameForRepo(repoName string) string {
	return fmt.Sprintf("apprepo-%s", repoName)
}

// SecretNameForNamespacedRepo returns a name suitable for recording a copy of
// a per-namespace repository secret in the kubeapps namespace.
func SecretNameForNamespacedRepo(repoName, namespace string) string {
	return fmt.Sprintf("%s-%s", namespace, SecretNameForRepo(repoName))
}
