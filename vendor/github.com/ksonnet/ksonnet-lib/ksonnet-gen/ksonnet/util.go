package ksonnet

import (
	"github.com/ksonnet/ksonnet-lib/ksonnet-gen/kubespec"
)

const constructorName = "new"

// isMixinRef will check whether a `ObjectRef` refers to an API object
// that can be turned into a mixin. This should be true of the vast
// majority of non-nil `ObjectRef`s. The most common exception is
// `IntOrString`, which should not be turned into a mixin, and should
// instead by transformed into a property method that behaves
// identically to one taking an int or a ref as argument.
func isMixinRef(or *kubespec.ObjectRef) bool {
	return or != nil && *or != "#/definitions/io.k8s.apimachinery.pkg.util.intstr.IntOrString"
}

var specialProperties = map[kubespec.PropertyName]kubespec.PropertyName{
	"apiVersion": "apiVersion",
	"kind":       "kind",
}

var specialPropertiesList = []string{"apiVersion", "kind"}

func isSpecialProperty(pn kubespec.PropertyName) bool {
	_, ok := specialProperties[pn]
	return ok
}
