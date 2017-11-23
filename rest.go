package schemaToRest

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
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
func WrapSchema(schema *graphql.Schema) echo.HandlerFunc {
	return func(context echo.Context) error {
		// findout the query-function
		split := strings.Split(context.Request().RequestURI, "/")
		function := split[len(split)-1]

		// get requestBody as Map
		bodyMap := new(map[string]interface{})

		if context.Request().Method == echo.POST {
			defer context.Request().Body.Close()
			bodyByte, err := ioutil.ReadAll(context.Request().Body)
			if err != nil {
				return context.JSON(HTTPStatusNoArguments, "no argments in body")
			}
			if err := json.Unmarshal(bodyByte, &bodyMap); err != nil {
				return context.JSON(HTTPStatusCantUnmarshal, "can't unmarshal map of arguments")
			}
		} else {
			bodyMap = &map[string]interface{}{}
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
		data, err := introspectSchema(*schema)
		if err != nil {
			return err
		}
		data["URL"] = url
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

	// tmpl, err := template.New(name).Delims("[[", "]]").ParseFiles(name)
	tmpl, err := fetchTemplate(name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
	}

	return tmpl.ExecuteTemplate(w, name, data)
}

func fetchTemplate(name string) (*template.Template, error) {
	tmpl := template.New(name).Delims("[[", "]]")
	if _, err := os.Stat(name); os.IsNotExist(err) {
		response, err := http.Get("https://raw.githubusercontent.com/fino-digital/schemaToRest/master/docu.html")
		if err != nil {
			log.Println(err)
		}

		defer response.Body.Close()
		bodyBytes, err := ioutil.ReadAll(response.Body)

		return tmpl.Parse(string(bodyBytes))
	}
	return tmpl.ParseFiles(name)
}
