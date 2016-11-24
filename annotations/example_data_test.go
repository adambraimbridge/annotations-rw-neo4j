package annotations

var (
	conceptWithoutAgent = annotations{annotation{
		Thing: thing{ID: getURI(conceptUUID),
			PrefLabel: "prefLabel",
			Types: []string{
				"http://www.ft.com/ontology/organisation/Organisation",
				"http://www.ft.com/ontology/core/Thing",
				"http://www.ft.com/ontology/concept/Concept",
			}},
		Provenances: []provenance{
			{
				Scores: []score{
					score{ScoringSystem: relevanceScoringSystem, Value: 0.9},
					score{ScoringSystem: confidenceScoringSystem, Value: 0.8},
				},
				AgentRole: "",
				AtTime:    "",
			},
		},
	}}
	conceptWithoutID = annotations{annotation{
		Thing: thing{
			PrefLabel: "prefLabel",
			Types: []string{
				"http://www.ft.com/ontology/organisation/Organisation",
				"http://www.ft.com/ontology/core/Thing",
				"http://www.ft.com/ontology/concept/Concept",
			}},
		Provenances: []provenance{
			{
				Scores: []score{
					score{ScoringSystem: relevanceScoringSystem, Value: 0.9},
					score{ScoringSystem: confidenceScoringSystem, Value: 0.8},
				},
				AgentRole: "http://api.ft.com/things/0edd3c31-1fd0-4ef6-9230-8d545be3880a",
				AtTime:    "2016-01-01T19:43:47.314Z",
			},
		},
	}}
	conceptWithPredicate = annotation{
		Thing: thing{ID: getURI(oldConceptUUID),
			PrefLabel: "prefLabel",
			Types: []string{
				"http://www.ft.com/ontology/organisation/Organisation",
				"http://www.ft.com/ontology/core/Thing",
				"http://www.ft.com/ontology/concept/Concept",
			},
			Predicate: "isClassifiedBy",
		},
		Provenances: []provenance{
			{
				Scores: []score{
					score{ScoringSystem: relevanceScoringSystem, Value: 0.9},
					score{ScoringSystem: confidenceScoringSystem, Value: 0.8},
				},
				AgentRole: "http://api.ft.com/things/0edd3c31-1fd0-4ef6-9230-8d545be3880a",
				AtTime:    "2016-01-01T19:43:47.314Z",
			},
		},
	}
	conceptWithAboutPredicate = annotation{
		Thing: thing{ID: getURI(oldConceptUUID),
			PrefLabel: "prefLabel",
			Types: []string{
				"http://www.ft.com/ontology/organisation/Organisation",
				"http://www.ft.com/ontology/core/Thing",
				"http://www.ft.com/ontology/concept/Concept",
			},
			Predicate: "about",
		},
		Provenances: []provenance{
			{
				Scores: []score{
					score{ScoringSystem: relevanceScoringSystem, Value: 0.9},
					score{ScoringSystem: confidenceScoringSystem, Value: 0.8},
				},
				AgentRole: "http://api.ft.com/things/0edd3c31-1fd0-4ef6-9230-8d545be3880a",
				AtTime:    "2016-01-01T19:43:47.314Z",
			},
		},
	}
	multiConceptAnnotations = annotations{annotation{
		Thing: thing{ID: getURI(conceptUUID),
			PrefLabel: "prefLabel",
			Types: []string{
				"http://www.ft.com/ontology/organisation/Organisation",
				"http://www.ft.com/ontology/core/Thing",
				"http://www.ft.com/ontology/concept/Concept",
			}},
		Provenances: []provenance{
			{
				Scores: []score{
					score{ScoringSystem: relevanceScoringSystem, Value: 0.9},
					score{ScoringSystem: confidenceScoringSystem, Value: 0.8},
				},
				AgentRole: "http://api.ft.com/things/0edd3c31-1fd0-4ef6-9230-8d545be3880a",
				AtTime:    "2016-01-01T19:43:47.314Z",
			},
		},
	}, annotation{
		Thing: thing{ID: getURI(secondConceptUUID),
			PrefLabel: "prefLabel",
			Types: []string{
				"http://www.ft.com/ontology/organisation/Organisation",
				"http://www.ft.com/ontology/core/Thing",
				"http://www.ft.com/ontology/concept/Concept",
			}},
		Provenances: []provenance{
			{
				Scores: []score{
					score{ScoringSystem: relevanceScoringSystem, Value: 0.4},
					score{ScoringSystem: confidenceScoringSystem, Value: 0.5},
				},
				AgentRole: "http://api.ft.com/things/0edd3c31-1fd0-4ef6-9230-8d545be3880a",
				AtTime:    "2016-01-01T19:43:47.314Z",
			},
		},
	}}
)

func exampleConcepts(uuid string) annotations {
	return annotations{exampleConcept(uuid)}
}

func exampleConcept(uuid string) annotation {
	return annotation{
		Thing: thing{ID: getURI(uuid),
			PrefLabel: "prefLabel",
			Types: []string{
				"http://www.ft.com/ontology/organisation/Organisation",
				"http://www.ft.com/ontology/core/Thing",
				"http://www.ft.com/ontology/concept/Concept",
			}},
		Provenances: []provenance{
			{
				Scores: []score{
					score{ScoringSystem: relevanceScoringSystem, Value: 0.9},
					score{ScoringSystem: confidenceScoringSystem, Value: 0.8},
				},
				AgentRole: "http://api.ft.com/things/0edd3c31-1fd0-4ef6-9230-8d545be3880a",
				AtTime:    "2016-01-01T19:43:47.314Z",
			},
		},
	}
}
