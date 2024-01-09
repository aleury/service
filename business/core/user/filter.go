package user

import (
	"fmt"
	"net/mail"
	"time"

	"github.com/aleury/service/business/sys/validate"
	"github.com/google/uuid"
)

// QueryFilter holds the available fields a query can be filtered on.
type QueryFilter struct {
	ID               *uuid.UUID    `validate:"omitempty,uuid4"`
	Name             *string       `validate:"omitempty,min=3"`
	Email            *mail.Address `validate:"omitempty,email"`
	StartCreatedDate *time.Time    `validate:"omitempty"`
	EndCreatedDate   *time.Time    `validate:"omitempty"`
}

// Validate checks the data in the model is considered clean.
func (qf *QueryFilter) Validate() error {
	if err := validate.Check(qf); err != nil {
		return fmt.Errorf("validate: %w", err)
	}
	return nil
}

// WithUserID sets the ID field of the QueryFilter value.
func (qf *QueryFilter) WithUserID(userID uuid.UUID) {
	qf.ID = &userID
}

// WithName sets the Name field of the QueryFilter value.
func (qf *QueryFilter) WithName(name string) {
	qf.Name = &name
}

// WithEmail  sets the Email field of the QueryFilter value.
func (qf *QueryFilter) WithEmail(email mail.Address) {
	qf.Email = &email
}

// WithStartCreatedDate sets the StartCreatedDate field of the QueryFilter value.
func (qf *QueryFilter) WithStartCreatedDate(start time.Time) {
	qf.StartCreatedDate = &start
}

// WithEndCreatedDate sets the EndCreatedDate field of the QueryFilter value.
func (qf *QueryFilter) WithEndCreatedDate(end time.Time) {
	qf.EndCreatedDate = &end
}
