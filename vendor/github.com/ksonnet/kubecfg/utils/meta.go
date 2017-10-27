package utils

import (
	"fmt"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/discovery"
)

// ServerVersion captures k8s major.minor version in a parsed form
type ServerVersion struct {
	Major int
	Minor int
}

// ParseVersion parses version.Info into a ServerVersion struct
func ParseVersion(v *version.Info) (ret ServerVersion, err error) {
	ret.Major, err = strconv.Atoi(v.Major)
	if err != nil {
		return
	}
	ret.Minor, err = strconv.Atoi(v.Minor)
	if err != nil {
		return
	}
	return
}

// FetchVersion fetches version information from discovery client, and parses
func FetchVersion(v discovery.ServerVersionInterface) (ret ServerVersion, err error) {
	version, err := v.ServerVersion()
	if err != nil {
		return ServerVersion{}, err
	}
	return ParseVersion(version)
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
