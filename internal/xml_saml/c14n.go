package xml_saml

import (
	"encoding/xml"
	"errors"
	"fmt"
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

// c14nNode returns the canonicalized node; if the node is to be removed it returns nil.
func (node *AbstractNode) c14nNode(opts *filterOpts) (newNode *AbstractNode) {
	switch val := (node.value).(type) {
	case xml.ProcInst:
		for _, filter := range opts.filteredProcInst {
			if val.Target == filter {
				//print(fmt.Sprintf("dropping procInst (%p): %T %+v\n", &val, val, val))
				return nil
			}
		}
		return &AbstractNode{value: xml.CopyToken(val), children: nil}
	case xml.CharData:
		newCharData := xml.CopyToken(val)
		if opts.isDocRoot {
			// @TODO Normalize whitespace/check for all whitespace instead of arbitrary nuke
			return nil
		}
		return &AbstractNode{value: newCharData, children: nil}
	case xml.Directive:
		// @TODO Handle directives
		//print(fmt.Sprintf("dropping directive (%p): %T %+v\n", val, val, val))
	case xml.Comment:
		if opts.keepComments {
			return &AbstractNode{value: xml.CopyToken(val), children: nil}
		}
		//print(fmt.Sprintf("dropping comment (%p): %T %+v\n", val, val, val))
	case *WrappedElement:
		return &AbstractNode{
			value: &WrappedElement{
				start: val.start.Copy(),
				end:   val.end,
			},
			children: nil,
		}
	default:
		print(fmt.Sprintf("value (%p): %T %+v\n", val, val, val))
		panic("womp")
	}

	return nil
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
		newChild := oldChild.c14nNode(activeOpts)
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
