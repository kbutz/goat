package xsd

import (
	"fmt"
	"github.com/pkg/errors"
	"strings"

	"github.com/sezzle/sezzle-go-xml"
)

type ComplexType struct {
	XMLName  xml.Name `xml:"http://www.w3.org/2001/XMLSchema complexType"`
	Name     string   `xml:"name,attr"`
	Abstract bool     `xml:"abstract,attr"`
	Sequence Sequence `xml:"sequence"`
	//Sequence []Element       `xml:"sequence>element"`
	//Sequence []Element        `xml:"sequence>element"`
	//Choice []Element 	     `xml:"choice>element"`
	//SequenceChoice []Element `xml:"sequence>choice>element"`
	Content *ComplexContent `xml:"http://www.w3.org/2001/XMLSchema complexContent"`
}

type Sequence struct {
	Elements []Element `xml:"element"`
	Choice   []Element `xml:"choice>element"`
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
	for _, e := range c.Sequence.Elements {
		err := e.Encode(enc, sr, ga, params, useNamespace, keepUsingNamespace, path...)
		if err != nil {
			err = errors.Wrap(err, "Error encoding Sequence.Element")
			return err
		}
	}

	// First, verify that one nad only one of the choices for this path has been submitted on the params
	// If none, continue do not encode
	// If more than one, return error
	// If one, start encoding - if any of the child element types are also choices, they will need to meet the same criteria
	submittedChoices := 0
	for _, e := range c.Sequence.Choice {
		if hasPrefix(params, MakePath(append(path, e.Name))) {
			submittedChoices++
		}
	}

	fmt.Println(fmt.Sprintf("submittedChoices: %+v for %s", submittedChoices, c.Name))

	//expectedChoiceErrors := len(c.Sequence.Choice) - 1
	//choiceErrorCount := 0
	//for _, e := range c.Sequence.Choice {
	//	fmt.Println("Encoding CHOICE: " + fmt.Sprintf("%+v", e))
	//	choiceErr := e.Encode(enc, sr, ga, params, useNamespace, keepUsingNamespace, path...)
	//	if choiceErr != nil {
	//		choiceErrorCount++
	//		fmt.Println("ERROR COUNT++ ", choiceErr)
	//	}
	//}
	//
	//// TODO: The last condition needs to go away. I think this library has just been disrespecting the choice element, allowing us to send malformed (but somehow passing) XML...
	//if len(c.Sequence.Choice) > 0 && choiceErrorCount != expectedChoiceErrors && choiceErrorCount != len(c.Sequence.Choice) {
	//	err := fmt.Errorf("choice error of %+v was not equal to expect choice error count of %+v or equal to choice count of %+v", choiceErrorCount, expectedChoiceErrors, len(c.Sequence.Choice))
	//	return err
	//}

	fmt.Println("c.Content: " + fmt.Sprintf("%+v", c.Content))
	if c.Content != nil {
		parts := strings.Split(c.Content.Extension.Base, ":")
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
			err := fmt.Errorf("malformed base '%s' in path %q", c.Content.Extension.Base, path)
			return err
		}

		for _, e := range c.Content.Extension.Sequence {
			err := e.Encode(enc, sr, ga, params, useNamespace, keepUsingNamespace, path...)
			if err != nil {
				err = errors.Wrap(err, "Error encoding from Content.Extension.Sequence")
				return err
			}
		}
	}

	return nil
}
