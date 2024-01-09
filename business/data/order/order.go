package order

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/aleury/service/business/sys/validate"
)

// Set of directions for data ordering
const (
	ASC  = "ASC"
	DESC = "DESC"
)

var directions = map[string]string{
	ASC:  "ASC",
	DESC: "DESC",
}

// =============================================================================

// By repesents a field used to order by and direction.
type By struct {
	Field     string
	Direction string
}

// NewBy constructs a new By value with no checks.
func NewBy(field string, direction string) By {
	return By{
		Field:     field,
		Direction: direction,
	}
}

// =============================================================================

// Parse constructs a order.By value by parsing a string in the form
// of "field,direction".
func Parse(r *http.Request, defaultOrder By) (By, error) {
	v := r.URL.Query().Get("orderBy")
	if v == "" {
		return defaultOrder, nil
	}

	orderParts := strings.Split(v, ",")

	var by By
	switch len(orderParts) {
	case 1:
		by = NewBy(strings.Trim(orderParts[0], " "), ASC)
	case 2:
		by = NewBy(strings.Trim(orderParts[0], " "), strings.Trim(orderParts[1], " "))
	default:
		return By{}, validate.NewFieldsError(v, errors.New("unknown order field"))
	}

	if _, ok := directions[by.Direction]; !ok {
		return By{}, validate.NewFieldsError(v, fmt.Errorf("unknown direction: %s", by.Direction))
	}

	return by, nil
}
