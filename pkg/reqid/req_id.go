package reqid

import (
	"context"

	"github.com/google/uuid"
)

func Gen() string {
	return uuid.New().String()
}

func ToContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, "req_id", Gen())
}

func FromContext(ctx context.Context) string {

	if reqId := ctx.Value("req_id"); reqId != nil {
		return reqId.(string)
	}

	return ""
}
