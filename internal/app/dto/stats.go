package dto

type StatsDTO struct {
	ReviewAssignments map[string]int64 `json:"review_assignments"`
}
