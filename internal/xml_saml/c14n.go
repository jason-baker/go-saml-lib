package xml_saml

import (
	"encoding/xml"
	"errors"
	"sort"
)

type CanonicalizationType string

const (
	XML_C14N          CanonicalizationType = "http://www.w3.org/TR/2001/REC-xml-c14n-20010315"
	XML_C14N_COMMENTS                      = "http://www.w3.org/TR/2001/REC-xml-c14n-20010315#WithComments"
	//XML_C14N11                                 = "http://www.w3.org/2001/10/xml-exc-c14n#"
	//XML_C14N11_COMMENTS                        = "http://www.w3.org/2001/10/xml-exc-c14n#WithComments"
	XML_EXC_C14N          = "http://www.w3.org/2001/10/xml-exc-c14n#"
	XML_EXC_C14N_COMMENTS = "http://www.w3.org/2001/10/xml-exc-c14n#WithComments"
)

// filterOpts provides common filtering operations across canonicalizations.
type filterOpts struct {
	isDocRoot        bool
	keepComments     bool
	filteredProcInst []string
}

// Attribute sorting interface
type AttrSort []xml.Attr

func (a AttrSort) Len() int      { return len(a) }
func (a AttrSort) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a AttrSort) Less(i, j int) bool {
	return -1 == compareElementAttributes(&a[i], &a[j])
}

// c14nNode returns the canonicalized node; if the node is to be removed it returns nil.
func (node *AbstractNode) c14nNode(newParent *AbstractNode, opts *filterOpts) (newNode *AbstractNode) {
	newNode = nil
	switch val := (node.value).(type) {
	case xml.ProcInst:
		for _, filter := range opts.filteredProcInst {
			if val.Target == filter {
				return
			}
		}
		// @TODO Handle instructions somewhere?
		newNode = node.CopyAbstractNode(newParent)
	case xml.CharData:
		if opts.isDocRoot {
			// @TODO Normalize whitespace/check for all whitespace instead of arbitrary nuke
			return
		}
		newNode = node.CopyAbstractNode(newParent)
	case xml.Directive:
		// @TODO Handle directives
	case xml.Comment:
		if opts.keepComments {
			newNode = node.CopyAbstractNode(newParent)
		}
	case *WrappedElement:
		newNode = node.CopyAbstractNode(newParent)
		newElm, _ := newNode.value.(*WrappedElement)

		// Clean up the elements and sort them
		for _, attrib := range newElm.start.Attr {
			trimAttributeWhiteSpace(&attrib)
		}
		sort.Sort(AttrSort(newElm.start.Attr))
	default:
		panic("Unexpected value inside AbstractNode")
	}

	return
}

func (oldNode *AbstractNode) xml_c14n(newParent *AbstractNode, opts *filterOpts) error {
	activeOpts := opts

	// Do special work in the document root
	if oldNode.value == nil {
		rootOpts := filterOpts{}
		rootOpts = *activeOpts
		rootOpts.isDocRoot = true
		activeOpts = &rootOpts
	}

	for _, oldChild := range oldNode.children {
		newChild := oldChild.c14nNode(newParent, activeOpts)
		if newChild != nil {
			newParent.children = append(newParent.children, newChild)
			if len(oldChild.children) > 0 {
				err := oldChild.xml_c14n(newChild, opts)
				if err != nil {
					return err
				}
			}
		}
	}

	// Trime out repeated whitespace lines
	if activeOpts.isDocRoot {
		expanded := make([]*AbstractNode, len(newParent.children)*2-1)
		// Count the number of elements to be stripped
		for i, v := range newParent.children {
			expanded[i*2] = v
			if i < len(newParent.children)-1 {
				expanded[i*2+1] = &AbstractNode{
					value:    xml.CharData([]byte{'\x0a'}),
					nsLookup: nil,
					parent:   nil,
					children: nil,
				}
			}
		}
		newParent.children = expanded
	}

	return nil
}

func (root *AbstractNode) Canonicalize(c14nType CanonicalizationType) (*AbstractNode, error) {
	newRoot := AbstractNode{
		value:    nil,
		nsLookup: nil,
		parent:   nil,
		children: nil,
	}
	switch c14nType {
	case XML_C14N:
		root.xml_c14n(&newRoot, &filterOpts{
			keepComments:     false,
			filteredProcInst: []string{"xml"},
		})
		return &newRoot, nil
	}
	return nil, errors.New("Unsupported canonicalization type")
}
