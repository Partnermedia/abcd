package lifeboat

// TEMPORARY integration stub — replaced by Agent B's graveyard_abandoned.go at merge.
func buildAbandoned(ctx *SourceContext) Abandoned {
	return Abandoned{SchemaVersion: GraveyardSchemaVersion, Findings: []Finding{}}
}
