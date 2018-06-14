package goat

import (
	"sezzle/goat/wsdl"
)

type Webservice struct {
	services map[string]*wsdl.Definitions
	header   map[string]interface{}
}

func NewWebservice(header map[string]interface{}) Webservice {
	return Webservice{
		services: map[string]*wsdl.Definitions{},
		header:   header,
	}
}

func (self *Webservice) AddServices(urls ...string) (err error) {
	for _, u := range urls {
		service := &wsdl.Definitions{
			Aliases:           make(map[string]string),
			ImportDefinitions: make(map[string]wsdl.Definitions),
		}
		err = service.GetService(u, self.header)
		if err != nil {
			return
		}
		self.services[service.Service.Name] = service
	}

	return
}
