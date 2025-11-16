package dto

// DTO для простых статистических эндпоинтов.

type StatsDTO struct {
	ReviewAssignments map[string]int64 `json:"review_assignments"`
}


