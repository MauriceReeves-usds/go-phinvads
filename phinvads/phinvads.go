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

type Composition struct {
	LockedDate string
	Inactive   bool
	Include    []ValueSet
}

type Resource struct {
	ResourceType string
	Id           string
	Version      string
	Name         string
	Title        string
	Status       string
	Date         string
	Description  string
	Publisher    string
	Identifier   []Identifier
	Compose      Composition
}

type Entry struct {
	FullUrl  string
	Resource Resource
}

type Response struct {
	ResourceType string
	Id           string
	Type         string
	Total        int64
	Link         []Link
	Entry        []Entry
}
