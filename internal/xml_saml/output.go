package xml_saml

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

func (elm *WrappedElement) iprint(indent string) {
	print(fmt.Sprintf("%svalue (%p): %T %+v\n", indent, &elm.start, elm.start, elm.start))
}

func (node *AbstractNode) iprint(indent string) {
	var endElm *xml.EndElement
	endElm = nil

	switch val := (node.value).(type) {
	case nil:
		print("Document Root\n")
	case xml.ProcInst:
		print(fmt.Sprintf("%svalue (%p): %T %+v\n", indent, &val, val, val))
	case xml.CharData:
		print(fmt.Sprintf("%svalue (%p): %T %+v\n", indent, val, val, val))
	case xml.Directive:
		print(fmt.Sprintf("%svalue (%p): %T %+v\n", indent, val, val, val))
	case xml.Comment:
		print(fmt.Sprintf("%svalue (%p): %T %+v\n", indent, val, val, val))
	case *WrappedElement:
		val.iprint(indent)
		endElm = &val.end
	default:
		print(fmt.Sprintf("%svalue (%p): %T %+v\n", indent, val, val, val))
		panic("Unexpected")
	}
	for _, value := range node.children {
		value.iprint(indent + "  ")
	}
	if endElm != nil {
		print(fmt.Sprintf("%svalue (%p): %T %+v\n", indent, endElm, endElm, endElm))
	}
}

func (node *AbstractNode) debugPrint() {
	node.iprint("")
}

func attrib_string_array(attribs []xml.Attr) []string {
	strs := make([]string, len(attribs))
	for i, v := range attribs {
		strs[i] = fmt.Sprintf("%s=\"%s\"", v.Name, v.Value)
	}
	return strs
}

func (node *AbstractNode) Print(writer io.Writer) {
	var out string
	end := ""
	//
	switch val := node.value.(type) {
	case nil:
		// Do nothing
	case xml.ProcInst:
		if len(val.Inst) > 0 {
			out = fmt.Sprintf("<?%s %s?>", val.Target, val.Inst)
		} else {
			out = fmt.Sprintf("<?%s?>", val.Target)
		}
	case xml.Comment:
		out = fmt.Sprintf("<!--%s-->", val)
	case xml.CharData:
		out = fmt.Sprintf("%s", val)
	case xml.Directive:
		out = fmt.Sprintf("<!%s>", val)
		print(fmt.Sprintf("Test Data: `%s`", out))
	case *WrappedElement:
		name := val.start.Name.Local
		end = val.end.Name.Local
		if len(val.start.Name.Space) > 0 {
			name = fmt.Sprintf("%s:%s", val.start.Name.Space, name)
			end = fmt.Sprintf("%s:%s", val.end.Name.Space, end)
		}
		if len(val.start.Attr) > 0 {
			out = fmt.Sprintf("<%s %s>", name,
				strings.Join(attrib_string_array(val.start.Attr), " "))
		} else {
			out = fmt.Sprintf("<%s>", name)
		}
		end = fmt.Sprintf("</%s>", end)
	default:
		panic("Unexpected XML type!")
	}

	fmt.Fprint(writer, out)
	for _, child := range node.children {
		child.Print(writer)
	}
	fmt.Fprint(writer, end)
}
