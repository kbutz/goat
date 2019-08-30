package goat

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/kbutz/goat/client"
	"github.com/kbutz/goat/wsdl"
)

func TestWebservice_AddServices_ErrorPath(t *testing.T) {
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

func TestWebservice_AddServices_HappyPath_SingleImport(t *testing.T) {
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

	// Assert that the definition map was created and has some expected values pulled from the public ec2.wsdl file stored in /fixtures
	amazonEC2ServiceDefinition, ok := testService.services["AmazonEC2"]
	if !ok {
		t.Errorf("Expected top level AmazonECT testService definition to exist")
	}

	topLevelSchemaMap := amazonEC2ServiceDefinition.Types.Schemas["http://ec2.amazonaws.com/doc/2013-10-15/"]
	if !ok {
		t.Errorf("Expected top level schema map to exist")
	}

	var foundNormalElement bool
	for i := range topLevelSchemaMap.Elements {
		if topLevelSchemaMap.Elements[i].Name == "GetConsoleOutput" {
			foundNormalElement = true
		}

		if topLevelSchemaMap.Elements[i].Name == "GetConsoleOutput" && topLevelSchemaMap.Elements[i].Type != "tns:GetConsoleOutputType" {
			t.Errorf("Expected Type %v for GetConsoleOutput to equal tns:GetConsoleOutputType", topLevelSchemaMap.Elements[i].Type)
		}
	}

	if !foundNormalElement {
		t.Errorf("Expected element GetConsoleOutput to exist")
	}

	var foundChoices bool
	var foundSequenceChoices bool
	var foundNormalSequenceWithChoices bool
	for i := range topLevelSchemaMap.ComplexTypes {
		if topLevelSchemaMap.ComplexTypes[i].Name == "BlockDeviceMappingItemType" {
			if len(topLevelSchemaMap.ComplexTypes[i].SequenceChoice) == 3 {
				foundSequenceChoices = true
			}
		}

		if topLevelSchemaMap.ComplexTypes[i].Name == "DeleteSecurityGroupType" {
			if len(topLevelSchemaMap.ComplexTypes[i].Choice) == 2 {
				foundChoices = true
			}
		}

		if topLevelSchemaMap.ComplexTypes[i].Name == "DescribeVpcAttributeResponseType" {
			if len(topLevelSchemaMap.ComplexTypes[i].Sequence) == 2 && len(topLevelSchemaMap.ComplexTypes[i].SequenceChoice) == 2 {
				foundNormalSequenceWithChoices = true
			}
		}
	}

	if !foundChoices {
		t.Errorf("Expected complexType with exactly two choice elements for DeleteSecurityGroupType")
	}

	if !foundSequenceChoices {
		t.Errorf("Expected complexType with exactly three sequence>choice elements for BlockDeviceMappingItemType")
	}

	if !foundNormalSequenceWithChoices {
		t.Errorf("Expected complexType with exactly two sequence>choice elements and two normal elements for DescribeVpcAttributeResponseType")
	}

	if !foundChoices {
		t.Errorf("Expected complexType with exactly three sequence>choice elements for BlockDeviceMappingItemType")
	}
}

func TestWebservice_AddServices_HappyPath_MultipleImports(t *testing.T) {
	operations, err := ioutil.ReadFile("fixtures/chromedata_operations.wsdl")
	if err != nil {
		t.Errorf("Error reading wsdl file")
	}

	typeDefinitions, err := ioutil.ReadFile("fixtures/chromedata_types.wsdl")
	if err != nil {
		t.Errorf("Error reading wsdl file")
	}

	mockClientController := gomock.NewController(t)
	defer mockClientController.Finish()

	mockClient := client.NewMockHTTPClientDoer(mockClientController)

	mockClient.EXPECT().Do(gomock.Any()).Return(&http.Response{StatusCode: http.StatusOK, Body: ioutil.NopCloser(bytes.NewBuffer(operations))}, nil).Times(1)
	mockClient.EXPECT().Do(gomock.Any()).Return(&http.Response{StatusCode: http.StatusOK, Body: ioutil.NopCloser(bytes.NewBuffer(typeDefinitions))}, nil).Times(1)

	testService := Webservice{
		services: map[string]*wsdl.Definitions{},
		client:   client.Client{Client: mockClient},
	}

	err = testService.AddServices("http://example.com/mockecchrome/ws?WSDL")
	if err != nil {
		t.Errorf("Expected err to be nil")
	}
}
