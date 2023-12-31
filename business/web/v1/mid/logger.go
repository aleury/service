package mid

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/aleury/service/foundation/web"
	"go.uber.org/zap"
)

func Logger(log *zap.SugaredLogger) web.Middleware {
	return func(handler web.Handler) web.Handler {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			v := web.GetValues(ctx)

			path := r.URL.Path
			if r.URL.RawQuery != "" {
				path = fmt.Sprintf("%s?%s", path, r.URL.RawQuery)
			}

			log.Infow("request started",
				"trace_id", v.TraceID,
				"method", r.Method,
				"path", path,
				"remoteAddr", r.RemoteAddr,
			)

			err := handler(ctx, w, r)

			log.Infow("request completed",
				"trace_id", v.TraceID,
				"method", r.Method,
				"path", path,
				"remoteAddr", r.RemoteAddr,
				"statusCode", fmt.Sprint(v.StatusCode),
				"since", time.Since(v.Now).String(),
			)

			return err
		}
	}
}
