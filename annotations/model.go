package annotations

//Annotations represents a collection of Annotation instances
type Annotations []Annotation

//Annotation is the main struct used to create and return structures
type Annotation struct {
	Provenances []struct {
		AgentRole string `json:"agentRole,omitempty"`
		AtTime    string `json:"atTime,omitempty"`
		Scores    []struct {
			ScoringSystem string  `json:"scoringSystem,omitempty"`
			Value         float64 `json:"value,omitempty"`
		} `json:"scores,omitempty"`
	} `json:"provenances,omitempty"`
	Thing struct {
		ID        string   `json:"id,omitempty"`
		PrefLabel string   `json:"prefLabel,omitempty"`
		Types     []string `json:"types,omitempty"`
	} `json:"thing,omitempty"`
}

const (
	mentionsPred = "http://www.ft.com/ontology/annotation/mentions"
	mentionsRel  = "MENTIONS"
)
