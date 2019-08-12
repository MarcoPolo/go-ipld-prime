package fluent

import (
	"testing"

	. "github.com/warpfork/go-wish"

	ipld "github.com/ipld/go-ipld-prime"
	ipldfree "github.com/ipld/go-ipld-prime/impl/free"
)

func TestRecover(t *testing.T) {
	t.Run("simple lookup error should capture", func(t *testing.T) {
		Wish(t,
			Recover(func() {
				WrapNode(&ipldfree.Node{}).LookupIndex(0).AsString()
				t.Fatal("should not be reached")
			}),
			ShouldEqual,
			Error{ipld.ErrWrongKind{MethodName: "LookupIndex", AppropriateKind: ipld.ReprKindSet_JustList, ActualKind: ipld.ReprKind_Invalid}},
		)
	})
	t.Run("correct lookup should return nil", func(t *testing.T) {
		Wish(t,
			Recover(func() {
				n, _ := ipldfree.NodeBuilder().CreateString("foo")
				WrapNode(n).AsString()
			}),
			ShouldEqual,
			nil,
		)
	})
	t.Run("other panics should continue to rise", func(t *testing.T) {
		Wish(t,
			func() (r interface{}) {
				defer func() { r = recover() }()
				Recover(func() {
					panic("fuqawds")
				})
				return
			}(),
			ShouldEqual,
			"fuqawds",
		)
	})
}
