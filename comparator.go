package jmap

// Comparator represents one entry of an Email/query 'sort' argument.
// It specifies the property to sort by, the sort direction, and an optional collation.
type Comparator struct {
	// Property is the name of the property to sort on (e.g. "receivedAt", "from", "subject", "size").
	Property string `json:"property"`

	// IsAscending indicates whether the sort order is ascending.
	// If nil, the default value true is assumed.
	IsAscending *bool `json:"isAscending,omitempty"`

	// Collation specifies the collation to use for string comparison.
	// Nil means no collation is specified.
	Collation *string `json:"collation,omitempty"`
}
