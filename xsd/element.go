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
	// If minOccurs="0" and the current schema element was not submitted on the parameters, we don't need it
	// If a value for the current schema element was submitted, continue with encoding
	if self.MinOccurs == "0" && !hasPrefix(params, MakePath(append(path, self.Name))) {
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
	//fmt.Println("*Element: " + fmt.Sprintf("%+v", self))
	//fmt.Println("Name: " + fmt.Sprintf("%+v", self.Name))
	//fmt.Println("Type: " + fmt.Sprintf("%+v", self.Type))
	//fmt.Println("ComplexTypes: " + fmt.Sprintf("%+v", self.ComplexTypes))
	//if self.ComplexTypes != nil {
	//	fmt.Println("ComplexTypes.Sequence" + fmt.Sprintf("%+v", self.ComplexTypes.Sequence))
	//}

	// If we've reached a an element with a Type, try to encode the type.
	// EncodeType will get the cached schema definition from self.Definitions and attempt to encode the type
	// based on the complexType or simpleType schema definition it has stored.
	// If the current element itself is an empty ComplexType tag, recursively call Encode until all elements have been encoded
	// TODO: For complexTypes with sequence>choice>element and choice>element, we need special handling since we don't want
	//		to force *all* of the elements to be encoded or abort. Here, we want *at least one* element encoded...
	//		This gets a little tricky with how this implementation handles encoding the XML - we need a way to abort/rollback a
	//		choice element that could not be encoded...
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

			// This isn't a helpful debug statement - the schema is the FULL wsdl definition map
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

		//submittedChoices := 0
		//for _, e := range self.ComplexTypes.Sequence.Choice {
		//	if hasPrefix(params, MakePath(append(path, e.Name))) {
		//		submittedChoices++
		//	}
		//}
		//
		//fmt.Println(fmt.Sprintf("submittedChoices: %+v for %s", submittedChoices, self.Name))

		for _, e := range self.ComplexTypes.Sequence.Choice {
			// First, verify that one nad only one of the choices for this path has been submitted on the params
			// If none, continue do not encode
			// If more than one, return error
			// If one, start encoding - if any of the child element types are also choices, they will need to meet the same criteria
			// or abort with the error.

			err = e.Encode(enc, sr, ga, params, keepUsingNamespace, keepUsingNamespace, append(path, self.Name)...)
			err = errors.Wrap(err, "Error encoding ComplexTypes.Sequence.Choice")
			if err != nil {
				return err
			}
		}
	}

	// If an error was thrown above while trying to add a choice element that is not required, we won't close the tag here
	err = enc.EncodeToken(start.End())
	if err != nil {
		err = errors.Wrap(err, "Error encoding end token")
		return err
	}

	return nil
}
