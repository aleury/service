package usersummary

import (
	"fmt"

	"github.com/aleury/service/business/sys/validate"
	"github.com/google/uuid"
)

// QueryFilter holds the available fields a query can be filtered on.
type QueryFilter struct {
	UserID   *uuid.UUID `validate:"omitempty"`
	UserName *string    `validate:"omitempty,min=3"`
}

func (qf *QueryFilter) Validate() error {
	if err := validate.Check(qf); err != nil {
		return fmt.Errorf("validate: %w", err)
	}
	return nil
}

// WithUserID sets the UserID field of the QueryFilter value.
func (qf *QueryFilter) WithUserID(userID uuid.UUID) {
	qf.UserID = &userID
}

// WithUserName sets the UserName field of the QueryFilter value.
func (qf *QueryFilter) WithUserName(userName string) {
	qf.UserName = &userName
}
