package schemaToRest

import (
	"encoding/json"
	"fmt"

	"github.com/Jeffail/gabs"
	"github.com/graphql-go/graphql"
)

func introspectSchema(schema graphql.Schema) (map[string]interface{}, error) {
	data := map[string]interface{}{}

	// add methodes
	type Field struct {
		graphql.FieldDefinition
		Types []interface{}
	}
	type methode struct {
		Methode string
		Fields  []Field
	}
	methodes := []methode{}

	introspecType := func(fieldDefinitionMap graphql.FieldDefinitionMap) []Field {
		fields := []Field{}
		for _, field := range fieldDefinitionMap {
			params := graphql.Params{Schema: schema, RequestString: inspectType(field.Type.Name())}
			result := graphql.Do(params)
			dataByte, _ := json.Marshal(result.Data)
			dataParsed, _ := gabs.ParseJSON(dataByte)
			if typFields, ok := dataParsed.Path("__type.fields").Data().([]interface{}); ok {
				fields = append(fields, Field{*field, typFields})
			}
		}
		return fields
	}

	if schema.QueryType().Fields() != nil {
		methodes = append(methodes, methode{Methode: "Queries", Fields: introspecType(schema.QueryType().Fields())})
	}
	if schema.MutationType() != nil {
		methodes = append(methodes, methode{Methode: "Mutations", Fields: introspecType(schema.MutationType().Fields())})
	}
	data["Methodes"] = methodes

	return data, nil
}

func inspectType(typ string) string {
	return fmt.Sprintf(`query{
        __type(name: "%s") {
            name
            fields {
            name
            type {
              name
              kind
              ofType {
              name
              kind
              }
            }
            }
        }
      }`, typ)
}
