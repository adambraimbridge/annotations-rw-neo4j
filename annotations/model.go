package people

type person struct {
	BirthYear   int          `json:"birthYear,omitempty"`
	Identifiers []identifier `json:"identifiers,omitempty"`
	Name        string       `json:"name,omitempty"`
	UUID        string       `json:"uuid"`
	Salutation  string       `json:"salutation,omitempty"`
}

type identifier struct {
	Authority       string `json:"authority"`
	IdentifierValue string `json:"identifierValue"`
}
