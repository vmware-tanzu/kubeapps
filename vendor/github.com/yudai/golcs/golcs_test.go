package lcs_test

import (
	. "github.com/yudai/golcs"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Lcs", func() {
	Describe("Lcs", func() {
		It("Calculates Longest Common Subsequence", func() {
			var left, right []interface{}
			var lcs Lcs

			left = []interface{}{1, 2, 3}
			right = []interface{}{2, 3}
			lcs = New(left, right)
			Expect(lcs.IndexPairs()).To(Equal([]IndexPair{{1, 0}, {2, 1}}))
			Expect(lcs.Values()).To(Equal([]interface{}{2, 3}))
			Expect(lcs.Length()).To(Equal(2))

			left = []interface{}{2, 3}
			right = []interface{}{1, 2, 3}
			lcs = New(left, right)
			Expect(lcs.IndexPairs()).To(Equal([]IndexPair{{0, 1}, {1, 2}}))
			Expect(lcs.Values()).To(Equal([]interface{}{2, 3}))
			Expect(lcs.Length()).To(Equal(2))

			left = []interface{}{2, 3}
			right = []interface{}{2, 5, 3}
			lcs = New(left, right)
			Expect(lcs.IndexPairs()).To(Equal([]IndexPair{{0, 0}, {1, 2}}))
			Expect(lcs.Values()).To(Equal([]interface{}{2, 3}))
			Expect(lcs.Length()).To(Equal(2))

			left = []interface{}{2, 3, 3}
			right = []interface{}{2, 5, 3}
			lcs = New(left, right)
			Expect(lcs.IndexPairs()).To(Equal([]IndexPair{{0, 0}, {2, 2}}))
			Expect(lcs.Values()).To(Equal([]interface{}{2, 3}))
			Expect(lcs.Length()).To(Equal(2))

			left = []interface{}{1, 2, 5, 3, 1, 1, 5, 8, 3}
			right = []interface{}{1, 2, 3, 3, 4, 4, 5, 1, 6}
			lcs = New(left, right)
			Expect(lcs.Values()).To(Equal([]interface{}{1, 2, 5, 1}))
			Expect(lcs.Length()).To(Equal(4))

			left = []interface{}{}
			right = []interface{}{2, 5, 3}
			lcs = New(left, right)
			Expect(lcs.Values()).To(Equal([]interface{}{}))
			Expect(lcs.Length()).To(Equal(0))

			left = []interface{}{3, 4}
			right = []interface{}{}
			lcs = New(left, right)
			Expect(lcs.Values()).To(Equal([]interface{}{}))
			Expect(lcs.Length()).To(Equal(0))

			left = []interface{}{"foo"}
			right = []interface{}{"baz", "foo"}
			lcs = New(left, right)
			Expect(lcs.Values()).To(Equal([]interface{}{"foo"}))
			Expect(lcs.Length()).To(Equal(1))

			leftBytes := []byte("TGAGTA")
			rightBytes := []byte("GATA")
			left = make([]interface{}, len(leftBytes))
			for i, v := range leftBytes {
				left[i] = v
			}
			right = make([]interface{}, len(rightBytes))
			for i, v := range rightBytes {
				right[i] = v
			}
			lcs = New(left, right)
			Expect(lcs.Values()).To(Equal([]interface{}{byte('G'), byte('A'), byte('T'), byte('A')}))
			Expect(lcs.Length()).To(Equal(4))

		})
	})
})
