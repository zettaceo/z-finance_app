package entity

import "time"

type JurisdictionRisk string

const (
	JurisdictionRiskLow    JurisdictionRisk = "LOW"
	JurisdictionRiskMedium JurisdictionRisk = "MEDIUM"
	JurisdictionRiskHigh   JurisdictionRisk = "HIGH"
)

type AMLTier string

const (
	AMLTierBasic         AMLTier = "BASIC"
	AMLTierEnhanced      AMLTier = "ENHANCED"
	AMLTierInstitutional AMLTier = "INSTITUTIONAL"
)

type RegulatoryProfile struct {
	ID                         string
	UserID                     string
	JurisdictionCode           string
	JurisdictionRisk           JurisdictionRisk
	AMLTier                    AMLTier
	TravelRuleRequired         bool
	SanctionsScreeningRequired bool
	CreatedAt                  time.Time
	UpdatedAt                  time.Time
}
