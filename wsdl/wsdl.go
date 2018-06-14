package wsdl

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/sezzle/sezzle-go-xml"

	"sezzle/goat/xsd"
)

type InnerDefinitions struct {
	TargetNamespace string `xml:"targetNamespace,attr"`

	Imports  []Import  `xml:"import"`
	Types    Type      `xml:"types"`
	Messages []Message `xml:"message"`
	PortType PortType  `xml:"portType"`
	Binding  []Binding `xml:"binding"`
	Service  Service   `xml:"service"`
}

type Definitions struct {
	XMLName           xml.Name               `xml:"definitions"`
	Aliases           map[string]string      `xml:"-"` // mapping of [alias]namespace
	ImportDefinitions map[string]Definitions `xml:"-"` // mapping of [alias]definitions
	InnerDefinitions
}

func (self *Definitions) GetNamespace(alias string) (space string) {
	return self.Aliases[alias]
}

func (self *Definitions) GetAlias(namespace string) (alias string) {
	for key, val := range self.Aliases {
		if val == namespace {
			return key
		}
	}
	return ""
}

func (self *Definitions) UnmarshalXML(decoder *xml.Decoder, start xml.StartElement) (err error) {
	err = decoder.DecodeElement(&self.InnerDefinitions, &start)
	if err != nil {
		return
	}

	self.XMLName = start.Name
	self.Aliases = map[string]string{}

	self.Types.Schemas = xsd.SchemaMap{}
	for _, schema := range self.Types.Schemata {
		self.Types.Schemas[schema.TargetNamespace] = schema
	}

	for _, attr := range start.Attr {
		if _, ok := self.Aliases[attr.Name.Local]; !ok {
			self.Aliases[attr.Name.Local] = attr.Value
		}

		for k := range self.Types.Schemas {
			if _, ok := self.Types.Schemas[k].Aliases[attr.Name.Local]; !ok {
				self.Types.Schemas[k].Aliases[attr.Name.Local] = attr.Value
			}
		}
	}

	return
}

func copyMap(src map[string]interface{}) map[string]interface{} {
	dst := map[string]interface{}{}
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func (self *Definitions) WriteRequest(operation string, w io.Writer, headerParams, bodyParams map[string]interface{}) (err error) {
	headerParams = copyMap(headerParams)
	bodyParams = copyMap(bodyParams)

	var bndOp BindingOperation
	var ptOp PortTypeOperation
	bndOp, ptOp, err = self.getOperations(operation)
	if err != nil {
		return
	}
	// fmt.Println("bndOp", bndOp)
	// fmt.Println("ptOp", ptOp)

	var body xsd.Schema
	var bodyElement string
	var bodyService *Definitions
	// TODO: implement proper handling, tho I can't really find a SoapHeader part for the binding operation
	/*
		header, headerElement, err = self.getSchema(bndOp.Input.SoapHeader.PortTypeOperationMessage)
		if err != nil {
			return
		}
	*/

	body, bodyElement, bodyService, err = self.getSchema(bndOp.Input.SoapBody.PortTypeOperationMessage, ptOp.Input)
	if err != nil {
		return
	}

	fmt.Fprint(w, xml.Header)
	enc := xml.NewEncoder(w)
	//enc := xml.NewEncoder(io.MultiWriter(w, os.Stdout))
	enc.Indent("", "  ")
	defer func() {
		if err == nil {
			err = enc.Flush()
		}
	}()

	envName := "soap-env"
	envelope := xml.StartElement{
		Name: xml.Name{
			Space:  "http://schemas.xmlsoap.org/soap/envelope/",
			Prefix: envName,
			Local:  "Envelope",
		},
	}

	/*type test struct {
		Name   xml.Name
		Prefix string `xml:"xmlns:soapenv,attr"`
	}

	envelope := test{
		Name: xml.Name{
			Space: "ada",
			Local: envName + ":" + "Envelope",
		},
		Prefix: "http://schemas.xmlsoap.org/soap/envelope/",
	}

	op, _ := xml.MarshalIndent(envelope, "  ", "    ")
	fmt.Println(string(op))*/

	enc.EncodeToken(envelope)
	defer enc.EncodeToken(envelope.End())

	// soapHeader := xml.StartElement{
	// 	Name: xml.Name{
	// 		Space: "http://schemas.xmlsoap.org/soap/envelope/",
	// 		Local: "Header",
	// 	},
	// }
	//enc.EncodeToken(soapHeader)

	// err = header.EncodeElement(headerElement, enc, self.Types.Schemas, headerParams)
	// if err != nil {
	// 	return
	// }
	//enc.EncodeToken(soapHeader.End())

	soapBody := xml.StartElement{
		Name: xml.Name{
			Prefix: envName,
			Local:  "Body",
		},
	}
	enc.EncodeToken(soapBody)

	err = body.EncodeElement(bodyElement, enc, bodyService.Types.Schemas, bodyParams, true, false)
	if err != nil {
		return
	}
	enc.EncodeToken(soapBody.End())

	return
}

func (self *Definitions) getSchema(msg ...PortTypeOperationMessage) (schema xsd.Schema, element string, service *Definitions, err error) {
	for _, s := range msg {
		service = self
		if s.Message == "" {
			continue
		}

		parts := strings.Split(s.Message, ":")
		if len(parts) == 2 {
			if service.GetNamespace(parts[0]) != service.TargetNamespace {
				if _, ok := service.ImportDefinitions[parts[0]]; !ok {
					err = fmt.Errorf("cannot find '%s' namespace", parts[0])
					return
				}
				temp := service.ImportDefinitions[parts[0]]
				service = &temp
			}

			element = parts[1]
			var ok bool
			schema, ok = service.Types.Schemas[service.GetNamespace(parts[0])]
			if ok {
				for _, m := range service.Messages {
					if m.Name == element {
						p := strings.Split(m.Part.Element, ":")
						if len(p) != 2 {
							err = fmt.Errorf("invalid message part element name '%s'", m.Part.Element)
							return
						}

						element = p[1]
						return
					}
				}

				err = fmt.Errorf("did not find message '%s'", element)
				return
			}
		} else {
			err = fmt.Errorf("invalid soapheader message format '%s'", s.Message)
		}
	}

	err = fmt.Errorf("did not find schema in %q", msg)
	return
}

func (self *Definitions) getOperations(operation string) (bndOp BindingOperation, ptOp PortTypeOperation, err error) {
	service := *self
	parts := strings.Split(service.Service.Port.Binding, ":")
	switch len(parts) {
	case 2:
		if service.GetNamespace(parts[0]) != service.TargetNamespace {
			err = fmt.Errorf("have '%s', want '%s' as target namespace", parts[0], service.TargetNamespace)
			return
		}

		parts[0] = parts[1]
		fallthrough
	case 1:
		for _, bnd := range service.Binding {
			if bnd.Name == parts[0] {
				parts = strings.Split(bnd.Type, ":")

				switch len(parts) {
				case 2:
					if service.GetNamespace(parts[0]) != service.TargetNamespace {
						if _, ok := service.ImportDefinitions[parts[0]]; !ok {
							err = fmt.Errorf("cannot find '%s' namespace in binding %s", parts[0], bnd.Name)
							return
						}
						service = service.ImportDefinitions[parts[0]]
					}
					parts[0] = parts[1]
					fallthrough
				case 1:
					if service.PortType.Name != parts[0] {
						err = fmt.Errorf("have '%s', want '%s' as target namespace in binding '%s'", parts[0], service.PortType.Name, bnd.Name)
						return
					}

					var found bool
					for _, ptOp = range service.PortType.Operations {
						found = ptOp.Name == operation
						if found {
							break
						}
					}

					if !found {
						err = fmt.Errorf("did not find porttype operation '%s' in binding '%s'", operation, bnd.Name)
						return
					}
				default:
					err = fmt.Errorf("malformed binding information '%s' in binding '%s'", bnd.Type, bnd.Name)
					return
				}

				for _, bndOp = range bnd.Operations {
					if bndOp.Name == operation {
						return
					}
				}

				err = fmt.Errorf("did not find operation '%s' in binding '%s'", operation, bnd.Name)
				return
			}
		}

		err = fmt.Errorf("did not find binding '%s'", parts[0])
	default:
		err = fmt.Errorf("malformed binding information: '%s'", self.Service.Port.Binding)
	}

	return
}

func (self *Definitions) GetDefinitions(url string, headers map[string]interface{}) (err error) {
	var resp *http.Response
	var req *http.Request
	req, err = http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}

	for key, val := range headers {
		req.Header.Set(key, val.(string))
	}

	// bts, _ := httputil.DumpRequest(req, true)
	// fmt.Println(string(bts))

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	// bts, _ = httputil.DumpResponse(resp, true)
	// fmt.Println(string(bts))

	err = xml.NewDecoder(resp.Body).Decode(self)
	if err != nil {
		return
	}
	return
}

func (self *Definitions) GetService(url string, headers map[string]interface{}) (err error) {
	err = self.GetDefinitions(url, headers)
	if err != nil {
		return
	}

	if self.Service.Name == "" {
		err = fmt.Errorf("invalid service name '%s' for url '%s'", self.Service.Name, url)
		return
	}
	log.Printf("adding service '%s' from '%s'", self.Service.Name, url)

	log.Printf("adding all imports")
	err = self.AddImports(headers)
	if err != nil {
		return
	}

	return
}

func (self *Definitions) AddImports(headers map[string]interface{}) (err error) {
	imports := []Import{}
	for _, val := range self.Imports {
		imports = append(imports, val)
	}

	for i := range imports {
		if _, ok := self.ImportDefinitions[self.GetAlias(imports[i].Namespace)]; ok {
			log.Printf("skipping import from '%s', already added", imports[i].Location)
			continue
		}

		log.Printf("adding import from '%s'", imports[i].Location)
		definitions := &Definitions{
			Aliases:           make(map[string]string),
			ImportDefinitions: make(map[string]Definitions),
		}
		err = definitions.GetDefinitions(imports[i].Location, headers)
		if err != nil {
			return
		}

		err = definitions.AddImports(headers)
		if err != nil {
			return
		}

		self.ImportDefinitions[self.GetAlias(imports[i].Namespace)] = *definitions
	}

	return
}
