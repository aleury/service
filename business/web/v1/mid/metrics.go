package mid

import (
	"context"
	"net/http"

	"github.com/aleury/service/business/web/metrics"
	"github.com/aleury/service/foundation/web"
)

// Metrics updates program counters.
func Metrics() web.Middleware {
	return func(handler web.Handler) web.Handler {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			ctx = metrics.Set(ctx)

			err := handler(ctx, w, r)
			if err != nil {
				metrics.AddErrors(ctx)
			}
			metrics.AddRequests(ctx)
			metrics.AddGoroutines(ctx)

			return err
		}
	}
}
