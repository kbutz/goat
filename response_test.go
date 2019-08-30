package goat

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/kbutz/goat/client"
	"github.com/kbutz/goat/wsdl"
)

func TestWebservice_Do_ErrorPath_NoServiceFound(t *testing.T) {
	testWSDL, err := ioutil.ReadFile("fixtures/ec2.wsdl")
	if err != nil {
		t.Errorf("Error reading wsdl file")
	}

	mockClientController := gomock.NewController(t)
	defer mockClientController.Finish()

	mockClient := client.NewMockHTTPClientDoer(mockClientController)

	mockClient.EXPECT().Do(gomock.Any()).Return(&http.Response{StatusCode: http.StatusOK, Body: ioutil.NopCloser(bytes.NewBuffer(testWSDL))}, nil).Times(1)

	testService := Webservice{
		services: map[string]*wsdl.Definitions{},
		client:   client.Client{Client: mockClient},
	}

	err = testService.AddServices("http://mocked.com/ws?WSDL")
	if err != nil {
		t.Errorf("Expected err to be nil")
	}

	params := map[string]interface{}{
		"UnassignPrivateIpAddresses/networkInterfaceId": "1234512345",
	}

	var responseString string
	err = testService.Do("NoServiceFound", "UnassignPrivateIpAddresses", &responseString, params)
	if err == nil {
		t.Errorf("Expected an error from testService.Do for no service found")
	}

	if err != nil && !strings.Contains(err.Error(), "no such service") {
		t.Errorf("Unexpected error message for no such service %+v", err)
	}
}

func TestWebservice_Do_ErrorPath_NoMethodFound(t *testing.T) {
	testWSDL, err := ioutil.ReadFile("fixtures/ec2.wsdl")
	if err != nil {
		t.Errorf("Error reading wsdl file")
	}

	mockClientController := gomock.NewController(t)
	defer mockClientController.Finish()

	mockClient := client.NewMockHTTPClientDoer(mockClientController)

	mockClient.EXPECT().Do(gomock.Any()).Return(&http.Response{StatusCode: http.StatusOK, Body: ioutil.NopCloser(bytes.NewBuffer(testWSDL))}, nil).Times(1)

	testService := Webservice{
		services: map[string]*wsdl.Definitions{},
		client:   client.Client{Client: mockClient},
	}

	err = testService.AddServices("http://mocked.com/ws?WSDL")
	if err != nil {
		t.Errorf("Expected err to be nil")
	}

	params := map[string]interface{}{
		"UnassignPrivateIpAddresses/networkInterfaceId": "1234512345",
	}

	var responseString string
	err = testService.Do("AmazonEC2", "UnknownMethod", &responseString, params)
	if err == nil {
		t.Errorf("Expected an error from testService.Do for unknown method")
	}

	if err != nil && !strings.Contains(err.Error(), "did not find porttype operation") {
		t.Errorf("Unexpected error message for unknown method %+v", err)
	}
}

func TestWebservice_Do_ErrorPath_InvalidRequestParams(t *testing.T) {
	testWSDL, err := ioutil.ReadFile("fixtures/ec2.wsdl")
	if err != nil {
		t.Errorf("Error reading wsdl file")
	}

	mockClientController := gomock.NewController(t)
	defer mockClientController.Finish()

	mockClient := client.NewMockHTTPClientDoer(mockClientController)

	mockClient.EXPECT().Do(gomock.Any()).Return(&http.Response{StatusCode: http.StatusOK, Body: ioutil.NopCloser(bytes.NewBuffer(testWSDL))}, nil).Times(1)

	testService := Webservice{
		services: map[string]*wsdl.Definitions{},
		client:   client.Client{Client: mockClient},
	}

	err = testService.AddServices("http://mocked.com/ws?WSDL")
	if err != nil {
		t.Errorf("Expected err to be nil")
	}

	params := map[string]interface{}{
		"UnassignPrivateIpAddresses/networkInterfaceId": "1234512345",
	}

	var responseString string
	err = testService.Do("AmazonEC2", "UnassignPrivateIpAddresses", &responseString, params)
	if err == nil {
		t.Errorf("Expected an error from testService.Do for invalid or missing sequence params")
	}

	if err != nil && !strings.Contains(err.Error(), "did not find data") {
		t.Errorf("Unexpected error message for missing required input params %+v", err)
	}
}

func TestWebservice_Do_HappyPathSimple(t *testing.T) {
	testWSDL, err := ioutil.ReadFile("fixtures/ec2.wsdl")
	if err != nil {
		t.Errorf("Error reading wsdl file")
	}

	mockClientController := gomock.NewController(t)
	defer mockClientController.Finish()

	mockClient := client.NewMockHTTPClientDoer(mockClientController)

	mockResponseString := `
<?xml version="1.0"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope/" soap:encodingStyle="http://www.w3.org/2003/05/soap-encoding">
	<soap:Body xmlns:m="http://www.example.org/stock">
		<m:GetStockPriceResponse>
			<m:Price>34.5</m:Price>
		</m:GetStockPriceResponse>
	</soap:Body>
</soap:Envelope>`

	mockClient.EXPECT().Do(gomock.Any()).Return(&http.Response{StatusCode: http.StatusOK, Body: ioutil.NopCloser(bytes.NewBuffer(testWSDL))}, nil).Times(1)
	mockClient.EXPECT().Do(gomock.Any()).Return(&http.Response{StatusCode: http.StatusOK, Body: ioutil.NopCloser(bytes.NewBuffer([]byte(mockResponseString)))}, nil).Times(1)

	testService := Webservice{
		services: map[string]*wsdl.Definitions{},
		client:   client.Client{Client: mockClient},
	}

	err = testService.AddServices("http://mocked.com/ws?WSDL")
	if err != nil {
		t.Errorf("Expected err to be nil")
	}

	params := map[string]interface{}{
		"UnassignPrivateIpAddresses/networkInterfaceId":                          "1234512345",
		"UnassignPrivateIpAddresses/privateIpAddressesSet/item/privateIpAddress": "192.168.0.1",
	}

	var responseString string
	err = testService.Do("AmazonEC2", "UnassignPrivateIpAddresses", &responseString, params)
	if err != nil {
		t.Errorf("Expected nil error from testService.Do for valid request")
	}
}

func TestWebservice_Do_ErrorPath_MoreThanOneChoice(t *testing.T) {
	testWSDL, err := ioutil.ReadFile("fixtures/ec2.wsdl")
	if err != nil {
		t.Errorf("Error reading wsdl file")
	}

	mockClientController := gomock.NewController(t)
	defer mockClientController.Finish()

	mockClient := client.NewMockHTTPClientDoer(mockClientController)

	mockResponseString := `
<?xml version="1.0"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope/" soap:encodingStyle="http://www.w3.org/2003/05/soap-encoding">
	<soap:Body xmlns:m="http://www.example.org/stock">
		<m:GetStockPriceResponse>
			<m:Price>34.5</m:Price>
		</m:GetStockPriceResponse>
	</soap:Body>
</soap:Envelope>`

	mockClient.EXPECT().Do(gomock.Any()).Return(&http.Response{StatusCode: http.StatusOK, Body: ioutil.NopCloser(bytes.NewBuffer(testWSDL))}, nil).Times(1)
	mockClient.EXPECT().Do(gomock.Any()).Return(&http.Response{StatusCode: http.StatusOK, Body: ioutil.NopCloser(bytes.NewBuffer([]byte(mockResponseString)))}, nil).Times(0)

	testService := Webservice{
		services: map[string]*wsdl.Definitions{},
		client:   client.Client{Client: mockClient},
	}

	err = testService.AddServices("http://mocked.com/ws?WSDL")
	if err != nil {
		t.Errorf("Expected err to be nil")
	}

	params := map[string]interface{}{
		"ModifyNetworkInterfaceAttribute/networkInterfaceId": "1234512345",
		"ModifyNetworkInterfaceAttribute/description/":       "too many choices",
		"ModifyNetworkInterfaceAttribute/sourceDestCheck":    "1234512345",
	}

	var responseString string
	err = testService.Do("AmazonEC2", "ModifyNetworkInterfaceAttribute", &responseString, params)
	if err == nil {
		t.Errorf("Expected an error from testService.Do for too many choice parameters")
	}

	if err != nil && !strings.Contains(err.Error(), "A max of one choice element can be submitted") {
		t.Errorf("Unexpected error for %+v", err)
	}
}

func TestWebservice_Do_HappyPath_NoChoices(t *testing.T) {
	testWSDL, err := ioutil.ReadFile("fixtures/ec2.wsdl")
	if err != nil {
		t.Errorf("Error reading wsdl file")
	}

	mockClientController := gomock.NewController(t)
	defer mockClientController.Finish()

	mockClient := client.NewMockHTTPClientDoer(mockClientController)

	mockResponseString := `
<?xml version="1.0"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope/" soap:encodingStyle="http://www.w3.org/2003/05/soap-encoding">
	<soap:Body xmlns:m="http://www.example.org/stock">
		<m:GetStockPriceResponse>
			<m:Price>34.5</m:Price>
		</m:GetStockPriceResponse>
	</soap:Body>
</soap:Envelope>`

	mockClient.EXPECT().Do(gomock.Any()).Return(&http.Response{StatusCode: http.StatusOK, Body: ioutil.NopCloser(bytes.NewBuffer(testWSDL))}, nil).Times(1)
	mockClient.EXPECT().Do(gomock.Any()).Return(&http.Response{StatusCode: http.StatusOK, Body: ioutil.NopCloser(bytes.NewBuffer([]byte(mockResponseString)))}, nil).Times(1)

	testService := Webservice{
		services: map[string]*wsdl.Definitions{},
		client:   client.Client{Client: mockClient},
	}

	err = testService.AddServices("http://mocked.com/ws?WSDL")
	if err != nil {
		t.Errorf("Expected err to be nil")
	}

	params := map[string]interface{}{
		"ModifyNetworkInterfaceAttribute/networkInterfaceId": "1234512345",
	}

	var responseString string
	err = testService.Do("AmazonEC2", "ModifyNetworkInterfaceAttribute", &responseString, params)
	if err != nil {
		t.Errorf("Expected nil error from testService.Do for no choices submitted")
	}
}

func TestWebservice_Do_HappyPath_WithChoice(t *testing.T) {
	testWSDL, err := ioutil.ReadFile("fixtures/ec2.wsdl")
	if err != nil {
		t.Errorf("Error reading wsdl file")
	}

	mockClientController := gomock.NewController(t)
	defer mockClientController.Finish()

	mockClient := client.NewMockHTTPClientDoer(mockClientController)

	mockResponseString := `
<?xml version="1.0"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope/" soap:encodingStyle="http://www.w3.org/2003/05/soap-encoding">
	<soap:Body xmlns:m="http://www.example.org/stock">
		<m:GetStockPriceResponse>
			<m:Price>34.5</m:Price>
		</m:GetStockPriceResponse>
	</soap:Body>
</soap:Envelope>`

	mockClient.EXPECT().Do(gomock.Any()).Return(&http.Response{StatusCode: http.StatusOK, Body: ioutil.NopCloser(bytes.NewBuffer(testWSDL))}, nil).Times(1)
	mockClient.EXPECT().Do(gomock.Any()).Return(&http.Response{StatusCode: http.StatusOK, Body: ioutil.NopCloser(bytes.NewBuffer([]byte(mockResponseString)))}, nil).Times(1)

	testService := Webservice{
		services: map[string]*wsdl.Definitions{},
		client:   client.Client{Client: mockClient},
	}

	err = testService.AddServices("http://mocked.com/ws?WSDL")
	if err != nil {
		t.Errorf("Expected err to be nil")
	}

	params := map[string]interface{}{
		"ModifyNetworkInterfaceAttribute/networkInterfaceId":             "1234512345",
		"ModifyNetworkInterfaceAttribute/attachment/attachmentId":        "1234512345",
		"ModifyNetworkInterfaceAttribute/attachment/deleteOnTermination": true,
	}

	var responseString string
	err = testService.Do("AmazonEC2", "ModifyNetworkInterfaceAttribute", &responseString, params)
	if err != nil {
		t.Errorf("Expected nil error from testService.Do for no choices submitted")
	}
}

func TestWebservice_NewRequest_HappyPath_WithChoice(t *testing.T) {
	testWSDL, err := ioutil.ReadFile("fixtures/ec2.wsdl")
	if err != nil {
		t.Errorf("Error reading wsdl file")
	}

	mockClientController := gomock.NewController(t)
	defer mockClientController.Finish()

	mockClient := client.NewMockHTTPClientDoer(mockClientController)

	mockResponseString := `
<?xml version="1.0"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope/" soap:encodingStyle="http://www.w3.org/2003/05/soap-encoding">
	<soap:Body xmlns:m="http://www.example.org/stock">
		<m:GetStockPriceResponse>
			<m:Price>34.5</m:Price>
		</m:GetStockPriceResponse>
	</soap:Body>
</soap:Envelope>`

	mockClient.EXPECT().Do(gomock.Any()).Return(&http.Response{StatusCode: http.StatusOK, Body: ioutil.NopCloser(bytes.NewBuffer(testWSDL))}, nil).Times(1)
	mockClient.EXPECT().Do(gomock.Any()).Return(&http.Response{StatusCode: http.StatusOK, Body: ioutil.NopCloser(bytes.NewBuffer([]byte(mockResponseString)))}, nil).Times(0)

	testService := Webservice{
		services: map[string]*wsdl.Definitions{},
		client:   client.Client{Client: mockClient},
	}

	err = testService.AddServices("http://mocked.com/ws?WSDL")
	if err != nil {
		t.Errorf("Expected err to be nil")
	}

	params := map[string]interface{}{
		"ModifyNetworkInterfaceAttribute/networkInterfaceId":             "1234512345",
		"ModifyNetworkInterfaceAttribute/attachment/attachmentId":        "1234512345",
		"ModifyNetworkInterfaceAttribute/attachment/deleteOnTermination": true,
	}

	buf := new(bytes.Buffer)
	err = testService.NewRequest("AmazonEC2", "ModifyNetworkInterfaceAttribute", params, buf)
	if err != nil {
		t.Errorf("Expected nil error from testService.Do for no choices submitted")
	}

	expectedRequestString := `<?xml version="1.0" encoding="UTF-8"?>
<soap-env:Envelope xmlns:soap-env="http://schemas.xmlsoap.org/soap/envelope/">
  <soap-env:Body>
    <ns0:ModifyNetworkInterfaceAttribute xmlns:ns0="http://ec2.amazonaws.com/doc/2013-10-15/">
      <networkInterfaceId>1234512345</networkInterfaceId>
      <attachment>
        <attachmentId>1234512345</attachmentId>
        <deleteOnTermination>true</deleteOnTermination>
      </attachment>
    </ns0:ModifyNetworkInterfaceAttribute>
  </soap-env:Body>
</soap-env:Envelope>`

	if buf.String() != expectedRequestString {
		t.Errorf("Unexpected XML request")
	}
}
