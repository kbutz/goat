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
	Client HTTPClientDoer
	Header *http.Header

	UseHistory bool
	History    []History
}

type HTTPClientDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

func (c *Client) MakeRequest(requestMethod, requestURL string, requestBody io.Reader, decodedResponse interface{}) error {
	val := reflect.ValueOf(decodedResponse)
	if val.Kind() != reflect.Ptr {
		return errors.New("non-pointer decodedResponse passed to MakeRequest")
	}

	var hs History
	var err error
	if c.UseHistory && requestBody != nil {
		var buf []byte
		buf, err = ioutil.ReadAll(requestBody)
		if err != nil {
			return err
		}
		hs.RequestBody = bytes.NewBuffer(buf)
		requestBody = ioutil.NopCloser(bytes.NewBuffer(buf))
	}

	req, err := http.NewRequest(requestMethod, requestURL, requestBody)
	if err != nil {
		return err
	}

	if c.Header != nil {
		req.Header = *c.Header
	}
	req.Header.Set("Content-Type", "application/soap+xml")

	if c.UseHistory {
		hs.Request = req
	}

	var resp *http.Response
	//client := http.DefaultClient
	//if c.Client != nil {
	//	client = c.Client
	//}
	resp, err = c.Client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		var b []byte
		b, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		err = errors.New(string(b))
		return err
	}

	var responseBody io.Reader
	responseBody = resp.Body
	if c.UseHistory {
		hs.Response = resp
		hs.ResponseBody = &bytes.Buffer{}
		responseBody = io.TeeReader(responseBody, io.Writer(hs.ResponseBody))
		c.History = append(c.History, hs)
	}
	err = xml.NewDecoder(responseBody).Decode(decodedResponse)
	if err != nil {
		return err
	}

	return nil
}
