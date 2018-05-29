package xml_saml

// For now we rely on golang XML to do tokenization and punt that problem away
import (
	"bytes"
	"encoding/xml"
	"errors"
	"io"
)

type WrappedElement struct {
	start xml.StartElement
	end   xml.EndElement
}

type AbstractNode struct {
	value    interface{}
	nsLookup map[string]*xml.Name
	parent   *AbstractNode
	children []*AbstractNode
}

func (node *AbstractNode) addWrappedElement(value *WrappedElement) *AbstractNode {
	newNode := AbstractNode{
		value:    value,
		nsLookup: nil,
		parent:   node,
		children: nil,
	}
	for _, attr := range value.start.Attr {
		if isAttributeNamespace(&attr) {
			if len(newNode.nsLookup) == 0 {
				newNode.nsLookup = make(map[string]*xml.Name)
			}
			newNode.nsLookup[attr.Value] = &attr.Name
		}
	}
	node.children = append(node.children, &newNode)
	return &newNode
}

func (node *AbstractNode) addTokenNode(value xml.Token) *AbstractNode {
	newNode := AbstractNode{
		value:    xml.CopyToken(value),
		nsLookup: nil,
		parent:   node,
		children: nil,
	}
	node.children = append(node.children, &newNode)
	return &newNode
}

type NormalizeLineFeedreader struct {
	rdr     io.Reader
	endedCr bool
}

func newNormalizeLineFeedReader(rdr io.Reader) io.Reader {
	return NormalizeLineFeedreader{
		rdr: rdr,
	}
}

func (nl NormalizeLineFeedreader) Read(b []byte) (n int, err error) {
	n, err = nl.rdr.Read(b)

	if n > 0 {
		// @TODO Look at optimizing this with 2 sliding pointers instead of memory wasteful helpers
		// @TODO Look into what we should be doing with `(\r)+\n` as well as `(\r)+[^\n]`
		tmp := bytes.Replace(b[:n], []byte("\r\n"), []byte("\n"), -1)
		if tmp[len(tmp)-1] == 0x0D {
			tmp = tmp[:len(tmp)-1]
		}
		if n != len(tmp) {
			n = len(tmp)
			for i := 0; i < len(tmp); i++ {
				b[i] = tmp[i]
			}
		}
	}

	return
}

// Parse turns utf-8 text into a valid element tree.
func Parse(doc io.Reader) (*AbstractNode, error) {
	dec := xml.NewDecoder(newNormalizeLineFeedReader(doc))

	root := AbstractNode{
		value:    nil,
		nsLookup: nil,
		parent:   nil,
		children: nil,
	}
	elmStack := make([]*AbstractNode, 1)
	elmStack[0] = &root

	// Process the Document
	for {
		t, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		switch tok := t.(type) {
		case xml.StartElement:
			startElm := xml.CopyToken(tok).(xml.StartElement)
			newElm := WrappedElement{
				start: startElm,
				end:   xml.EndElement{},
			}
			newNode := elmStack[len(elmStack)-1].addWrappedElement(&newElm)
			elmStack = append(elmStack, newNode)
		case xml.EndElement:
			pos := len(elmStack) - 1
			if pos <= 0 {
				return nil, errors.New("Unbalanced element tags")
			}

			wrapper, ok := elmStack[pos].value.(*WrappedElement)
			if !ok {
				return nil, errors.New("Unexpected entity in XML stack")
			}
			endElm := xml.CopyToken(tok).(xml.EndElement)
			// @TODO Make sure end element matches start element name
			wrapper.end = endElm
			elmStack = elmStack[:pos]
		case xml.Directive:
			elmStack[len(elmStack)-1].addTokenNode(t)
		case xml.Comment:
			elmStack[len(elmStack)-1].addTokenNode(t)
		case xml.CharData:
			elmStack[len(elmStack)-1].addTokenNode(t)
		case xml.ProcInst:
			elmStack[len(elmStack)-1].addTokenNode(t)
		default:
			return nil, errors.New("Unexpected XML type")
		}
	}

	if len(elmStack) != 1 {
		return nil, errors.New("Unclosed element tags.")
	}

	return &root, nil
}
