package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/discovery"
)

// Format v0.0.0(-master+$Format:%h$)
var gitVersionRe = regexp.MustCompile("v([0-9])+.([0-9])+.[0-9]+.*")

// ServerVersion captures k8s major.minor version in a parsed form
type ServerVersion struct {
	Major int
	Minor int
}

func parseGitVersion(gitVersion string) (ServerVersion, error) {
	parsedVersion := gitVersionRe.FindStringSubmatch(gitVersion)
	if len(parsedVersion) != 3 {
		return ServerVersion{}, fmt.Errorf("Unable to parse git version %s", gitVersion)
	}
	var ret ServerVersion
	var err error
	ret.Major, err = strconv.Atoi(parsedVersion[1])
	if err != nil {
		return ServerVersion{}, err
	}
	ret.Minor, err = strconv.Atoi(parsedVersion[2])
	if err != nil {
		return ServerVersion{}, err
	}
	return ret, nil
}

// ParseVersion parses version.Info into a ServerVersion struct
func ParseVersion(v *version.Info) (ServerVersion, error) {
	var ret ServerVersion
	var err error
	ret.Major, err = strconv.Atoi(v.Major)
	if err != nil {
		// Try to parse using GitVersion
		return parseGitVersion(v.GitVersion)
	}

	// trim "+" in minor version (happened on GKE)
	v.Minor = strings.TrimSuffix(v.Minor, "+")
	ret.Minor, err = strconv.Atoi(v.Minor)
	if err != nil {
		// Try to parse using GitVersion
		return parseGitVersion(v.GitVersion)
	}
	return ret, err
}

// FetchVersion fetches version information from discovery client, and parses
func FetchVersion(v discovery.ServerVersionInterface) (ret ServerVersion, err error) {
	version, err := v.ServerVersion()
	if err != nil {
		return ServerVersion{}, err
	}
	return ParseVersion(version)
}

// GetDefaultVersion returns a default server version. This value will be updated
// periodically to match a current/popular version corresponding to the age of this code
// Current default version: 1.8
func GetDefaultVersion() ServerVersion {
	return ServerVersion{Major: 1, Minor: 8}
}

// Compare returns -1/0/+1 iff v is less than / equal / greater than major.minor
func (v ServerVersion) Compare(major, minor int) int {
	a := v.Major
	b := major

	if a == b {
		a = v.Minor
		b = minor
	}

	var res int
	if a > b {
		res = 1
	} else if a == b {
		res = 0
	} else {
		res = -1
	}
	return res
}

func (v ServerVersion) String() string {
	return fmt.Sprintf("%d.%d", v.Major, v.Minor)
}

// SetMetaDataAnnotation sets an annotation value
func SetMetaDataAnnotation(obj metav1.Object, key, value string) {
	a := obj.GetAnnotations()
	if a == nil {
		a = make(map[string]string)
	}
	a[key] = value
	obj.SetAnnotations(a)
}

// ResourceNameFor returns a lowercase plural form of a type, for
// human messages.  Returns lowercased kind if discovery lookup fails.
func ResourceNameFor(disco discovery.ServerResourcesInterface, o runtime.Object) string {
	gvk := o.GetObjectKind().GroupVersionKind()
	rls, err := disco.ServerResourcesForGroupVersion(gvk.GroupVersion().String())
	if err != nil {
		log.Debugf("Discovery failed for %s: %s, falling back to kind", gvk, err)
		return strings.ToLower(gvk.Kind)
	}

	for _, rl := range rls.APIResources {
		if rl.Kind == gvk.Kind {
			return rl.Name
		}
	}

	log.Debugf("Discovery failed to find %s, falling back to kind", gvk)
	return strings.ToLower(gvk.Kind)
}

// FqName returns "namespace.name"
func FqName(o metav1.Object) string {
	if o.GetNamespace() == "" {
		return o.GetName()
	}
	return fmt.Sprintf("%s.%s", o.GetNamespace(), o.GetName())
}
