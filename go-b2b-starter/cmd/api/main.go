// Package main provides the entry point for the B2B SaaS Starter
//
//	@title			B2B SaaS Starter API
//	@version		1.0
//	@description	This is the API server for B2B SaaS Starter.
//	@termsOfService	http://swagger.io/terms/
//
//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io
//
//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html
//
//	@host		localhost:8080
//	@BasePath	/api
//
//	@securityDefinitions.basic	BasicAuth
//
//	@externalDocs.description	OpenAPI
//	@externalDocs.url			https://swagger.io/resources/open-api/
package main

import "github.com/moasq/go-b2b-starter/internal/bootstrap"

func main() {
	bootstrap.Execute()
}
