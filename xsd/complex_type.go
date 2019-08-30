package xsd

import (
	"errors"
	"fmt"
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
	// TODO: All   - Missing the <xs:all schema component, which specifies that the child elements can appear in any order.
	// TODO: Group - Missing <xs:group schema component (remove me - see DescribeVpcAttributesGroup)
	// TODO: Does not support choice>sequence>elements nested schemas
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

func (c *ComplexType) Encode(enc *xml.Encoder, sr SchemaRepository, ga GetAliaser, params map[string]interface{}, useNamespace, keepUsingNamespace bool, path ...string) error {
	for _, e := range c.Sequence {
		err := e.Encode(enc, sr, ga, params, useNamespace, keepUsingNamespace, path...)
		if err != nil {
			return err
		}
	}

	// TODO: After writing some tests, refactor this to use EncodeChoice
	// First, verify that one and only one of the choices for this path has been submitted on the params
	// If none, continue do not encode
	// If more than one, return error
	// If one, start encoding - if any of the child element types are also choices, they will need to meet the same criteria
	submittedChoices := 0
	for _, e := range c.Choice {
		if hasPrefix(params, MakePath(append(path, e.Name))) {
			submittedChoices++
		}
	}

	if submittedChoices > 1 {
		return errors.New("A max of one choice element can be submitted")
	}

	if submittedChoices == 1 {
		for _, e := range c.Choice {
			if hasPrefix(params, MakePath(append(path, e.Name))) {
				err := e.Encode(enc, sr, ga, params, useNamespace, keepUsingNamespace, path...)
				if err != nil {
					return err
				}
			}
		}
	}

	submittedSequenceChoices := 0
	for _, e := range c.SequenceChoice {
		if hasPrefix(params, MakePath(append(path, e.Name))) {
			submittedSequenceChoices++
		}
	}

	if submittedSequenceChoices > 1 {
		return errors.New("A max of one choice element can be submitted")
	}

	if submittedSequenceChoices == 1 {
		for _, e := range c.SequenceChoice {
			if hasPrefix(params, MakePath(append(path, e.Name))) {
				err := e.Encode(enc, sr, ga, params, useNamespace, keepUsingNamespace, path...)
				if err != nil {
					return err
				}
			}
		}
	}

	if c.Content != nil {
		parts := strings.Split(c.Content.Extension.Base, ":")
		switch len(parts) {
		case 2:
			var schema Schemaer
			schema, err := sr.GetSchema(ga.GetAlias(parts[0]))
			if err != nil {
				return err
			}

			err = schema.EncodeType(parts[1], enc, sr, params, keepUsingNamespace, keepUsingNamespace, path...)
			if err != nil {
				return err
			}
		default:
			err := fmt.Errorf("malformed base '%s' in path %q", c.Content.Extension.Base, path)
			return err
		}

		for _, e := range c.Content.Extension.Sequence {
			err := e.Encode(enc, sr, ga, params, useNamespace, keepUsingNamespace, path...)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *ComplexType) EncodeChoice(choiceElements []Element, enc *xml.Encoder, sr SchemaRepository, ga GetAliaser, params map[string]interface{}, useNamespace, keepUsingNamespace bool, path ...string) error {
	// First, verify that one and only one of the choices for this path has been submitted on the params
	// If none, continue do not encode
	// If more than one, return error
	// If one, start encoding - if any of the child element types are also choices, they will need to meet the same criteria
	submittedChoices := 0
	for _, e := range choiceElements {
		if hasPrefix(params, MakePath(append(path, e.Name))) {
			submittedChoices++
		}
	}

	if submittedChoices > 1 {
		return errors.New("A max of one choice element can be submitted")
	}

	if submittedChoices == 1 {
		for _, e := range choiceElements {
			if hasPrefix(params, MakePath(append(path, e.Name))) {
				err := e.Encode(enc, sr, ga, params, useNamespace, keepUsingNamespace, path...)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}
