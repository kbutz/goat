package goat

import (
	"bytes"
	"github.com/golang/mock/gomock"
	"github.com/kbutz/goat/client"
	"github.com/kbutz/goat/wsdl"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestWebservice_NewRequest(t *testing.T) {
	//data, err := ioutil.ReadFile("fixtures/ec2.wsdl")
	//if err != nil {
	//	t.Errorf("Error reading wsdl file")
	//}
	mockClientController := gomock.NewController(t)
	defer mockClientController.Finish()

	mockClient := client.NewMockHTTPClientDoer(mockClientController)
	mockClient.EXPECT().Do(gomock.Any()).Return(&http.Response{StatusCode: http.StatusInternalServerError, Body: ioutil.NopCloser(bytes.NewBuffer([]byte("")))}, nil).Times(1)

	testService := Webservice{
		services: map[string]*wsdl.Definitions{},
		client:   client.Client{Client: mockClient},
	}

	err := testService.AddServices("http://mocked.com/ws?WSDL")
	if err == nil {
		t.Errorf("Expected err to not be nil")
	}
}
