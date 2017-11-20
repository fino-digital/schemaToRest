package schemaToRest

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"

	"github.com/Jeffail/gabs"
	"github.com/graphql-go/graphql"
)

type Field struct {
	graphql.FieldDefinition
	Types     []interface{}
	Arguments []interface{}
}
type methode struct {
	Methode string
	Fields  []Field
}

func introspectSchema(schema graphql.Schema) (map[string]interface{}, error) {
	data := map[string]interface{}{}
	// add methodes
	methodes := []methode{}

	// function to introspec a methode
	introspecMethode := func(fieldDefinitionMap graphql.FieldDefinitionMap) []Field {
		fields := []Field{}
		for _, field := range fieldDefinitionMap {
			newField := Field{FieldDefinition: *field}
			// output
			if typFields, ok := inspectType(schema, field.Type.String()); ok {
				newField.Types = typFields
			}
			// input
			for _, argument := range field.Args {
				if argumentFields, ok := inspectType(schema, argument.Type.Name()); ok {
					newField.Arguments = argumentFields
				}
			}
			fields = append(fields, newField)
		}
		return fields
	}

	if schema.QueryType().Fields() != nil {
		methodes = append(methodes, methode{Methode: "Queries", Fields: introspecMethode(schema.QueryType().Fields())})
	}
	if schema.MutationType() != nil {
		methodes = append(methodes, methode{Methode: "Mutations", Fields: introspecMethode(schema.MutationType().Fields())})
	}

	data["Methodes"] = methodes

	return data, nil
}

func inspectType(schema graphql.Schema, typ string) ([]interface{}, bool) {
	reg, _ := regexp.Compile("[^a-zA-Z0-9]+")
	typ = reg.ReplaceAllString(typ, "")
	introspectionQuery := fmt.Sprintf(`query{
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

	params := graphql.Params{Schema: schema, RequestString: introspectionQuery}
	result := graphql.Do(params)
	dataByte, _ := json.Marshal(result.Data)
	dataParsed, _ := gabs.ParseJSON(dataByte)
	log.Println(string(dataByte))

	log.Println(typ)

	fields, ok := dataParsed.Path("__type.fields").Data().([]interface{})
	return fields, ok
}
