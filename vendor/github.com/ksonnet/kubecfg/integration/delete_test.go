// +build integration

package integration

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/pkg/api/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func objNames(list *v1.ConfigMapList) []string {
	ret := make([]string, 0, len(list.Items))
	for _, obj := range list.Items {
		ret = append(ret, obj.GetName())
	}
	return ret
}

var _ = Describe("delete", func() {
	var c corev1.CoreV1Interface
	var ns string
	var objs []runtime.Object

	BeforeEach(func() {
		c = corev1.NewForConfigOrDie(clusterConfigOrDie())
		ns = createNsOrDie(c, "delete")

		objs = []runtime.Object{
			&v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{Name: "foo"},
			},
			&v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{Name: "bar"},
			},
		}
	})
	AfterEach(func() {
		deleteNsOrDie(c, ns)
	})

	Describe("Simple delete", func() {
		JustBeforeEach(func() {
			err := runKubecfgWith([]string{"delete", "-vv", "-n", ns}, objs)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("With no existing state", func() {
			It("should succeed", func() {
				Expect(c.ConfigMaps(ns).List(metav1.ListOptions{})).
					To(WithTransform(objNames, BeEmpty()))
			})
		})

		Context("With existing objects", func() {
			BeforeEach(func() {
				toCreate := []*v1.ConfigMap{}
				for _, cm := range objs {
					toCreate = append(toCreate, cm.(*v1.ConfigMap))
				}
				// .. and one extra (that should not be deleted)
				baz := &v1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{Name: "baz"},
				}
				toCreate = append(toCreate, baz)

				for _, cm := range toCreate {
					_, err := c.ConfigMaps(ns).Create(cm)
					Expect(err).To(Not(HaveOccurred()))
				}
			})

			It("should delete mentioned objects", func() {
				Eventually(func() (*v1.ConfigMapList, error) {
					return c.ConfigMaps(ns).List(metav1.ListOptions{})
				}).Should(WithTransform(objNames, ConsistOf("baz")))
			})
		})
	})
})
