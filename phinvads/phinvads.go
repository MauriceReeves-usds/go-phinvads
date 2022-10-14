package phinvads

// all of these objects map back to the FHIR valueset schema found at:
// https://build.fhir.org/valueset.html

// Link is a link for walking the structure of PHINVADS data
type Link struct {
	Relation string
	Url      string
}

// Identifier comes from the FHIR datatypes. Schema is here:
// https://build.fhir.org/datatypes.html#Identifier
type Identifier struct {
	Use      string
	Type     string
	System   string
	value    string
	period   string
	assigner string
}

// Concept is the actual data elements that make up a valueset, for example "U": "Unknown", etc
type Concept struct {
	Code    string
	Display string
}

// ValueSet is the collection of concepts that make up a valueset
type ValueSet struct {
	Version string
	System  string
	Concept []Concept
}

// Composition is the makeup of the actual valueset
type Composition struct {
	LockedDate string
	Inactive   bool
	Include    []ValueSet
}

// Metadata is a collection of metadata provided by the Bundle object
type Metadata struct {
	LastUpdated string
}

// Resource is metadata about our actual valueset, which is in the bundle
type Resource struct {
	ResourceType string
	Id           string
	Version      string
	Name         string
	Title        string
	Status       string
	Experimental bool
	Date         string
	Description  string
	Publisher    string
	Identifier   []Identifier
	Compose      Composition
}

// Entry is the wrapper around the actual FHIR valueset, in that it has
// the two properties: fullUrl, and resource. FullUrl is where to access this
// from and Resource is the valueset itself
type Entry struct {
	FullUrl  string
	Resource Resource
}

// Response is the "Bundle" resource type that includes information about
// the paging as well as total entries available and the ID for the page.
// this is not part of the FHIR valueset schema, but is an extension on the
// FHIR Bundle type: https://build.fhir.org/bundle.html#bundle
type Response struct {
	ResourceType string
	Id           string
	Type         string
	Total        int64
	Meta         Metadata
	Link         []Link
	Entry        []Entry
}
