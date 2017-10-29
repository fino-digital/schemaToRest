package schemaToRest

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/graphql-go/graphql"
	"github.com/labstack/echo"
)

// FunctionParamKey is the param
const FunctionParamKey = "param"

// ToRest returns HandlerFunc
func ToRest(route string, schema graphql.Schema) (string, echo.HandlerFunc) {
	return fmt.Sprintf("%s/:%s", route, FunctionParamKey), func(context echo.Context) error {
		// findout the query-function
		function := context.Param(FunctionParamKey)

		// get requestBody as Map
		bodyMap := new(map[string]interface{})

		defer context.Request().Body.Close()
		bodyByte, err := ioutil.ReadAll(context.Request().Body)
		if err != nil {
			return context.JSON(http.StatusBadRequest, "no argments in body")
		}
		if err := json.Unmarshal(bodyByte, &bodyMap); err != nil {
			return context.JSON(http.StatusBadRequest, "can't unmarshal map of arguments")
		}

		if field, ok := schema.QueryType().Fields()[function]; ok {
			resolveParams := graphql.ResolveParams{}
			if field.Resolve == nil {
				context.JSON(http.StatusInternalServerError, "Can't find resolve-methode")
			}
			response, err := field.Resolve(resolveParams)
			if err != nil {
				return context.JSON(http.StatusInternalServerError,
					fmt.Sprintf("Error in Resolve-Methode: %s", err))
			}
			return context.JSON(http.StatusOK, response)
		}
		return context.JSON(http.StatusBadRequest, fmt.Sprintf("Can't find '%s'", function))
	}
}
