package contracts

import (
	"encoding/json"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/moasq/go-b2b-starter/apicontract"
	"github.com/moasq/go-b2b-starter/internal/modules/auth"
	"github.com/moasq/go-b2b-starter/internal/modules/billing"
	"github.com/moasq/go-b2b-starter/internal/modules/organizations"
)

type resolver struct{ middleware *auth.Middleware }

func (r resolver) Get(name string) gin.HandlerFunc {
	switch name {
	case "auth":
		return r.middleware.RequireAuth()
	case "org_context":
		return r.middleware.RequireOrganization()
	default:
		panic(name)
	}
}

func TestMountedBusinessRoutesMatchOpenAPIAndRequireSession(t *testing.T) {
	var document struct {
		Paths map[string]map[string]json.RawMessage `json:"paths"`
	}
	if err := json.Unmarshal(apicontract.Document, &document); err != nil {
		t.Fatal(err)
	}
	router := gin.New()
	group := router.Group("/api")
	// Nil downstream services deliberately make an auth bypass fail the test.
	middleware := resolver{auth.NewMiddleware(nil, nil)}
	organizations.NewRoutes(nil).Routes(group, middleware)
	billing.NewHandler(nil).Routes(group, middleware)
	actual := map[string]bool{}
	parameter := regexp.MustCompile(`:([^/]+)`)
	for _, route := range router.Routes() {
		path := parameter.ReplaceAllString(strings.TrimPrefix(route.Path, "/api"), `{$1}`)
		key := route.Method + " " + path
		actual[key] = true
		if _, ok := document.Paths[path][strings.ToLower(route.Method)]; !ok {
			t.Errorf("mounted operation is undocumented: %s", key)
		}
		t.Run(key, func(t *testing.T) {
			request := httptest.NewRequest(route.Method, parameter.ReplaceAllString(route.Path, "fixture-member"), nil)
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, request)
			if recorder.Code != 401 {
				t.Fatalf("mounted operation must deny missing session before dependencies: %d %s", recorder.Code, recorder.Body)
			}
		})
	}
	for path, operations := range document.Paths {
		for method := range operations {
			if !actual[strings.ToUpper(method)+" "+path] {
				t.Errorf("documented operation is not mounted: %s %s", method, path)
			}
		}
	}
}
