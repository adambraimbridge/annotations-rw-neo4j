package annotations

var (
	conceptWithoutAgent = Annotations{Annotation{
		Thing: Thing{ID: getURI(conceptUUID),
			PrefLabel: "prefLabel",
			Types: []string{
				"http://www.ft.com/ontology/organisation/Organisation",
				"http://www.ft.com/ontology/core/Thing",
				"http://www.ft.com/ontology/concept/Concept",
			}},
		Provenances: []Provenance{
			{
				Scores: []Score{
					Score{ScoringSystem: relevanceScoringSystem, Value: 0.9},
					Score{ScoringSystem: confidenceScoringSystem, Value: 0.8},
				},
				AgentRole: "",
				AtTime:    "",
			},
		},
	}}
	conceptWithoutID = Annotations{Annotation{
		Thing: Thing{
			PrefLabel: "prefLabel",
			Types: []string{
				"http://www.ft.com/ontology/organisation/Organisation",
				"http://www.ft.com/ontology/core/Thing",
				"http://www.ft.com/ontology/concept/Concept",
			}},
		Provenances: []Provenance{
			{
				Scores: []Score{
					Score{ScoringSystem: relevanceScoringSystem, Value: 0.9},
					Score{ScoringSystem: confidenceScoringSystem, Value: 0.8},
				},
				AgentRole: "http://api.ft.com/things/0edd3c31-1fd0-4ef6-9230-8d545be3880a",
				AtTime:    "2016-01-01T19:43:47.314Z",
			},
		},
	}}
	conceptWithPredicate = Annotation{
		Thing: Thing{ID: getURI(oldConceptUUID),
			PrefLabel: "prefLabel",
			Types: []string{
				"http://www.ft.com/ontology/organisation/Organisation",
				"http://www.ft.com/ontology/core/Thing",
				"http://www.ft.com/ontology/concept/Concept",
			},
			Predicate: "isClassifiedBy",
		},
		Provenances: []Provenance{
			{
				Scores: []Score{
					Score{ScoringSystem: relevanceScoringSystem, Value: 0.9},
					Score{ScoringSystem: confidenceScoringSystem, Value: 0.8},
				},
				AgentRole: "http://api.ft.com/things/0edd3c31-1fd0-4ef6-9230-8d545be3880a",
				AtTime:    "2016-01-01T19:43:47.314Z",
			},
		},
	}
	conceptWithAboutPredicate = Annotation{
		Thing: Thing{ID: getURI(oldConceptUUID),
			PrefLabel: "prefLabel",
			Types: []string{
				"http://www.ft.com/ontology/organisation/Organisation",
				"http://www.ft.com/ontology/core/Thing",
				"http://www.ft.com/ontology/concept/Concept",
			},
			Predicate: "about",
		},
		Provenances: []Provenance{
			{
				Scores: []Score{
					Score{ScoringSystem: relevanceScoringSystem, Value: 0.9},
					Score{ScoringSystem: confidenceScoringSystem, Value: 0.8},
				},
				AgentRole: "http://api.ft.com/things/0edd3c31-1fd0-4ef6-9230-8d545be3880a",
				AtTime:    "2016-01-01T19:43:47.314Z",
			},
		},
	}
	conceptWithHasAuthorPredicate = Annotation{
		Thing: Thing{ID: getURI(oldConceptUUID),
			PrefLabel: "prefLabel",
			Types: []string{
				"http://www.ft.com/ontology/person/Person",
				"http://www.ft.com/ontology/core/Thing",
				"http://www.ft.com/ontology/concept/Concept",
			},
			Predicate: "hasAuthor",
		},
		Provenances: []Provenance{
			{
				Scores: []Score{
					Score{ScoringSystem: relevanceScoringSystem, Value: 0.9},
					Score{ScoringSystem: confidenceScoringSystem, Value: 0.8},
				},
				AgentRole: "http://api.ft.com/things/0edd3c31-1fd0-4ef6-9230-8d545be3880a",
				AtTime:    "2016-01-01T19:43:47.314Z",
			},
		},
	}
	conceptWithHasContributorPredicate = Annotation{
		Thing: Thing{ID: getURI(oldConceptUUID),
			PrefLabel: "prefLabel",
			Types: []string{
				"http://www.ft.com/ontology/person/Person",
				"http://www.ft.com/ontology/core/Thing",
				"http://www.ft.com/ontology/concept/Concept",
			},
			Predicate: "hasContributor",
		},
		Provenances: []Provenance{
			{
				Scores: []Score{
					Score{ScoringSystem: relevanceScoringSystem, Value: 0.9},
					Score{ScoringSystem: confidenceScoringSystem, Value: 0.8},
				},
				AgentRole: "http://api.ft.com/things/0edd3c31-1fd0-4ef6-9230-8d545be3880a",
				AtTime:    "2016-01-01T19:43:47.314Z",
			},
		},
	}

	conceptWithHasDispayTagPredicate = Annotation{
		Thing: Thing{ID: getURI(oldConceptUUID),
			PrefLabel: "prefLabel",
			Types: []string{
				"http://www.ft.com/ontology/person/Person",
				"http://www.ft.com/ontology/core/Thing",
				"http://www.ft.com/ontology/concept/Concept",
			},
			Predicate: "hasDisplayTag",
		},
		Provenances: []Provenance{
			{
				Scores: []Score{
					Score{ScoringSystem: relevanceScoringSystem, Value: 0.9},
					Score{ScoringSystem: confidenceScoringSystem, Value: 0.8},
				},
				AgentRole: "http://api.ft.com/things/0edd3c31-1fd0-4ef6-9230-8d545be3880a",
				AtTime:    "2016-01-01T19:43:47.314Z",
			},
		},
	}
	multiConceptAnnotations = Annotations{Annotation{
		Thing: Thing{ID: getURI(conceptUUID),
			PrefLabel: "prefLabel",
			Types: []string{
				"http://www.ft.com/ontology/organisation/Organisation",
				"http://www.ft.com/ontology/core/Thing",
				"http://www.ft.com/ontology/concept/Concept",
			}},
		Provenances: []Provenance{
			{
				Scores: []Score{
					Score{ScoringSystem: relevanceScoringSystem, Value: 0.9},
					Score{ScoringSystem: confidenceScoringSystem, Value: 0.8},
				},
				AgentRole: "http://api.ft.com/things/0edd3c31-1fd0-4ef6-9230-8d545be3880a",
				AtTime:    "2016-01-01T19:43:47.314Z",
			},
		},
	}, Annotation{
		Thing: Thing{ID: getURI(secondConceptUUID),
			PrefLabel: "prefLabel",
			Types: []string{
				"http://www.ft.com/ontology/organisation/Organisation",
				"http://www.ft.com/ontology/core/Thing",
				"http://www.ft.com/ontology/concept/Concept",
			}},
		Provenances: []Provenance{
			{
				Scores: []Score{
					Score{ScoringSystem: relevanceScoringSystem, Value: 0.4},
					Score{ScoringSystem: confidenceScoringSystem, Value: 0.5},
				},
				AgentRole: "http://api.ft.com/things/0edd3c31-1fd0-4ef6-9230-8d545be3880a",
				AtTime:    "2016-01-01T19:43:47.314Z",
			},
		},
	}}
)

func exampleConcepts(uuid string) Annotations {
	return Annotations{exampleConcept(uuid)}
}

func exampleConcept(uuid string) Annotation {
	return Annotation{
		Thing: Thing{ID: getURI(uuid),
			PrefLabel: "prefLabel",
			Types: []string{
				"http://www.ft.com/ontology/organisation/Organisation",
				"http://www.ft.com/ontology/core/Thing",
				"http://www.ft.com/ontology/concept/Concept",
			}},
		Provenances: []Provenance{
			{
				Scores: []Score{
					Score{ScoringSystem: relevanceScoringSystem, Value: 0.9},
					Score{ScoringSystem: confidenceScoringSystem, Value: 0.8},
				},
				AgentRole: "http://api.ft.com/things/0edd3c31-1fd0-4ef6-9230-8d545be3880a",
				AtTime:    "2016-01-01T19:43:47.314Z",
			},
		},
	}
}
