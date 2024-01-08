package mid

import (
	"context"
	"net/http"

	"github.com/aleury/service/business/web/auth"
	v1 "github.com/aleury/service/business/web/v1"
	"github.com/aleury/service/foundation/web"
	"go.uber.org/zap"
)

// Errors handles errors coming out of the call chain. It detects normal
// application errors which are used to respond to the client in a uniform way.
// Unexpected errors (status >= 500) are logged.
func Errors(log *zap.SugaredLogger) web.Middleware {
	return func(handler web.Handler) web.Handler {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			if err := handler(ctx, w, r); err != nil {
				log.Errorw("ERROR", "trace_id", web.GetTraceID(ctx), "message", err)

				var status int
				var er v1.ErrorResponse

				switch {
				case v1.IsRequestError(err):
					reqErr := v1.GetRequestError(err)
					status = reqErr.Status
					er = v1.ErrorResponse{
						Error: reqErr.Error(),
					}

				case auth.IsAuthError(err):
					status = http.StatusUnauthorized
					er = v1.ErrorResponse{
						Error: http.StatusText(status),
					}

				default:
					status = http.StatusInternalServerError
					er = v1.ErrorResponse{
						Error: http.StatusText(status),
					}
				}

				if err := web.Respond(ctx, w, er, status); err != nil {
					return err
				}

				// If we receive the shutdown err we need to return it
				// back to the base handler to shut down the service.
				if web.IsShutdown(err) {
					return err
				}
			}
			return nil
		}
	}
}
