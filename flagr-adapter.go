package vexilla

import (
	"encoding/json"
)

// FlagrResponse representa a resposta completa do Flagr API
type FlagrResponse struct {
	ID                 int64          `json:"id"`
	Key                string         `json:"key"`
	Description        string         `json:"description"`
	Enabled            bool           `json:"enabled"`
	Tags               []interface{}  `json:"tags"`
	DataRecordsEnabled bool           `json:"dataRecordsEnabled"`
	Segments           []FlagrSegment `json:"segments"`
	Variants           []FlagrVariant `json:"variants"`
	UpdatedAt          string         `json:"updatedAt"`
}

// FlagrSegment representa um segmento do Flagr
type FlagrSegment struct {
	ID             int64               `json:"id"`
	Description    string              `json:"description"`
	Constraints    []FlagrConstraint   `json:"constraints"`
	Distributions  []FlagrDistribution `json:"distributions"`
	RolloutPercent int                 `json:"rolloutPercent"`
	Rank           int                 `json:"rank"`
}

// FlagrConstraint representa uma constraint do Flagr
type FlagrConstraint struct {
	ID       int64       `json:"id"`
	Property string      `json:"property"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
}

// FlagrDistribution representa uma distribuição de variantes
type FlagrDistribution struct {
	ID         int64  `json:"id"`
	VariantID  int64  `json:"variantID"`
	VariantKey string `json:"variantKey"`
	Percent    int    `json:"percent"`
}

// FlagrVariant representa uma variante do Flagr
type FlagrVariant struct {
	ID         int64                  `json:"id"`
	Key        string                 `json:"key"`
	Attachment map[string]interface{} `json:"attachment"`
}

// parseFlagrResponse converte a resposta do Flagr para o formato Vexilla
func parseFlagrResponse(flagrFlags []FlagrResponse) []Flag {
	vexillaFlags := make([]Flag, 0, len(flagrFlags))

	for _, ff := range flagrFlags {
		// Converter segmentos
		segments := make([]Segment, 0, len(ff.Segments))
		for _, fs := range ff.Segments {
			// Converter constraints
			constraints := make([]Constraint, 0, len(fs.Constraints))
			for _, fc := range fs.Constraints {
				constraints = append(constraints, Constraint{
					Property: fc.Property,
					Operator: fc.Operator,
					Value:    fc.Value,
				})
			}

			// Converter distributions
			distributions := make([]VariantDistribution, 0, len(fs.Distributions))
			for _, fd := range fs.Distributions {
				// Encontrar o attachment correspondente
				var attachment interface{}
				for _, variant := range ff.Variants {
					if variant.ID == fd.VariantID {
						attachment = variant.Attachment
						break
					}
				}

				distributions = append(distributions, VariantDistribution{
					VariantID:  string(rune(fd.VariantID)),
					VariantKey: fd.VariantKey,
					Percentage: fd.Percent,
					Attachment: attachment,
				})
			}

			segments = append(segments, Segment{
				ID:             string(rune(fs.ID)),
				RolloutPercent: fs.RolloutPercent,
				Constraints:    constraints,
				Distributions:  distributions,
			})
		}

		// Determinar valor default
		// Se houver variants, pegar o attachment do primeiro variant habilitado
		var defaultValue interface{} = false
		if len(ff.Variants) > 0 {
			defaultValue = ff.Variants[0].Attachment
		}

		vexillaFlag := Flag{
			ID:       int(ff.ID),
			Key:      ff.Key,
			Default:  defaultValue,
			Rules:    []FlagRule{}, // Flagr não usa rules dessa forma
			Segments: segments,
		}

		vexillaFlags = append(vexillaFlags, vexillaFlag)
	}

	return vexillaFlags
}

// parseFlagrResponseBytes é um helper que parse direto do JSON
func parseFlagrResponseBytes(data []byte) ([]Flag, error) {
	var flagrFlags []FlagrResponse
	if err := json.Unmarshal(data, &flagrFlags); err != nil {
		return nil, err
	}
	return parseFlagrResponse(flagrFlags), nil
}
