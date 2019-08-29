package client

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"

	"github.com/sezzle/sezzle-go-xml"
)

type History struct {
	Request      *http.Request
	RequestBody  *bytes.Buffer
	Response     *http.Response
	ResponseBody *bytes.Buffer
}

type Client struct {
	Client *http.Client
	Header *http.Header

	UseHistory bool
	History    []History
}

func (self *Client) MakeRequest(requestMethod, requestURL string, requestBody io.Reader, decodedResponse interface{}) (err error) {
	val := reflect.ValueOf(decodedResponse)
	if val.Kind() != reflect.Ptr {
		return errors.New("non-pointer decodedResponse passed to MakeRequest")
	}

	var hs History
	if self.UseHistory && requestBody != nil {
		var buf []byte
		buf, err = ioutil.ReadAll(requestBody)
		if err != nil {
			return
		}
		hs.RequestBody = bytes.NewBuffer(buf)
		requestBody = ioutil.NopCloser(bytes.NewBuffer(buf))
	}

	req, err := http.NewRequest(requestMethod, requestURL, requestBody)
	if err != nil {
		return
	}

	if self.Header != nil {
		req.Header = *self.Header
	}
	req.Header.Set("Content-Type", "application/soap+xml")

	if self.UseHistory {
		hs.Request = req
	}

	var resp *http.Response
	client := http.DefaultClient
	if self.Client != nil {
		client = self.Client
	}
	resp, err = client.Do(req)
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

	var responseBody io.Reader
	responseBody = resp.Body
	if self.UseHistory {
		hs.Response = resp
		hs.ResponseBody = &bytes.Buffer{}
		responseBody = io.TeeReader(responseBody, io.Writer(hs.ResponseBody))
		self.History = append(self.History, hs)
	}
	err = xml.NewDecoder(responseBody).Decode(decodedResponse)
	if err != nil {
		return
	}

	return
}
