package goat

import (
	"bytes"
	"fmt"
	"io"

	"github.com/sezzle/sezzle-go-xml"
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

func (self *Webservice) NewRequest(service, method string, params map[string]interface{}, buf io.Writer) (err error) {
	s := self.services[service]
	if s == nil {
		err = fmt.Errorf("no such service '%s'", service)
		return
	}

	err = s.WriteRequest(method, buf, params)
	return
}

func (self *Webservice) SendBuffer(service string, res interface{}, buf io.Reader) (err error) {
	s := self.services[service]
	if s == nil {
		err = fmt.Errorf("no such service '%s'", service)
		return
	}

	e := new(ResponseEnvelope)
	err = self.client.MakeRequest("POST", s.Service.Port.Address.Location, buf, e)
	if err != nil {
		return
	}
	err = xml.Unmarshal(e.Body.Data, res)
	if err != nil {
		return
	}
	return
}

func (self *Webservice) Do(service, method string, res interface{}, params map[string]interface{}) (err error) {
	buf := new(bytes.Buffer)
	err = self.NewRequest(service, method, params, buf)
	if err != nil {
		return
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
	err = self.SendBuffer(service, res, buf)
	return err
}
