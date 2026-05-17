package entity

import "time"

type PlanLimit struct {
	ID         string
	PlanID     string
	LimitCode  string
	LimitValue int64
	LimitWindow string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
