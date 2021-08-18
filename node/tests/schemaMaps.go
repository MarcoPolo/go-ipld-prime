package tests

import (
	"testing"

	. "github.com/warpfork/go-wish"

	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/ipld/go-ipld-prime/fluent"
	"github.com/ipld/go-ipld-prime/must"
	"github.com/ipld/go-ipld-prime/schema"
)

func SchemaTestMapsContainingMaybe(t *testing.T, engine Engine) {
	ts := schema.TypeSystem{}
	ts.Init()
	ts.Accumulate(schema.SpawnString("String"))
	ts.Accumulate(schema.SpawnMap("Map__String__String",
		"String", "String", false))
	ts.Accumulate(schema.SpawnMap("Map__String__nullableString",
		"String", "String", true))
	engine.Init(t, ts)

	t.Run("non-nullable", func(t *testing.T) {
		np := engine.PrototypeByName("Map__String__String")
		nrp := engine.PrototypeByName("Map__String__String.Repr")
		var n schema.TypedNode
		t.Run("typed-create", func(t *testing.T) {
			n = fluent.MustBuildMap(np, 2, func(ma fluent.MapAssembler) {
				ma.AssembleEntry("one").AssignString("1")
				ma.AssembleEntry("two").AssignString("2")
			}).(schema.TypedNode)
			t.Run("typed-read", func(t *testing.T) {
				Require(t, n.Kind(), ShouldEqual, datamodel.Kind_Map)
				Wish(t, n.Length(), ShouldEqual, int64(2))
				Wish(t, must.String(must.Node(n.LookupByString("one"))), ShouldEqual, "1")
				Wish(t, must.String(must.Node(n.LookupByString("two"))), ShouldEqual, "2")
				_, err := n.LookupByString("miss")
				Wish(t, err, ShouldBeSameTypeAs, datamodel.ErrNotExists{})
			})
			t.Run("repr-read", func(t *testing.T) {
				nr := n.Representation()
				Require(t, nr.Kind(), ShouldEqual, datamodel.Kind_Map)
				Wish(t, nr.Length(), ShouldEqual, int64(2))
				Wish(t, must.String(must.Node(nr.LookupByString("one"))), ShouldEqual, "1")
				Wish(t, must.String(must.Node(nr.LookupByString("two"))), ShouldEqual, "2")
				_, err := nr.LookupByString("miss")
				Wish(t, err, ShouldBeSameTypeAs, datamodel.ErrNotExists{})
			})
		})
		t.Run("repr-create", func(t *testing.T) {
			nr := fluent.MustBuildMap(nrp, 2, func(ma fluent.MapAssembler) {
				ma.AssembleEntry("one").AssignString("1")
				ma.AssembleEntry("two").AssignString("2")
			})
			Wish(t, datamodel.DeepEqual(n, nr), ShouldEqual, true)
		})
	})
	t.Run("nullable", func(t *testing.T) {
		np := engine.PrototypeByName("Map__String__nullableString")
		nrp := engine.PrototypeByName("Map__String__nullableString.Repr")
		var n schema.TypedNode
		t.Run("typed-create", func(t *testing.T) {
			n = fluent.MustBuildMap(np, 2, func(ma fluent.MapAssembler) {
				ma.AssembleEntry("one").AssignString("1")
				ma.AssembleEntry("none").AssignNull()
			}).(schema.TypedNode)
			t.Run("typed-read", func(t *testing.T) {
				Require(t, n.Kind(), ShouldEqual, datamodel.Kind_Map)
				Wish(t, n.Length(), ShouldEqual, int64(2))
				Wish(t, must.String(must.Node(n.LookupByString("one"))), ShouldEqual, "1")
				Wish(t, must.Node(n.LookupByString("none")), ShouldEqual, datamodel.Null)
				_, err := n.LookupByString("miss")
				Wish(t, err, ShouldBeSameTypeAs, datamodel.ErrNotExists{})
			})
			t.Run("repr-read", func(t *testing.T) {
				nr := n.Representation()
				Require(t, nr.Kind(), ShouldEqual, datamodel.Kind_Map)
				Wish(t, nr.Length(), ShouldEqual, int64(2))
				Wish(t, must.String(must.Node(nr.LookupByString("one"))), ShouldEqual, "1")
				Wish(t, must.Node(nr.LookupByString("none")), ShouldEqual, datamodel.Null)
				_, err := nr.LookupByString("miss")
				Wish(t, err, ShouldBeSameTypeAs, datamodel.ErrNotExists{})
			})
		})
		t.Run("repr-create", func(t *testing.T) {
			nr := fluent.MustBuildMap(nrp, 2, func(ma fluent.MapAssembler) {
				ma.AssembleEntry("one").AssignString("1")
				ma.AssembleEntry("none").AssignNull()
			})
			Wish(t, datamodel.DeepEqual(n, nr), ShouldEqual, true)
		})
	})
}

// TestMapsContainingMaps is probing *two* things:
//   - that maps can nest, obviously
//   - that representation semantics are held correctly when we recurse, both in builders and in reading
// To cover that latter situation, this depends on structs (so we can use rename directives on the representation to make it distinctive).
func SchemaTestMapsContainingMaps(t *testing.T, engine Engine) {
	ts := schema.TypeSystem{}
	ts.Init()
	ts.Accumulate(schema.SpawnString("String"))
	ts.Accumulate(schema.SpawnStruct("Frub", // "type Frub struct { field String (rename "encoded") }"
		[]schema.StructField{
			schema.SpawnStructField("field", "String", false, false), // plain field.
		},
		schema.SpawnStructRepresentationMap(map[string]string{
			"field": "encoded",
		}),
	))
	ts.Accumulate(schema.SpawnMap("Map__String__Frub", // "{String:Frub}"
		"String", "Frub", false))
	ts.Accumulate(schema.SpawnMap("Map__String__nullableMap__String__Frub", // "{String:nullable {String:Frub}}"
		"String", "Map__String__Frub", true))
	engine.Init(t, ts)

	np := engine.PrototypeByName("Map__String__nullableMap__String__Frub")
	nrp := engine.PrototypeByName("Map__String__nullableMap__String__Frub.Repr")
	creation := func(t *testing.T, np datamodel.NodePrototype, fieldName string) schema.TypedNode {
		return fluent.MustBuildMap(np, 3, func(ma fluent.MapAssembler) {
			ma.AssembleEntry("one").CreateMap(2, func(ma fluent.MapAssembler) {
				ma.AssembleEntry("zot").CreateMap(1, func(ma fluent.MapAssembler) { ma.AssembleEntry(fieldName).AssignString("11") })
				ma.AssembleEntry("zop").CreateMap(1, func(ma fluent.MapAssembler) { ma.AssembleEntry(fieldName).AssignString("12") })
			})
			ma.AssembleEntry("two").CreateMap(1, func(ma fluent.MapAssembler) {
				ma.AssembleEntry("zim").CreateMap(1, func(ma fluent.MapAssembler) { ma.AssembleEntry(fieldName).AssignString("21") })
			})
			ma.AssembleEntry("none").AssignNull()
		}).(schema.TypedNode)
	}
	reading := func(t *testing.T, n datamodel.Node, fieldName string) {
		withNode(n, func(n datamodel.Node) {
			Require(t, n.Kind(), ShouldEqual, datamodel.Kind_Map)
			Wish(t, n.Length(), ShouldEqual, int64(3))
			withNode(must.Node(n.LookupByString("one")), func(n datamodel.Node) {
				Require(t, n.Kind(), ShouldEqual, datamodel.Kind_Map)
				Wish(t, n.Length(), ShouldEqual, int64(2))
				withNode(must.Node(n.LookupByString("zot")), func(n datamodel.Node) {
					Require(t, n.Kind(), ShouldEqual, datamodel.Kind_Map)
					Wish(t, n.Length(), ShouldEqual, int64(1))
					Wish(t, must.String(must.Node(n.LookupByString(fieldName))), ShouldEqual, "11")
				})
				withNode(must.Node(n.LookupByString("zop")), func(n datamodel.Node) {
					Require(t, n.Kind(), ShouldEqual, datamodel.Kind_Map)
					Wish(t, n.Length(), ShouldEqual, int64(1))
					Wish(t, must.String(must.Node(n.LookupByString(fieldName))), ShouldEqual, "12")
				})
			})
			withNode(must.Node(n.LookupByString("two")), func(n datamodel.Node) {
				Wish(t, n.Length(), ShouldEqual, int64(1))
				withNode(must.Node(n.LookupByString("zim")), func(n datamodel.Node) {
					Require(t, n.Kind(), ShouldEqual, datamodel.Kind_Map)
					Wish(t, n.Length(), ShouldEqual, int64(1))
					Wish(t, must.String(must.Node(n.LookupByString(fieldName))), ShouldEqual, "21")
				})
			})
			withNode(must.Node(n.LookupByString("none")), func(n datamodel.Node) {
				Wish(t, datamodel.DeepEqual(n, datamodel.Null), ShouldEqual, true)
			})
			_, err := n.LookupByString("miss")
			Wish(t, err, ShouldBeSameTypeAs, datamodel.ErrNotExists{})
		})
	}
	var n schema.TypedNode
	t.Run("typed-create", func(t *testing.T) {
		n = creation(t, np, "field")
		t.Run("typed-read", func(t *testing.T) {
			reading(t, n, "field")
		})
		t.Run("repr-read", func(t *testing.T) {
			reading(t, n.Representation(), "encoded")
		})
	})
	t.Run("repr-create", func(t *testing.T) {
		nr := creation(t, nrp, "encoded")
		Wish(t, datamodel.DeepEqual(n, nr), ShouldEqual, true)
	})
}

func SchemaTestMapsWithComplexKeys(t *testing.T, engine Engine) {
	ts := schema.TypeSystem{}
	ts.Init()
	ts.Accumulate(schema.SpawnString("String"))
	ts.Accumulate(schema.SpawnStruct("StringyStruct",
		[]schema.StructField{
			schema.SpawnStructField("foo", "String", false, false),
			schema.SpawnStructField("bar", "String", false, false),
		},
		schema.SpawnStructRepresentationStringjoin(":"),
	))
	ts.Accumulate(schema.SpawnMap("Map__StringyStruct__String",
		"StringyStruct", "String", false))
	engine.Init(t, ts)

	np := engine.PrototypeByName("Map__StringyStruct__String")
	nrp := engine.PrototypeByName("Map__StringyStruct__String.Repr")
	var n schema.TypedNode
	t.Run("typed-create", func(t *testing.T) {
		n = fluent.MustBuildMap(np, 3, func(ma fluent.MapAssembler) {
			ma.AssembleKey().CreateMap(2, func(ma fluent.MapAssembler) {
				ma.AssembleEntry("foo").AssignString("a")
				ma.AssembleEntry("bar").AssignString("b")
			})
			ma.AssembleValue().AssignString("1")
			ma.AssembleKey().CreateMap(2, func(ma fluent.MapAssembler) {
				ma.AssembleEntry("foo").AssignString("c")
				ma.AssembleEntry("bar").AssignString("d")
			})
			ma.AssembleValue().AssignString("2")
			ma.AssembleKey().CreateMap(2, func(ma fluent.MapAssembler) {
				ma.AssembleEntry("foo").AssignString("e")
				ma.AssembleEntry("bar").AssignString("f")
			})
			ma.AssembleValue().AssignString("3")
		}).(schema.TypedNode)
		t.Run("typed-read", func(t *testing.T) {
			Require(t, n.Kind(), ShouldEqual, datamodel.Kind_Map)
			Wish(t, n.Length(), ShouldEqual, int64(3))
			n2 := must.Node(n.LookupByString("c:d"))
			Require(t, n2.Kind(), ShouldEqual, datamodel.Kind_String)
			Wish(t, must.String(n2), ShouldEqual, "2")
		})
		t.Run("repr-read", func(t *testing.T) {
			nr := n.Representation()
			Require(t, nr.Kind(), ShouldEqual, datamodel.Kind_Map)
			Wish(t, nr.Length(), ShouldEqual, int64(3))
			n2 := must.Node(nr.LookupByString("c:d"))
			Require(t, n2.Kind(), ShouldEqual, datamodel.Kind_String)
			Wish(t, must.String(n2), ShouldEqual, "2")
		})
	})
	t.Run("repr-create", func(t *testing.T) {
		nr := fluent.MustBuildMap(nrp, 3, func(ma fluent.MapAssembler) {
			ma.AssembleEntry("a:b").AssignString("1")
			ma.AssembleEntry("c:d").AssignString("2")
			ma.AssembleEntry("e:f").AssignString("3")
		})
		Wish(t, datamodel.DeepEqual(n, nr), ShouldEqual, true)
	})
}
