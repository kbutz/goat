package goat

import (
	"github.com/kbutz/goat/client"
	"github.com/kbutz/goat/wsdl"
	"net/http"
)

type Webservice struct {
	services map[string]*wsdl.Definitions
	client   client.Client
}

func NewWebservice() Webservice {
	ws := Webservice{
		services: map[string]*wsdl.Definitions{},
		client:   client.Client{},
	}

	temp := make(http.Header)
	ws.client.Client = http.DefaultClient
	ws.client.Header = &temp
	ws.client.History = []client.History{}
	return ws
}

func (self *Webservice) UseHeader(header *http.Header) {
	if header != nil {
		self.client.Header = header
	}
}

func (self *Webservice) UseHistory() {
	self.client.UseHistory = true
	self.ClearHistory()
}

func (self *Webservice) UseClient(client *http.Client) {
	if client != nil {
		self.client.Client = client
	}
}

func (self *Webservice) ClearHistory() {
	self.client.History = []client.History{}
}

func (self *Webservice) IgnoreHistory() {
	self.client.UseHistory = false
	self.ClearHistory()
}

func (self *Webservice) GetLatestHistory() (history *client.History) {
	if self.client.UseHistory == false {
		return nil
	}

	return &self.client.History[len(self.client.History)-1]
}

func (self *Webservice) GetHistory() (history *[]client.History) {
	if self.client.UseHistory == false {
		return nil
	}

	return &self.client.History
}

func (self *Webservice) AddServices(urls ...string) (err error) {
	for _, u := range urls {
		service := &wsdl.Definitions{
			Aliases:           make(map[string]string),
			ImportDefinitions: make(map[string]wsdl.Definitions),
		}
		err = service.GetService(&self.client, u)
		if err != nil {
			return
		}
		self.services[service.Service.Name] = service
	}

	return
}
