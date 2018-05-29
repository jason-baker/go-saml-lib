package xml_saml

import (
	"encoding/xml"
	"fmt"
	"strings"
)

const (
	xmlnsPrefix           = "xmlns"
	xmlPrefix             = "xml"
	attribWhitespaceChars = "\t "
)

// Determines if a given attribute is represents a namespace
func isAttributeNamespace(attr *xml.Attr) bool {
	return strings.HasPrefix(attr.Name.Space, xmlnsPrefix) ||
		(attr.Name.Space == "" && attr.Name.Local == xmlnsPrefix)
}

func trimAttributeWhiteSpace(attr *xml.Attr) {
	attr.Name.Space = strings.Trim(attr.Name.Space, attribWhitespaceChars)
	attr.Name.Local = strings.Trim(attr.Name.Local, attribWhitespaceChars)
}

// compareElementAttributes returns an integer comparing two xml attributes by c14n rules.
// The result will be 0 if a==b, -1 if a < b, and +1 if a > b.
func compareElementAttributes(a *xml.Attr, b *xml.Attr) int {
	iIsNS := isAttributeNamespace(a)
	jIsNS := isAttributeNamespace(b)
	val := 1

	// Handle all 4 cases:
	//  1/2. one of the two attributes isn't a namespace
	//      a. a is the namespace and should come first
	//      b. b is the namespace and should come first
	//  3/4. Both are namespaces or both are attributes
	//      a. namespaces different compare on that
	//      b. namespaces equivalent and compare local names
	if iIsNS != jIsNS {
		if iIsNS {
			val = -1
		}
	} else {
		val = strings.Compare(a.Name.Space, b.Name.Space)
		if val == 0 {
			val = strings.Compare(a.Name.Local, b.Name.Local)
		}
	}

	return val
}

func createPrefixedName(attr *xml.Attr, node *AbstractNode) (name string) {
	name = attr.Name.Local
	if attr.Name.Space == "" {
		return
	}

	if isAttributeNamespace(attr) {
		name = fmt.Sprintf("%s:%s", attr.Name.Space, name)
		return
	}

	for ; node != nil; node = node.parent {
		if len(node.nsLookup) > 0 {
		}
		ns, ok := node.nsLookup[attr.Name.Space]
		if ok {
			// xmlns prefix is ignored on a wrapped element that set it
			if ns.Space == "" && xmlnsPrefix == ns.Local {
				return
			}

			name = ns.Local + ":" + name
			return
		}
	}

	// @TODO properly handle this later
	panic("Attribute malformed namespace reference!")
}

func (elm *WrappedElement) createPrefixedName(node *AbstractNode) (name string) {
	name = elm.start.Name.Local
	if elm.start.Name.Space == "" {
		return
	}

	for ; node != nil; node = node.parent {
		ns, ok := node.nsLookup[elm.start.Name.Space]
		if ok {
			// xmlns prefix is ignored on a wrapped element that set it
			if ns.Space == "" && xmlnsPrefix == ns.Local {
				return
			}

			name = ns.Local + ":" + name
			return
		}
	}

	// @TODO properly handle this later
	panic("Element malformed namespace reference!")
}

// CopyAbstractNode makes a new copy of an abstract node without the children under a new parent.
func (node *AbstractNode) CopyAbstractNode(parent *AbstractNode) (newNode *AbstractNode) {
	newNode = &AbstractNode{
		parent:   parent,
		children: nil,
	}

	switch val := (node.value).(type) {
	case xml.ProcInst:
		newNode.value = xml.CopyToken(val)
	case xml.CharData:
		newNode.value = xml.CopyToken(val)
	case xml.Directive:
		newNode.value = xml.CopyToken(val)
	case xml.Comment:
		newNode.value = xml.CopyToken(val)
	case *WrappedElement:
		newStart := val.start.Copy()
		newNode.value = &WrappedElement{
			start: newStart,
			end:   val.end,
		}

		if len(node.nsLookup) != 0 {
			newNode.nsLookup = make(map[string]*xml.Name)
		}
		for _, attr := range newStart.Attr {
			if isAttributeNamespace(&attr) {
				// Deep copy the Name and take the address
				newName := attr.Name
				newNode.nsLookup[attr.Value] = &newName
			}
		}
	default:
		panic("Unexpected value inside AbstractNode")
	}

	return
}
