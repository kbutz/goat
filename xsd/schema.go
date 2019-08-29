package xsd

import (
	"fmt"

	"github.com/sezzle/sezzle-go-xml"
)

type InnerSchema struct {
	TargetNamespace    string        `xml:"targetNamespace,attr"`
	ElementFormDefault string        `xml:"elementFormDefault,attr"`
	Version            string        `xml:"version,attr"`
	ComplexTypes       []ComplexType `xml:"http://www.w3.org/2001/XMLSchema complexType"`
	SimpleTypes        []SimpleType  `xml:"http://www.w3.org/2001/XMLSchema simpleType"`
	Elements           []Element     `xml:"http://www.w3.org/2001/XMLSchema element"`
	// TODO: need to handle Choice schema type
	//		https://www.w3schools.com/xml/el_choice.asp
	//		https://medium.com/eaciit-engineering/soap-wsdl-request-in-go-language-3861cfb5949e
}

type Schema struct {
	XMLName xml.Name `xml:"http://www.w3.org/2001/XMLSchema schema"`
	Aliases map[string]string
	InnerSchema
}

func (self *Schema) UnmarshalXML(d *xml.Decoder, start xml.StartElement) (err error) {
	err = d.DecodeElement(&self.InnerSchema, &start)
	if err != nil {
		return
	}

	self.XMLName = start.Name
	self.Aliases = map[string]string{}

	for _, attr := range start.Attr {
		self.Aliases[attr.Name.Local] = attr.Value
	}
	return
}

func (self *Schema) Namespace() string {
	return self.TargetNamespace
}

func (self *Schema) GetAlias(alias string) (space string) {
	return self.Aliases[alias]
}

// EncodeElement : Begins encoding to XML from the top level body element, calling Encode and EncodeType recursively on the
// nested elements until there are no more to be encoded.
func (self *Schema) EncodeElement(name string, enc *xml.Encoder, sr SchemaRepository, params map[string]interface{}, useNamespace, keepUsingNamespace bool, path ...string) error {
	// Starts encoding the top level xml element
	//fmt.Println(fmt.Sprintf("Elements: %+v", self.Elements))
	for _, elem := range self.Elements {
		if elem.Name == name {
			fmt.Println("elem.Name == name: " + fmt.Sprintf("%+v", elem))
			// elem.Name == "transaction-continue" or "transaction-identity-verification", for example
			err := elem.Encode(enc, sr, self, params, useNamespace, keepUsingNamespace, path...)
			if err != nil {
				fmt.Println(err)
				return err
			}
			return nil
		}
	}

	return fmt.Errorf("did not find element '%s'", name)
}

func (self *Schema) EncodeType(name string, enc *xml.Encoder, sr SchemaRepository, params map[string]interface{}, useNamespace, keepUsingNamespace bool, path ...string) error {
	for _, cmplx := range self.ComplexTypes {
		if cmplx.Name == name {
			fmt.Println("cmplx.Name == name, cmplx: " + fmt.Sprintf("%+v", cmplx))
			return cmplx.Encode(enc, sr, self, params, useNamespace, keepUsingNamespace, path...)
		}
	}

	for _, smpl := range self.SimpleTypes {
		if smpl.Name == name {
			fmt.Println("smpl.Name == name, cmplx: " + fmt.Sprintf("%+v", smpl))
			return smpl.Encode(enc, sr, self, params, useNamespace, keepUsingNamespace, path...)
		}
	}

	return fmt.Errorf("did not find type '%s'", name)
}
