package entity

import (
	"encoding/json"
	"time"
)

type ComplianceCaseStatus string

const (
	ComplianceCaseOpen     ComplianceCaseStatus = "OPEN"
	ComplianceCaseReview   ComplianceCaseStatus = "REVIEW"
	ComplianceCaseClosed   ComplianceCaseStatus = "CLOSED"
)

type ComplianceCaseType string

const (
	ComplianceCaseKYCReview    ComplianceCaseType = "KYC_REVIEW"
	ComplianceCaseAMLAlert     ComplianceCaseType = "AML_ALERT"
	ComplianceCaseKYTAlert     ComplianceCaseType = "KYT_ALERT"
	ComplianceCaseSanctionsHit ComplianceCaseType = "SANCTIONS_HIT"
	ComplianceCaseTravelRule   ComplianceCaseType = "TRAVEL_RULE"
	ComplianceCaseFraud        ComplianceCaseType = "FRAUD"
)

type ComplianceRiskLevel string

const (
	ComplianceRiskLow      ComplianceRiskLevel = "LOW"
	ComplianceRiskMedium   ComplianceRiskLevel = "MEDIUM"
	ComplianceRiskHigh     ComplianceRiskLevel = "HIGH"
	ComplianceRiskCritical ComplianceRiskLevel = "CRITICAL"
)

type ComplianceCase struct {
	ID        string
	UserID    string
	Type      ComplianceCaseType
	Status    ComplianceCaseStatus
	RiskLevel ComplianceRiskLevel
	Title     string
	Summary   string
	Metadata  json.RawMessage
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ComplianceEvent struct {
	ID        string
	CaseID    string
	EventType string
	Payload   json.RawMessage
	CreatedAt time.Time
}
