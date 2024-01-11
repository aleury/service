package usergrp

import (
	"fmt"
	"net/mail"
	"time"

	"github.com/aleury/service/business/core/user"
	"github.com/aleury/service/business/core/usersummary"
	"github.com/aleury/service/business/sys/validate"
)

// AppUser represents information about an individual user.
type AppUser struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Email        string   `json:"email"`
	Roles        []string `json:"roles"`
	PasswordHash []byte   `json:"-"`
	Department   string   `json:"department"`
	Enabled      bool     `json:"enabled"`
	DateCreated  string   `json:"dateCreated"`
	DateUpdated  string   `json:"dateUpdated"`
}

func toAppUser(usr user.User) AppUser {
	roles := make([]string, len(usr.Roles))
	for i, role := range usr.Roles {
		roles[i] = role.Name()
	}

	return AppUser{
		ID:           usr.ID.String(),
		Name:         usr.Name,
		Email:        usr.Email.Address,
		Roles:        roles,
		PasswordHash: usr.PasswordHash,
		Department:   usr.Department,
		Enabled:      usr.Enabled,
		DateCreated:  usr.DateCreated.Format(time.RFC3339),
		DateUpdated:  usr.DateUpdated.Format(time.RFC3339),
	}
}

// =============================================================================

// AppNewUser contains information needed to create a new user.
type AppNewUser struct {
	Name            string   `json:"name" validate:"required"`
	Email           string   `json:"email" validate:"required,email"`
	Roles           []string `json:"roles" validate:"required"`
	Department      string   `json:"department"`
	Password        string   `json:"password" validate:"required"`
	PasswordConfirm string   `json:"passwordConfirm" validate:"eqfield=Password"`
}

func toCoreNewUser(appUsr AppNewUser) (user.NewUser, error) {
	roles := make([]user.Role, len(appUsr.Roles))
	for i, roleStr := range appUsr.Roles {
		role, err := user.ParseRole(roleStr)
		if err != nil {
			return user.NewUser{}, fmt.Errorf("parsing role: %w", err)
		}
		roles[i] = role
	}

	addr, err := mail.ParseAddress(appUsr.Email)
	if err != nil {
		return user.NewUser{}, fmt.Errorf("parsing email: %w", err)
	}

	usr := user.NewUser{
		Name:            appUsr.Name,
		Email:           *addr,
		Roles:           roles,
		Department:      appUsr.Department,
		Password:        appUsr.Password,
		PasswordConfirm: appUsr.PasswordConfirm,
	}
	return usr, nil
}

// Validate checks the data in the model is considered clean.
func (appUsr AppNewUser) Validate() error {
	if err := validate.Check(appUsr); err != nil {
		return err
	}
	return nil
}

// =============================================================================

// AppUpdateUser contains information needed to update a user.
type AppUpdateUser struct {
	Name            *string  `json:"name"`
	Email           *string  `json:"email" validate:"omitempty,email"`
	Roles           []string `json:"roles"`
	Department      *string  `json:"department"`
	Password        *string  `json:"password"`
	PasswordConfirm *string  `json:"passwordConfirm" validate:"omitempty,eqfield=Password"`
	Enabled         *bool    `json:"enabled"`
}

func toCoreUpdateUser(appUsr AppUpdateUser) (user.UpdateUser, error) {
	var roles []user.Role
	if appUsr.Roles != nil {
		roles = make([]user.Role, len(appUsr.Roles))
		for i, roleStr := range appUsr.Roles {
			role, err := user.ParseRole(roleStr)
			if err != nil {
				return user.UpdateUser{}, fmt.Errorf("parsing role: %w", err)
			}
			roles[i] = role
		}
	}

	var addr *mail.Address
	if appUsr.Email != nil {
		var err error
		addr, err = mail.ParseAddress(*appUsr.Email)
		if err != nil {
			return user.UpdateUser{}, fmt.Errorf("parsing email: %w", err)
		}
	}

	usr := user.UpdateUser{
		Name:            appUsr.Name,
		Email:           addr,
		Roles:           roles,
		Department:      appUsr.Department,
		Password:        appUsr.Password,
		PasswordConfirm: appUsr.PasswordConfirm,
		Enabled:         appUsr.Enabled,
	}
	return usr, nil
}

// Validate checks the data in the model is considered clean.
func (appUsr AppUpdateUser) Validate() error {
	if err := validate.Check(appUsr); err != nil {
		return err
	}
	return nil
}

// =============================================================================

// AppSummary repesents informatin about an individual user and their products.
type AppSummary struct {
	UserID     string  `json:"userId"`
	UserName   string  `json:"userName"`
	TotalCount int     `json:"totalCount"`
	TotalCost  float64 `json:"totalCost"`
}

func toAppSummary(sum usersummary.Summary) AppSummary {
	return AppSummary{
		UserID:     sum.UserID.String(),
		UserName:   sum.UserName,
		TotalCount: sum.TotalCount,
		TotalCost:  sum.TotalCost,
	}
}
