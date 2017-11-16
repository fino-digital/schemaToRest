package schemaToRest

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
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

// DeliverDocu delivers the docu
func DeliverDocu(schema *graphql.Schema, url string) echo.HandlerFunc {
	return func(context echo.Context) error {
		data := map[string]interface{}{
			"URL": url,
		}

		// add methodes
		type methode struct {
			Methode string
			Fields  graphql.FieldDefinitionMap
		}
		methodes := []methode{}

		if schema.QueryType().Fields() != nil {
			methodes = append(methodes, methode{Methode: "Queries", Fields: schema.QueryType().Fields()})
		}
		if schema.MutationType() != nil {
			methodes = append(methodes, methode{Methode: "Mutations", Fields: schema.MutationType().Fields()})
		}

		log.Println(schema.TypeMap())

		data["Methodes"] = methodes
		return context.Render(http.StatusOK, "docu.html", data)
	}
}

// GetTemplateRenderer returns a new TemplateRenderer
func GetTemplateRenderer() *TemplateRenderer {
	return &TemplateRenderer{}
}

// TemplateRenderer is a custom html/template renderer for Echo framework
type TemplateRenderer struct{}

// Render renders a template document
func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {

	// Add global methods if data is a map
	if viewContext, isMap := data.(map[string]interface{}); isMap {
		viewContext["reverse"] = c.Echo().Reverse
	}

	tmpl, err := template.New(name).Delims("[[", "]]").ParseFiles(name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
	}

	return tmpl.ExecuteTemplate(w, name, data)
}
