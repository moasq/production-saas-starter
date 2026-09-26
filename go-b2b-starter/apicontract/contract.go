// Package apicontract owns the versioned public business API contract.
package apicontract

import _ "embed"

// Document is the canonical OpenAPI source for frontend generation and route tests.
//
//go:embed openapi.json
var Document []byte
