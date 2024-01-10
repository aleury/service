package userdb

import (
	"database/sql"
	"net/mail"
	"time"

	"github.com/aleury/service/business/core/user"
	"github.com/aleury/service/business/sys/database/pgx/dbarray"
	"github.com/google/uuid"
)

// dbUser repesents the structure we need for moving data
// between the app and the database.
type dbUser struct {
	ID           uuid.UUID      `db:"user_id"`
	Name         string         `db:"name"`
	Email        string         `db:"email"`
	Roles        dbarray.String `db:"roles"`
	PasswordHash []byte         `db:"password_hash"`
	Enabled      bool           `db:"enabled"`
	Department   sql.NullString `db:"department"`
	DateCreated  time.Time      `db:"date_created"`
	DateUpdated  time.Time      `db:"date_updated"`
}

func toDBUser(usr user.User) dbUser {
	roles := make([]string, len(usr.Roles))
	for i, role := range usr.Roles {
		roles[i] = role.Name()
	}

	return dbUser{
		ID:           usr.ID,
		Name:         usr.Name,
		Email:        usr.Email.String(),
		Roles:        dbarray.String(roles),
		PasswordHash: usr.PasswordHash,
		Enabled:      usr.Enabled,
		Department: sql.NullString{
			String: usr.Department,
			Valid:  usr.Department != "",
		},
		DateCreated: usr.DateCreated.UTC(),
		DateUpdated: usr.DateUpdated.UTC(),
	}
}

func toCoreUser(dbUsr dbUser) user.User {
	email := mail.Address{
		Address: dbUsr.Email,
	}

	roles := make([]user.Role, len(dbUsr.Roles))
	for i, dbRole := range dbUsr.Roles {
		roles[i] = user.MustParseRole(dbRole)
	}

	usr := user.User{
		ID:           dbUsr.ID,
		Name:         dbUsr.Name,
		Email:        email,
		Roles:        roles,
		PasswordHash: dbUsr.PasswordHash,
		Department:   dbUsr.Department.String,
		Enabled:      dbUsr.Enabled,
		DateCreated:  dbUsr.DateCreated.In(time.Local),
		DateUpdated:  dbUsr.DateUpdated.In(time.Local),
	}

	return usr
}

func toCoreUserSlice(dbUsers []dbUser) []user.User {
	users := make([]user.User, len(dbUsers))
	for i, dbUsr := range dbUsers {
		users[i] = toCoreUser(dbUsr)
	}
	return users
}
