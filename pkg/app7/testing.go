package app

import (
	"fmt"

	"github.com/maxence-charriere/go-app/v6/pkg/errors"
)

// TestUIDescriptor represents a descriptor that describes a UI element and its
// location from its parents.
type TestUIDescriptor struct {
	// The location of the node. It is used by the TestMatch to find the
	// element to test.
	Path []int

	// The element to compare with the element targetted by Path.
	//
	// If empty, the expected UI element is compared with the root of the tree.
	//
	// Otherwise, each integer represents the index of the element to traverse,
	// from the root's children to the element to compare
	Expected UI
}

// TestPath is a helper function that returns a path to use in a
// TestUIDescriptor.
func TestPath(p ...int) []int {
	return p
}

// TestMatch looks for the element targeted by the descriptor in the given tree
// and reports whether it matches with the expected element.
//
// Eg:
//  tree := app.Div().Body(
//      app.H2().Body(
//          app.Text("foo"),
//      ),
//      app.P().Body(
//          app.Text("bar"),
//      ),
//  )
//
//  // Testing root:
//  err := app.TestMatch(tree, app.TestUIDescriptor{
//      Path:     TestPath(),
//      Expected: app.Div(),
//  })
//  // OK => err == nil
//
//  // Testing h2:
//  err := app.TestMatch(tree, app.TestUIDescriptor{
//      Path:     TestPath(0),
//      Expected: app.H3(),
//  })
//  // KO => err != nil because we ask h2 to match with h3
//
//  // Testing text from p:
//  err = app.TestMatch(tree, app.TestUIDescriptor{
//      Path:     TestPath(1, 0),
//      Expected: app.Text("bar"),
//  })
//  // OK => err == nil
func TestMatch(tree UI, d TestUIDescriptor) error {
	if len(d.Path) != 0 {
		idx := d.Path[0]

		if idx < 0 || idx >= len(tree.children()) {
			return errors.New("ui element to match is out of range").
				Tag("name", d.Expected.name()).
				Tag("kind", d.Expected.Kind()).
				Tag("parent-name", tree.name()).
				Tag("parent-kind", tree.Kind()).
				Tag("parent-children-count", len(tree.children())).
				Tag("index", idx)
		}

		c := tree.children()[idx]
		p := c.parent()

		if p != tree {
			return errors.New("unexpected ui element parent").
				Tag("name", d.Expected.name()).
				Tag("kind", d.Expected.Kind()).
				Tag("parent-name", p.name()).
				Tag("parent-kind", p.Kind()).
				Tag("parent-addr", fmt.Sprintf("%p", p)).
				Tag("expected-parent-name", tree.name()).
				Tag("expected-parent-kind", tree.Kind()).
				Tag("expected-parent-addr", fmt.Sprintf("%p", tree))
		}

		d.Path = d.Path[1:]
		return TestMatch(c, d)
	}

	if d.Expected.name() != tree.name() || d.Expected.Kind() != tree.Kind() {
		return errors.New("the UI element is not matching the descriptor").
			Tag("expeced-name", d.Expected.name()).
			Tag("expected-kind", d.Expected.Kind()).
			Tag("current-name", tree.name()).
			Tag("current-kind", tree.Kind())
	}

	switch d.Expected.Kind() {
	case SimpleText:
		return matchText(tree, d)

	case HTML:
		if err := matchHTMLElemAttrs(tree, d); err != nil {
			return err
		}
		return matchHTMLElemEventHandlers(tree, d)

	case Component:
		return matchComponent(tree, d)

	default:
		return errors.New("the UI element is not matching the descriptor").
			Tag("reason", "unavailable matching for the kind").
			Tag("kind", d.Expected.Kind())
	}
}

func matchText(n UI, d TestUIDescriptor) error {
	a := n.(*text)
	b := d.Expected.(*text)

	if a.value != b.value {
		return errors.New("the UI element is not matching the descriptor").
			Tag("name", a.name()).
			Tag("kind", a.Kind()).
			Tag("reason", "unexpected text value").
			Tag("expected-value", b.value).
			Tag("current-value", a.value)
	}
	return nil
}

func matchHTMLElemAttrs(n UI, d TestUIDescriptor) error {
	aAttrs := n.attributes()
	bAttrs := d.Expected.attributes()

	if len(aAttrs) != len(bAttrs) {
		return errors.New("the UI element is not matching the descriptor").
			Tag("name", n.name()).
			Tag("kind", n.Kind()).
			Tag("reason", "unexpected attributes length").
			Tag("expected-attributes-length", len(bAttrs)).
			Tag("current-attributes-length", len(aAttrs))
	}

	for k, b := range bAttrs {
		a, exists := aAttrs[k]
		if !exists {
			return errors.New("the UI element is not matching the descriptor").
				Tag("name", n.name()).
				Tag("kind", n.Kind()).
				Tag("reason", "an attribute is missing").
				Tag("attribute", k)
		}

		if a != b {
			return errors.New("the UI element is not matching the descriptor").
				Tag("name", n.name()).
				Tag("kind", n.Kind()).
				Tag("reason", "unexpected attribute value").
				Tag("attribute", k).
				Tag("expected-value", b).
				Tag("current-value", a)
		}
	}

	for k := range bAttrs {
		_, exists := bAttrs[k]
		if !exists {
			return errors.New("the UI element is not matching the descriptor").
				Tag("name", n.name()).
				Tag("kind", n.Kind()).
				Tag("reason", "an unexpected attribute is present").
				Tag("attribute", k)
		}
	}

	return nil
}

func matchHTMLElemEventHandlers(n UI, d TestUIDescriptor) error {
	aevents := n.eventHandlers()
	bevents := d.Expected.eventHandlers()

	if len(aevents) != len(bevents) {
		return errors.New("the UI element is not matching the descriptor").
			Tag("name", n.name()).
			Tag("kind", n.Kind()).
			Tag("reason", "unexpected event handlers length").
			Tag("expected-event-handlers-length", len(bevents)).
			Tag("current-event-handlers-length", len(aevents))
	}

	for k := range bevents {
		_, exists := aevents[k]
		if !exists {
			return errors.New("the UI element is not matching the descriptor").
				Tag("name", n.name()).
				Tag("kind", n.Kind()).
				Tag("reason", "an event handler is missing").
				Tag("event-handler", k)
		}
	}

	for k := range bevents {
		_, exists := aevents[k]
		if !exists {
			return errors.New("the UI element is not matching the descriptor").
				Tag("name", n.name()).
				Tag("kind", n.Kind()).
				Tag("reason", "an unexpected event handler is present").
				Tag("event-handler", k)
		}
	}

	return nil

}

func matchComponent(n UI, d TestUIDescriptor) error {
	panic("not implemented")
}