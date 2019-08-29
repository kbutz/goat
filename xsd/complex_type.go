package xsd

import (
	"fmt"
	"github.com/pkg/errors"
	"strings"

	"github.com/sezzle/sezzle-go-xml"
)

type ComplexType struct {
	XMLName        xml.Name        `xml:"http://www.w3.org/2001/XMLSchema complexType"`
	Name           string          `xml:"name,attr"`
	Abstract       bool            `xml:"abstract,attr"`
	Sequence       []Element       `xml:"sequence>element"`        // Specifies that all of the elements the child elements must appear in a sequence 0 to any number of times
	Choice         []Element       `xml:"choice>element"`          // Allows only one or zero of the elements contained int the declaration to be present within the containing element
	SequenceChoice []Element       `xml:"sequence>choice>element"` // Allows only one or zero of the elements contained int the declaration to be present within the containing element
	Content        *ComplexContent `xml:"http://www.w3.org/2001/XMLSchema complexContent"`
	// TODO: Missing the <xs:all schema component, which specifies that the child elements can appear in any order.
	// NOTE: Does not support choice>sequence>elements nested schemas
}

type ComplexContent struct {
	XMLName   xml.Name  `xml:"http://www.w3.org/2001/XMLSchema complexContent"`
	Extension Extension `xml:"http://www.w3.org/2001/XMLSchema extension"`
}

type Extension struct {
	XMLName  xml.Name  `xml:"http://www.w3.org/2001/XMLSchema extension"`
	Base     string    `xml:"base,attr"`
	Sequence []Element `xml:"sequence>element"`
}

func (self *ComplexType) Encode(enc *xml.Encoder, sr SchemaRepository, ga GetAliaser, params map[string]interface{}, useNamespace, keepUsingNamespace bool, path ...string) error {
	for _, e := range self.Sequence {
		err := e.Encode(enc, sr, ga, params, useNamespace, keepUsingNamespace, path...)
		if err != nil {
			err = errors.Wrap(err, "Error encoding Sequence.Element")
			return err
		}
	}

	// First, verify that one and only one of the choices for this path has been submitted on the params
	// If none, continue do not encode
	// If more than one, return error
	// If one, start encoding - if any of the child element types are also choices, they will need to meet the same criteria
	submittedChoices := 0
	for _, e := range self.Choice {
		if hasPrefix(params, MakePath(append(path, e.Name))) {
			submittedChoices++
		}
	}

	fmt.Println(fmt.Sprintf("submittedChoices: %+v for %s", submittedChoices, self.Name))

	if submittedChoices > 1 {
		return errors.New("A max of one choice element can be submitted")
	}

	if submittedChoices == 1 {
		fmt.Println("There was a single choice block submitted on the params, attempt to encode the choice elements")
		for _, e := range self.Choice {
			if hasPrefix(params, MakePath(append(path, e.Name))) {
				fmt.Println("Encoding CHOICE: " + fmt.Sprintf("%+v", e))
				err := e.Encode(enc, sr, ga, params, useNamespace, keepUsingNamespace, path...)
				if err != nil {
					err = errors.Wrap(err, "Error encoding Sequence.Choice")
					return err
				}
			}
		}
	}

	submittedSequenceChoices := 0
	for _, e := range self.SequenceChoice {
		if hasPrefix(params, MakePath(append(path, e.Name))) {
			submittedSequenceChoices++
		}
	}

	fmt.Println(fmt.Sprintf("submittedSequenceChoices: %+v for %s", submittedSequenceChoices, self.Name))

	if submittedSequenceChoices > 1 {
		return errors.New("A max of one choice element can be submitted")
	}

	if submittedSequenceChoices == 1 {
		fmt.Println("There was a single choice block submitted on the params, attempt to encode the choice elements")
		for _, e := range self.SequenceChoice {
			if hasPrefix(params, MakePath(append(path, e.Name))) {
				fmt.Println("Encoding SEQUENCE>CHOICE: " + fmt.Sprintf("%+v", e))
				err := e.Encode(enc, sr, ga, params, useNamespace, keepUsingNamespace, path...)
				if err != nil {
					err = errors.Wrap(err, "Error encoding Sequence.Choice")
					return err
				}
			}
		}
	}

	fmt.Println("self.Content: " + fmt.Sprintf("%+v", self.Content))
	if self.Content != nil {
		parts := strings.Split(self.Content.Extension.Base, ":")
		switch len(parts) {
		case 2:
			var schema Schemaer
			schema, err := sr.GetSchema(ga.GetAlias(parts[0]))
			if err != nil {
				err = errors.Wrap(err, "Error getting schema from "+parts[0])
				return err
			}

			err = schema.EncodeType(parts[1], enc, sr, params, keepUsingNamespace, keepUsingNamespace, path...)
			if err != nil {
				err = errors.Wrap(err, "Error encoding type for "+parts[1])
				return err
			}
		default:
			err := fmt.Errorf("malformed base '%s' in path %q", self.Content.Extension.Base, path)
			return err
		}

		for _, e := range self.Content.Extension.Sequence {
			err := e.Encode(enc, sr, ga, params, useNamespace, keepUsingNamespace, path...)
			if err != nil {
				err = errors.Wrap(err, "Error encoding from Content.Extension.Sequence")
				return err
			}
		}
	}

	return nil
}
