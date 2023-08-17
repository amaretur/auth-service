package handler

import (
	"net/http"

	"github.com/amaretur/auth-service/internal/errors"
	errutil "github.com/amaretur/auth-service/pkg/errors"

	"github.com/amaretur/auth-service/pkg/reqid"
	"github.com/amaretur/auth-service/pkg/log"
)

var defErrHttpMapper = map[uint32]int{
	errors.InvalidToken.TypeId: http.StatusForbidden,
}

func errToHttpResp(err error, mapper map[uint32]int) (int, string) {

	if errutil.Has(err, errors.Internal) {
		return http.StatusInternalServerError,
			http.StatusText(http.StatusInternalServerError)
	}

	code := mapper[errutil.TypeId(err)]

	if code == 0 {
		return http.StatusInternalServerError,
			http.StatusText(http.StatusInternalServerError)
	}

	return code, err.Error()
}

func logger(
	r *http.Request,
	logger log.Logger,
	responseData any,
) log.Logger {

	return logger.WithFields(map[string]any{
		"req_id": reqid.FromContext(r.Context()),
		"request": map[string]any{
			"url": r.RequestURI,
		},
		"response": responseData,
	})
}
