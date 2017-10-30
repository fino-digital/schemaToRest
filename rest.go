package schemaToRest

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/graphql-go/graphql"
	"github.com/labstack/echo"
)

const (
	HTTPStatusNoArguments         = 460
	HTTPStatusCantUnmarshal       = 461
	HTTPStatusCantResolveMethode  = 462
	HTTPStatusCantFindFunction    = 463
	HTTPStatusErrorResolveMethode = 520
)

// ToRest returns HandlerFunc
func ToRest(route string, schema *graphql.Schema) (string, echo.HandlerFunc) {
	return "/*", WrapSchema(schema)
}

// WrapSchema is the plain wrapper.
// NOTICE: YOUR ROUTE HAVE TO IMPLEMENT THE FunctionParamKey!
func WrapSchema(schema *graphql.Schema) echo.HandlerFunc {
	return func(context echo.Context) error {
		// findout the query-function
		split := strings.Split(context.Request().RequestURI, "/")
		function := split[len(split)-1]

		// get requestBody as Map
		bodyMap := new(map[string]interface{})

		defer context.Request().Body.Close()
		bodyByte, err := ioutil.ReadAll(context.Request().Body)
		if err != nil {
			return context.JSON(HTTPStatusNoArguments, "no argments in body")
		}
		if err := json.Unmarshal(bodyByte, &bodyMap); err != nil {
			return context.JSON(HTTPStatusCantUnmarshal, "can't unmarshal map of arguments")
		}

		if field, ok := schema.QueryType().Fields()[function]; ok {
			if field.Resolve == nil {
				return context.JSON(HTTPStatusCantResolveMethode, "Can't find resolve-methode")
			}
			resolveParams := graphql.ResolveParams{
				Context: context.Request().Context(),
				Args:    *bodyMap,
			}
			response, err := field.Resolve(resolveParams)
			if err != nil {
				return context.JSON(http.StatusInternalServerError,
					fmt.Sprintf("Error in Resolve-Methode: %s", err))
			}
			return context.JSON(http.StatusOK, response)
		}
		return context.JSON(HTTPStatusCantFindFunction, fmt.Sprintf("Can't find '%s'", function))
	}
}
