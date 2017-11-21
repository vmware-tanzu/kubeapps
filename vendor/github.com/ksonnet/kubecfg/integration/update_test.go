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

func cmData(cm *v1.ConfigMap) map[string]string {
	return cm.Data
}

var _ = Describe("update", func() {
	var c corev1.CoreV1Interface
	var ns string
	const cmName = "testcm"

	BeforeEach(func() {
		c = corev1.NewForConfigOrDie(clusterConfigOrDie())
		ns = createNsOrDie(c, "update")
	})
	AfterEach(func() {
		deleteNsOrDie(c, ns)
	})

	Describe("A simple update", func() {
		var cm *v1.ConfigMap
		BeforeEach(func() {
			cm = &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{Name: cmName},
				Data:       map[string]string{"foo": "bar"},
			}
		})

		JustBeforeEach(func() {
			err := runKubecfgWith([]string{"update", "-vv", "-n", ns}, []runtime.Object{cm})
			Expect(err).NotTo(HaveOccurred())
		})

		Context("With no existing state", func() {
			It("should produce expected object", func() {
				Expect(c.ConfigMaps(ns).Get("testcm", metav1.GetOptions{})).
					To(WithTransform(cmData, HaveKeyWithValue("foo", "bar")))
			})
		})

		Context("With existing object", func() {
			BeforeEach(func() {
				_, err := c.ConfigMaps(ns).Create(cm)
				Expect(err).To(Not(HaveOccurred()))
			})

			It("should succeed", func() {

				Expect(c.ConfigMaps(ns).Get("testcm", metav1.GetOptions{})).
					To(WithTransform(cmData, HaveKeyWithValue("foo", "bar")))
			})
		})

		Context("With modified object", func() {
			BeforeEach(func() {
				otherCm := &v1.ConfigMap{
					ObjectMeta: cm.ObjectMeta,
					Data:       map[string]string{"foo": "not bar"},
				}

				_, err := c.ConfigMaps(ns).Create(otherCm)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should update the object", func() {
				Expect(c.ConfigMaps(ns).Get("testcm", metav1.GetOptions{})).
					To(WithTransform(cmData, HaveKeyWithValue("foo", "bar")))
			})
		})
	})

	Describe("An update with mixed namespaces", func() {
		var ns2 string
		BeforeEach(func() {
			ns2 = createNsOrDie(c, "update")
		})
		AfterEach(func() {
			deleteNsOrDie(c, ns2)
		})

		var objs []runtime.Object
		BeforeEach(func() {
			objs = []runtime.Object{
				&v1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{Name: "nons"},
				},
				&v1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: "ns1"},
				},
				&v1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{Namespace: ns2, Name: "ns2"},
				},
			}
		})

		JustBeforeEach(func() {
			err := runKubecfgWith([]string{"update", "-vv", "-n", ns}, objs)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should create objects in the correct namespaces", func() {
			Expect(c.ConfigMaps(ns).Get("nons", metav1.GetOptions{})).
				NotTo(BeNil())

			Expect(c.ConfigMaps(ns).Get("ns1", metav1.GetOptions{})).
				NotTo(BeNil())

			Expect(c.ConfigMaps(ns2).Get("ns2", metav1.GetOptions{})).
				NotTo(BeNil())
		})
	})
})
