package schemaToRest_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/TobiEiss/schemaToRest"

	"github.com/graphql-go/graphql"
	"github.com/labstack/echo"
)

var TestSchemaConfig = graphql.SchemaConfig{
	Query: graphql.NewObject(graphql.ObjectConfig{
		Name: "testQuery",
		Fields: graphql.Fields{
			"testField": &graphql.Field{
				Type: graphql.String,
				Resolve: func(param graphql.ResolveParams) (interface{}, error) {
					return "Hello", nil
				},
			},
			"withoutResolveMethode": &graphql.Field{
				Type: graphql.String,
			},
			"returnError": &graphql.Field{
				Type: graphql.String,
				Resolve: func(param graphql.ResolveParams) (interface{}, error) {
					return "Hello", errors.New("Error")
				},
			},
		},
	}),
}

func TestFoundFoundRoute(t *testing.T) {
	testData := []struct {
		ExpectedStatus int
		Route          string
		Body           map[string]interface{}
	}{
		{
			ExpectedStatus: 200,
			Route:          "/myRoute/testField",
			Body:           map[string]interface{}{"test": "test"},
		},
	}

	// build TestSchema
	schema, err := graphql.NewSchema(TestSchemaConfig)
	if err != nil {
		t.Errorf("Can't build schema: %s", err)
	}

	router := echo.New()
	router.POST(schemaToRest.ToRest("/myRoute", schema))

	// iterate all tests
	for testIndex, test := range testData {
		// build body
		jsonBodyByte, _ := json.Marshal(test.Body)

		// build requst
		request := httptest.NewRequest(echo.POST,
			"http://example.com"+test.Route,
			bytes.NewReader(jsonBodyByte))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, request)

		// check result
		if rec.Result().StatusCode != test.ExpectedStatus {
			t.Errorf("[%d] failed. actual %d expected %d; Body: %s",
				testIndex, rec.Result().StatusCode, test.ExpectedStatus, rec.Body.String())
		}
	}
}
