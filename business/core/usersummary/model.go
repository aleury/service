package usersummary

import "github.com/google/uuid"

// Summary repesents information about an individual user and their products.
type Summary struct {
	UserID     uuid.UUID
	UserName   string
	TotalCount int
	TotalCost  float64
}
