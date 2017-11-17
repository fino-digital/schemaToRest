package schemaToRest

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/testutil"
	"github.com/labstack/echo"
)

var TestSchemaConfig = graphql.SchemaConfig{
	Query: graphql.NewObject(graphql.ObjectConfig{
		Name: "testQuery",
		Fields: graphql.Fields{
			"testField": &graphql.Field{
				Type:        graphql.String,
				Description: "Description1",
				Resolve: func(param graphql.ResolveParams) (interface{}, error) {
					return "Hello", nil
				},
			},
			"withoutResolveMethode": &graphql.Field{
				Type:        graphql.String,
				Description: "Description2",
			},
			"returnError": &graphql.Field{
				Type:        graphql.String,
				Description: "Description3",
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
		Body           interface{}
	}{
		{
			ExpectedStatus: 200,
			Route:          "/myRoute/testField",
			Body:           map[string]interface{}{"test": "test"},
		},
		{
			ExpectedStatus: HTTPStatusCantUnmarshal,
			Route:          "/myRoute/testField",
			Body:           true,
		},
		{
			ExpectedStatus: HTTPStatusCantResolveMethode,
			Route:          "/myRoute/withoutResolveMethode",
			Body:           map[string]interface{}{"test": "test"},
		},
		{
			ExpectedStatus: http.StatusInternalServerError,
			Route:          "/myRoute/returnError",
			Body:           map[string]interface{}{"test": "test"},
		},
		{
			ExpectedStatus: HTTPStatusCantFindFunction,
			Route:          "/myRoute/fail",
			Body:           map[string]interface{}{"test": "test"},
		},
	}

	// build TestSchema
	schema, err := graphql.NewSchema(TestSchemaConfig)
	if err != nil {
		t.Errorf("Can't build schema: %s", err)
	}

	router := echo.New()
	router.POST(ToRest("/myRoute", &schema))

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
