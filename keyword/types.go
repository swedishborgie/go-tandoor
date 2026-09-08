// Package keyword contains types for Tandoor Keyword resources.
package keyword

import "time"

// Keyword is a label/category for recipes (tree-structured).
type Keyword struct {
	ID          int       `json:"id,omitempty"`
	Name        string    `json:"name,omitempty"`
	Label       string    `json:"label,omitempty"`
	Description string    `json:"description,omitempty"`
	Parent      int       `json:"parent,omitempty"`
	NumChild    int       `json:"numchild,omitempty"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"`
	FullName    string    `json:"full_name,omitempty"`
}

// Label is a minimal keyword reference (id, label).
type Label struct {
	ID    int    `json:"id"`
	Label string `json:"label"`
}
