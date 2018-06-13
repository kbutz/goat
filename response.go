package goat

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"os"
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

	err = s.WriteRequest(method, buf, self.header, params)
	return
}

func (self *Webservice) SendBuffer(service string, res interface{}, buf io.Reader) (err error) {
	fmt.Println("HELLO")
	s := self.services[service]
	if s == nil {
		err = fmt.Errorf("no such service '%s'", service)
		return
	}

	var resp *http.Response
	req, err := http.NewRequest("POST", s.Service.Port.Address.Location, buf)
	if err != nil {
		return
	}
	for key, val := range self.header {
		req.Header.Set(key, val.(string))
	}
	req.Header.Set("Content-Type", "application/soap+xml")

	bts, _ := httputil.DumpRequest(req, true)
	fmt.Println("REQUEST:", string(bts))

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var b []byte
		b, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return
		}

		err = errors.New(string(b))
		return
	}

	e := new(ResponseEnvelope)
	err = xml.NewDecoder(io.TeeReader(resp.Body, os.Stdout)).Decode(e)
	if err != nil {
		return
	}

	err = xml.Unmarshal(e.Body.Data, res)
	return
}

func (self *Webservice) Do(service, method string, res interface{}, params map[string]interface{}) (err error) {
	buf := new(bytes.Buffer)
	err = self.NewRequest(service, method, params, buf)
	if err != nil {
		return
	}

	fmt.Println("BYTES", []rune(string(buf.Bytes())))

	err = self.SendBuffer(service, res, buf)
	return
}
