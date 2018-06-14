package xsd

import (
	"fmt"
	"strings"

	"github.com/sezzle/sezzle-go-xml"
)

type Element struct {
	XMLName      xml.Name     `xml:"http://www.w3.org/2001/XMLSchema element"`
	Type         string       `xml:"type,attr"`
	Nillable     string       `xml:"nillable,attr"`
	MinOccurs    string       `xml:"minOccurs,attr"`
	MaxOccurs    string       `xml:"maxOccurs,attr"`
	Form         string       `xml:"form,attr"`
	Name         string       `xml:"name,attr"`
	ComplexTypes *ComplexType `xml:"http://www.w3.org/2001/XMLSchema complexType"`
}

var (
	envName = "ns0"
)

func (self *Element) Encode(enc *xml.Encoder, sr SchemaRepository, ga GetAliaser, params map[string]interface{}, useNamespace, keepUsingNamespace bool, path ...string) (err error) {
	if self.MinOccurs != "" && self.MinOccurs == "0" && !hasPrefix(params, MakePath(append(path, self.Name))) {
		return
	}

	/*if hasPrefix(params, MakePath(append(path, self.Name))) {
		// TODO: figure this out
	}*/

	var namespace, prefix string
	if useNamespace {
		namespace = ga.Namespace()
		prefix = envName
	}

	start := xml.StartElement{
		Name: xml.Name{
			Space:  namespace,
			Prefix: prefix,
			Local:  self.Name,
		},
		/*Name: xml.Name{
			//Space: ga.Namespace(),
			Local: self.Name,
		},
		/*Attr: []xml.Attr{
			xml.Attr{
				Name: xml.Name{
					Space: "xmlns",
					Local: "t",
				},
				Value: ga.Namespace(),
			},
		},*/
	}

	err = enc.EncodeToken(start)
	if err != nil {
		return
	}

	if self.Type != "" {
		parts := strings.Split(self.Type, ":")
		switch len(parts) {
		case 2:
			var schema Schemaer
			schema, err = sr.GetSchema(ga.GetAlias(parts[0]))
			if err != nil {
				return
			}

			err = schema.EncodeType(parts[1], enc, sr, params, keepUsingNamespace, keepUsingNamespace, append(path, self.Name)...)
			if err != nil {
				return
			}
		default:
			err = fmt.Errorf("malformed type '%s' in path %q", self.Type, path)
			return
		}
	} else if self.ComplexTypes != nil {
		for _, e := range self.ComplexTypes.Sequence {
			err = e.Encode(enc, sr, ga, params, keepUsingNamespace, keepUsingNamespace, append(path, self.Name)...)
			if err != nil {
				return
			}
		}
	}

	err = enc.EncodeToken(start.End())
	if err != nil {
		return
	}

	return
}
