package flagr

import (
	"github.com/OrlandoBitencourt/vexilla/pkg/domain"
)

// ToDomain converts Flagr API models to domain models

// FlagToDomain converts FlagrFlag to domain.Flag
func FlagToDomain(f *FlagrFlag) domain.Flag {
	return domain.Flag{
		ID:                 f.ID,
		Key:                f.Key,
		Description:        f.Description,
		Enabled:            f.Enabled,
		Segments:           SegmentsToDomain(f.Segments),
		Variants:           VariantsToDomain(f.Variants),
		UpdatedAt:          f.UpdatedAt,
		Tags:               TagsToDomain(f.Tags),
		DataRecordsEnabled: f.DataRecordsEnabled,
	}
}

// FlagsToDomain converts multiple FlagrFlags to domain.Flags
func FlagsToDomain(flags []FlagrFlag) []domain.Flag {
	result := make([]domain.Flag, len(flags))
	for i, f := range flags {
		result[i] = FlagToDomain(&f)
	}
	return result
}

// SegmentsToDomain converts FlagrSegments to domain.Segments
func SegmentsToDomain(segments []FlagrSegment) []domain.Segment {
	result := make([]domain.Segment, len(segments))
	for i, s := range segments {
		result[i] = SegmentToDomain(s)
	}
	return result
}

// SegmentToDomain converts FlagrSegment to domain.Segment
func SegmentToDomain(s FlagrSegment) domain.Segment {
	return domain.Segment{
		ID:             s.ID,
		Rank:           s.Rank,
		Description:    s.Description,
		RolloutPercent: int(s.RolloutPercent),
		Constraints:    ConstraintsToDomain(s.Constraints),
		Distributions:  DistributionsToDomain(s.Distributions),
	}
}

// ConstraintsToDomain converts FlagrConstraints to domain.Constraints
func ConstraintsToDomain(constraints []FlagrConstraint) []domain.Constraint {
	result := make([]domain.Constraint, len(constraints))
	for i, c := range constraints {
		result[i] = ConstraintToDomain(c)
	}
	return result
}

// ConstraintToDomain converts FlagrConstraint to domain.Constraint
func ConstraintToDomain(c FlagrConstraint) domain.Constraint {
	return domain.Constraint{
		ID:       c.ID,
		Property: c.Property,
		Operator: domain.Operator(c.Operator),
		Value:    c.Value,
	}
}

// DistributionsToDomain converts FlagrDistributions to domain.Distributions
func DistributionsToDomain(distributions []FlagrDistribution) []domain.Distribution {
	result := make([]domain.Distribution, len(distributions))
	for i, d := range distributions {
		result[i] = DistributionToDomain(d)
	}
	return result
}

// DistributionToDomain converts FlagrDistribution to domain.Distribution
func DistributionToDomain(d FlagrDistribution) domain.Distribution {
	return domain.Distribution{
		ID:        d.ID,
		VariantID: d.VariantID,
		Percent:   int(d.Percent),
	}
}

// VariantsToDomain converts FlagrVariants to domain.Variants
func VariantsToDomain(variants []FlagrVariant) []domain.Variant {
	result := make([]domain.Variant, len(variants))
	for i, v := range variants {
		result[i] = VariantToDomain(v)
	}
	return result
}

// VariantToDomain converts FlagrVariant to domain.Variant
func VariantToDomain(v FlagrVariant) domain.Variant {
	return domain.Variant{
		ID:         v.ID,
		Key:        v.Key,
		Attachment: v.Attachment,
	}
}

// TagsToDomain converts Flagr tags to string slice
func TagsToDomain(tags []Tag) []domain.Tag {
	result := make([]domain.Tag, len(tags))
	for i, t := range tags {
		result[i] = domain.Tag{Value: t.Value}
	}
	return result
}

// EvaluationResultToDomain converts EvaluationResponse to domain.EvaluationResult
func EvaluationResultToDomain(resp EvaluationResponse) domain.EvaluationResult {
	return domain.EvaluationResult{
		FlagID:            resp.FlagID,
		FlagKey:           resp.FlagKey,
		SegmentID:         resp.SegmentID,
		VariantID:         resp.VariantID,
		VariantKey:        resp.VariantKey,
		VariantAttachment: resp.VariantAttachment,
		EvaluationReason:  extractEvaluationReason(resp.EvalDebugLog),
		Timestamp:         resp.Timestamp,
	}
}

// extractEvaluationReason extracts reason from debug log
func extractEvaluationReason(log EvalDebugLog) string {
	if log.Msg != "" {
		return log.Msg
	}
	if len(log.SegmentDebugLogs) > 0 {
		return log.SegmentDebugLogs[0].Msg
	}
	return "evaluated successfully"
}

// FromDomain converts domain models to Flagr API models

// EvaluationContextFromDomain converts domain.EvaluationContext to EvaluationRequest
func EvaluationContextFromDomain(flagKey string, ctx domain.EvaluationContext) EvaluationRequest {
	return EvaluationRequest{
		FlagKey:       flagKey,
		EntityID:      ctx.EntityID,
		EntityType:    ctx.EntityType,
		EntityContext: ctx.Context,
	}
}
