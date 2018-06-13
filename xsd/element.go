package xsd

import (
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/davecgh/go-spew/spew"
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

type CustomStartElement struct {
	xml.StartElement
	Prefix Namespace
}

type Namespace struct {
	Prefix string
	Value  string
}

func (self *Element) Encode(enc *xml.Encoder, sr SchemaRepository, ga GetAliaser, params map[string]interface{}, path ...string) (err error) {
	fmt.Println("SELF:", self.Name)
	spew.Dump(self)

	if self.MinOccurs != "" && self.MinOccurs == "0" && !hasPrefix(params, MakePath(append(path, self.Name))) {
		return
	}

	/*if hasPrefix(params, MakePath(append(path, self.Name))) {
		// TODO: figure this out
	}*/

	envName := "ns0"
	start := xml.StartElement{
		Name: xml.Name{
			Local: envName + ":" + "Envelope",
		},
		Attr: []xml.Attr{
			xml.Attr{
				Name: xml.Name{
					Space: "xmlns",
					Local: envName,
				},
				Value: ga.Namespace(),
			},
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

			err = schema.EncodeType(parts[1], enc, sr, params, append(path, self.Name)...)
			if err != nil {
				return
			}
		default:
			err = fmt.Errorf("malformed type '%s' in path %q", self.Type, path)
			return
		}
	} else if self.ComplexTypes != nil {
		for _, e := range self.ComplexTypes.Sequence {
			err = e.Encode(enc, sr, ga, params, append(path, self.Name)...)
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
