package goat

import (
	"bytes"
	"fmt"
	"github.com/sezzle/sezzle-go-xml"
	"io"
)

type ResponseEnvelope struct {
	XMLName xml.Name `xml:"Envelope"`
	Header  struct {
		XMLName xml.Name `xml:"Header"`
		Data    []byte   `xml:",innerxml"`
	}
	Body struct {
		XMLName xml.Name `xml:"Body"`
		Data    []byte   `xml:",innerxml"`
	}
}

func (w *Webservice) NewRequest(service, method string, params map[string]interface{}, buf io.Writer) error {
	s := w.services[service]
	if s == nil {
		err := fmt.Errorf("no such service '%s'", service)
		return err
	}

	err := s.WriteRequest(method, buf, params)
	return err
}

func (w *Webservice) SendBuffer(service string, res interface{}, buf io.Reader) error {
	s := w.services[service]
	if s == nil {
		err := fmt.Errorf("no such service '%s'", service)
		return err
	}

	e := new(ResponseEnvelope)
	err := w.client.MakeRequest("POST", s.Service.Port.Address.Location, buf, e)
	if err != nil {
		return err
	}
	err = xml.Unmarshal(e.Body.Data, res)
	if err != nil {
		return err
	}
	return nil
}

func (w *Webservice) Do(service, method string, res interface{}, params map[string]interface{}) error {
	buf := new(bytes.Buffer)
	err := w.NewRequest(service, method, params, buf)
	if err != nil {
		return err
	}

	//fmt.Println("THE ERROR: ", err)
	//var byteSlice []byte
	//byteSlice, err = ioutil.ReadAll(buf)
	//fmt.Println("THE BUFFER: " + string(byteSlice))
	//if err != nil {
	//	return
	//}
	//
	//return nil

	// TODO: Webservice.SendBuffer with invalid SOAP request will return a generic SOAP faultCode
	err = w.SendBuffer(service, res, buf)
	return err
}
