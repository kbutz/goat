package xsd

import (
	"strings"

	"github.com/sezzle/sezzle-go-xml"
)

type Schemaer interface {
	EncodeElement(name string, enc *xml.Encoder, sr SchemaRepository, params map[string]interface{}, useNamespace, keepUsingNamespace bool, path ...string) (err error)
	EncodeType(name string, enc *xml.Encoder, sr SchemaRepository, params map[string]interface{}, useNamespace, keepUsingNamespace bool, path ...string) (err error)
}

type GetAliaser interface {
	GetAlias(string) string
	Namespace() string
}

type SchemaRepository interface {
	GetSchema(space string) (Schemaer, error)
}

func hasPrefix(m map[string]interface{}, prefix string) bool {
	for k := range m {
		ok := strings.HasPrefix(k, prefix)
		if ok {
			return true
		}
	}

	return false
}

func MakePath(path []string) string {
	return strings.Join(path, "/")
}
