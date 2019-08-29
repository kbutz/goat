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

func (e *Element) Encode(enc *xml.Encoder, sr SchemaRepository, ga GetAliaser, params map[string]interface{}, useNamespace, keepUsingNamespace bool, path ...string) error {
	// If minOccurs="0" and the current schema element was not submitted on the parameters, we don't need it
	// If a value for the current schema element was submitted, continue with encoding
	if e.MinOccurs == "0" && !hasPrefix(params, MakePath(append(path, e.Name))) {
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
			Local:  e.Name,
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
	if e.Type != "" {
		parts := strings.Split(e.Type, ":")
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
			err = schema.EncodeType(parts[1], enc, sr, params, keepUsingNamespace, keepUsingNamespace, append(path, e.Name)...)
			if err != nil {
				err = errors.Wrap(err, "Error encoding type for "+parts[1])
				return err
			}
		default:
			err = fmt.Errorf("malformed type '%s' in path %q", e.Type, path)
			return err
		}
	} else if e.ComplexTypes != nil {
		for _, element := range e.ComplexTypes.Sequence {
			err = element.Encode(enc, sr, ga, params, keepUsingNamespace, keepUsingNamespace, append(path, e.Name)...)
			if err != nil {
				err = errors.Wrap(err, "Error encoding ComplexTypes.Sequence.Elements")
				return err
			}
		}

		for _, element := range e.ComplexTypes.Choice {
			// TODO: We don't actually need to do any choice validation here, I think. e.Encode will attempt to encode
			//		a type once one is reached, which will either encode the simple type with no validation, or the
			//		complexType with the choice validations
			err = element.Encode(enc, sr, ga, params, keepUsingNamespace, keepUsingNamespace, append(path, e.Name)...)
			err = errors.Wrap(err, "Error encoding ComplexTypes.Sequence.Choice")
			if err != nil {
				return err
			}
		}

		for _, element := range e.ComplexTypes.SequenceChoice {
			// TODO: We don't actually need to do any choice validation here, I think. e.Encode will attempt to encode
			//		a type once one is reached, which will either encode the simple type with no validation, or the
			//		complexType with the choice validations
			err = element.Encode(enc, sr, ga, params, keepUsingNamespace, keepUsingNamespace, append(path, e.Name)...)
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
