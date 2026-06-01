// Package models mirrors the PandaProbe API response and request schemas. Field
// names and nullability follow the OpenAPI spec exactly. Arbitrary JSON payloads
// (trace/span input, output, metadata, etc.) are kept as json.RawMessage so the
// CLI passes them through to agents losslessly.
package models

import "encoding/json"

// JSON is an arbitrary JSON value passed through verbatim.
type JSON = json.RawMessage

// Paginated is the wire shape the API uses for every list endpoint.
type Paginated[T any] struct {
	Items  []T `json:"items"`
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

// Pagination is the block surfaced in CLI list output.
type Pagination struct {
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

// ListResult is the CLI's output shape for list commands: the items plus a
// nested pagination block. It decouples the rendered shape from the flat wire
// shape so output is stable regardless of upstream changes.
type ListResult[T any] struct {
	Items      []T        `json:"items"`
	Pagination Pagination `json:"pagination"`
}

// AsList converts a wire Paginated response to the CLI output shape.
func AsList[T any](p *Paginated[T]) ListResult[T] {
	items := p.Items
	if items == nil {
		items = []T{}
	}
	return ListResult[T]{
		Items:      items,
		Pagination: Pagination{Total: p.Total, Limit: p.Limit, Offset: p.Offset},
	}
}
