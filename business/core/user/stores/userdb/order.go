package userdb

import (
	"fmt"

	"github.com/aleury/service/business/core/user"
	"github.com/aleury/service/business/data/order"
)

var orderByFields = map[string]string{
	user.OrderByID:      "user_id",
	user.OrderByName:    "name",
	user.OrderByEmail:   "email",
	user.OrderByRoles:   "roles",
	user.OrderByEnabled: "enabled",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}
	return fmt.Sprintf(" ORDER BY %s %s", by, orderBy.Direction), nil
}
