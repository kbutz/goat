package xsd

import (
	"fmt"
	"github.com/pkg/errors"
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

func (self *Element) Encode(enc *xml.Encoder, sr SchemaRepository, ga GetAliaser, params map[string]interface{}, useNamespace, keepUsingNamespace bool, path ...string) error {
	if self.MinOccurs != "" && self.MinOccurs == "0" && !hasPrefix(params, MakePath(append(path, self.Name))) {
		return nil
	}

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
	}

	err := enc.EncodeToken(start)
	if err != nil {
		err = errors.Wrap(err, "Error encoding start token")
		return err
	}

	// TODO: Remove all of these debugging logs
	fmt.Println("*Element: " + fmt.Sprintf("%+v", self))
	//fmt.Println("Name: " + fmt.Sprintf("%+v", self.Name))
	//fmt.Println("Type: " + fmt.Sprintf("%+v", self.Type))
	//fmt.Println("ComplexTypes: " + fmt.Sprintf("%+v", self.ComplexTypes))
	if self.ComplexTypes != nil {
		fmt.Println("ComplexTypes.Sequence" + fmt.Sprintf("%+v", self.ComplexTypes.Sequence))
	}
	if self.Type != "" {
		parts := strings.Split(self.Type, ":")
		switch len(parts) {
		case 2:
			var schema Schemaer
			schema, err = sr.GetSchema(ga.GetAlias(parts[0]))
			if err != nil {
				err = errors.Wrap(err, "Error getting schema for "+parts[0])
				return err
			}

			// This isn't helpful - the schema is the FULL wsdl map
			//fmt.Println("self.Type != '', schema: " + fmt.Sprintf("%+v", schema))
			err = schema.EncodeType(parts[1], enc, sr, params, keepUsingNamespace, keepUsingNamespace, append(path, self.Name)...)
			if err != nil {
				err = errors.Wrap(err, "Error encoding type for "+parts[1])
				return err
			}
		default:
			err = fmt.Errorf("malformed type '%s' in path %q", self.Type, path)
			return err
		}
	} else if self.ComplexTypes != nil {
		for _, e := range self.ComplexTypes.Sequence.Elements {
			err = e.Encode(enc, sr, ga, params, keepUsingNamespace, keepUsingNamespace, append(path, self.Name)...)
			if err != nil {
				err = errors.Wrap(err, "Error encoding ComplexTypes.Sequence.Elements")
				return err
			}
		}

		for _, e := range self.ComplexTypes.Sequence.Choice {
			err = e.Encode(enc, sr, ga, params, keepUsingNamespace, keepUsingNamespace, append(path, self.Name)...)
			err = errors.Wrap(err, "Error encoding ComplexTypes.Sequence.Choice")
			if err != nil {
				return err
			}
		}
	}

	err = enc.EncodeToken(start.End())
	if err != nil {
		err = errors.Wrap(err, "Error encoding end token")
		return err
	}

	return nil
}
