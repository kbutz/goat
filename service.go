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

func (w *Webservice) UseHeader(header *http.Header) {
	if header != nil {
		w.client.Header = header
	}
}

func (w *Webservice) UseHistory() {
	w.client.UseHistory = true
	w.ClearHistory()
}

func (w *Webservice) UseClient(client *http.Client) {
	if client != nil {
		w.client.Client = client
	}
}

func (w *Webservice) ClearHistory() {
	w.client.History = []client.History{}
}

func (w *Webservice) IgnoreHistory() {
	w.client.UseHistory = false
	w.ClearHistory()
}

func (w *Webservice) GetLatestHistory() (history *client.History) {
	if w.client.UseHistory == false {
		return nil
	}

	return &w.client.History[len(w.client.History)-1]
}

func (w *Webservice) GetHistory() (history *[]client.History) {
	if w.client.UseHistory == false {
		return nil
	}

	return &w.client.History
}

// AddServices : Given a submitted url or urls, unmarshal the wsdl definitions and store the unmarshalled definitions in memory
// as "service". This will also fetch any additional imports on the WSDL
func (w *Webservice) AddServices(urls ...string) (err error) {
	for _, u := range urls {
		service := &wsdl.Definitions{
			Aliases:           make(map[string]string),
			ImportDefinitions: make(map[string]wsdl.Definitions),
		}
		err = service.GetService(&w.client, u)
		if err != nil {
			return
		}
		w.services[service.Service.Name] = service
	}

	return
}
