package xsd

import (
	"fmt"
	"log"
)

type SchemaMap map[string]Schema

func (m SchemaMap) GetSchema(space string) (Schemaer, error) {
	switch space {
	case "http://www.w3.org/2001/XMLSchema":
		return baseSchema{}, nil
	default:
		ss, ok := m[space]

		if !ok {
			err := fmt.Errorf("schema namespace not found: '%s'", space)
			return nil, err
		}

		return &ss, nil
	}
}

func (m SchemaMap) GetElement(space, name string) *Element {
	schema, ok := m[space]
	if !ok {
		log.Printf("element namespace not found: '%s'", space)
		return nil
	}

	for _, elem := range schema.Elements {
		if elem.Name == name {
			return &elem
		}
	}

	return nil
}
