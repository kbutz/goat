package wsdl

import (
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/sezzle/sezzle-go-xml"

	"github.com/kbutz/goat/client"
	"github.com/kbutz/goat/xsd"
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

func (d *Definitions) GetNamespace(alias string) string {
	return d.Aliases[alias]
}

func (d *Definitions) GetAlias(namespace string) string {
	for key, val := range d.Aliases {
		if val == namespace {
			return key
		}
	}
	return ""
}

func (d *Definitions) UnmarshalXML(decoder *xml.Decoder, start xml.StartElement) error {
	err := decoder.DecodeElement(&d.InnerDefinitions, &start)
	if err != nil {
		return err
	}

	d.XMLName = start.Name
	d.Aliases = map[string]string{}

	d.Types.Schemas = xsd.SchemaMap{}
	for _, schema := range d.Types.Schemata {
		d.Types.Schemas[schema.TargetNamespace] = schema
	}

	for _, attr := range start.Attr {
		if _, ok := d.Aliases[attr.Name.Local]; !ok {
			d.Aliases[attr.Name.Local] = attr.Value
		}

		for k := range d.Types.Schemas {
			if _, ok := d.Types.Schemas[k].Aliases[attr.Name.Local]; !ok {
				d.Types.Schemas[k].Aliases[attr.Name.Local] = attr.Value
			}
		}
	}

	return nil
}

func copyMap(src map[string]interface{}) map[string]interface{} {
	dst := map[string]interface{}{}
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func (d *Definitions) WriteRequest(operation string, w io.Writer, bodyParams map[string]interface{}) error {
	//headerParams = copyMap(headerParams)
	bodyParams = copyMap(bodyParams)

	var bndOp BindingOperation
	var ptOp PortTypeOperation
	var err error
	bndOp, ptOp, err = d.getOperations(operation)
	if err != nil {
		return err
	}
	// fmt.Println("bndOp", bndOp)
	// fmt.Println("ptOp", ptOp)

	var body xsd.Schema
	var bodyElement string
	var bodyService *Definitions
	// TODO: implement proper handling, tho I can't really find a SoapHeader part for the binding operation
	/*
		header, headerElement, err = d.getSchema(bndOp.Input.SoapHeader.PortTypeOperationMessage)
		if err != nil {
			return
		}
	*/

	body, bodyElement, bodyService, err = d.getSchema(bndOp.Input.SoapBody.PortTypeOperationMessage, ptOp.Input)
	if err != nil {
		return err
	}

	_, err = fmt.Fprint(w, xml.Header)
	if err != nil {
		return err
	}
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

	err = enc.EncodeToken(envelope)
	if err != nil {
		return err
	}
	defer func() {
		_ = enc.EncodeToken(envelope.End())
	}()

	// soapHeader := xml.StartElement{
	// 	Name: xml.Name{
	// 		Space: "http://schemas.xmlsoap.org/soap/envelope/",
	// 		Local: "Header",
	// 	},
	// }
	//enc.EncodeToken(soapHeader)

	// err = header.EncodeElement(headerElement, enc, d.Types.Schemas, headerParams)
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
	err = enc.EncodeToken(soapBody)
	if err != nil {
		return err
	}

	err = body.EncodeElement(bodyElement, enc, bodyService.Types.Schemas, bodyParams, true, false)
	if err != nil {
		return err
	}

	err = enc.EncodeToken(soapBody.End())
	if err != nil {
		return err
	}

	return nil
}

func (d *Definitions) getSchema(msg ...PortTypeOperationMessage) (schema xsd.Schema, element string, service *Definitions, err error) {
	for _, s := range msg {
		service = d
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

func (d *Definitions) getOperations(operation string) (bndOp BindingOperation, ptOp PortTypeOperation, err error) {
	service := *d
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
		err = fmt.Errorf("malformed binding information: '%s'", d.Service.Port.Binding)
	}

	return
}

// Unmarhsals the WSDL definitions into the Definitions struct
func (d *Definitions) GetDefinitions(client *client.Client, url string) error {
	return client.MakeRequest("GET", url, nil, d)
}

// Gets the base wsdl import, binding and operation definitions, adds imports and schema definitions
func (d *Definitions) GetService(client *client.Client, url string) error {
	err := d.GetDefinitions(client, url)
	if err != nil {
		return err
	}

	if d.Service.Name == "" {
		err = fmt.Errorf("invalid service name '%s' for url '%s'", d.Service.Name, url)
		return err
	}
	log.Printf("adding service '%s' from '%s'", d.Service.Name, url)

	log.Printf("adding all imports")
	err = d.AddImports(client)
	if err != nil {
		return err
	}

	return nil
}

// AddImports : Gets wsdl schema definitions and recursively adds any additional imports - for example, if the
// WSDL itself has an import to fetch the type definitions separately from the bindings and operations
func (d *Definitions) AddImports(client *client.Client) error {
	imports := []Import{}
	for _, val := range d.Imports {
		imports = append(imports, val)
	}

	for i := range imports {
		if _, ok := d.ImportDefinitions[d.GetAlias(imports[i].Namespace)]; ok {
			log.Printf("skipping import from '%s', already added", imports[i].Location)
			continue
		}

		log.Printf("adding import from '%s'", imports[i].Location)
		definitions := &Definitions{
			Aliases:           make(map[string]string),
			ImportDefinitions: make(map[string]Definitions),
		}

		err := definitions.GetDefinitions(client, imports[i].Location)
		if err != nil {
			return err
		}

		err = definitions.AddImports(client)
		if err != nil {
			return err
		}

		d.ImportDefinitions[d.GetAlias(imports[i].Namespace)] = *definitions
	}

	return nil
}
