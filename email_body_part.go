package jmap

// EmailBodyPart represents one node of the Email's MIME bodyStructure tree.
type EmailBodyPart struct {
	PartID      *string        `json:"partId,omitempty"`   // optional
	BlobID      *string        `json:"blobId,omitempty"`   // optional
	Size        int            `json:"size"`               // required
	Headers     []EmailHeader  `json:"headers"`            // required
	Name        *string        `json:"name,omitempty"`     // optional
	Type        string         `json:"type"`               // required
	Charset     *string        `json:"charset,omitempty"`  // optional
	Disposition *string        `json:"disposition,omitempty"` // optional
	CID         *string        `json:"cid,omitempty"`      // optional
	Language    []string       `json:"language"`           // nil encodes as null
	Location    *string        `json:"location,omitempty"` // optional
	SubParts    []EmailBodyPart `json:"subParts"`          // nil encodes as null
}
