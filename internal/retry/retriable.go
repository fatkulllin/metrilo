package retry

import (
	"errors"
	"io"
	"net"

	"github.com/fatkulllin/metrilo/internal/logger"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
)

func IsNetworkError(err error) bool {
	var netErr net.Error
	var opErr *net.OpError
	if errors.Is(err, io.EOF) {
		logger.Log.Warn("Detect io.EOF - is retriable")
		return true
	}
	return (errors.As(err, &netErr) && netErr.Timeout()) || errors.As(err, &opErr)
}

func IsPGError(err error) bool {
	var connErr *pgconn.ConnectError
	if errors.As(err, &connErr) {
		logger.Log.Warn("Detect pgconn.ConnectError - is retriable")
		return true
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		logger.Log.Warn("Detect pgconn.ConnectError - is retriable", zap.String("pgerr", pgErr.Code))
		return pgerrcode.IsConnectionException(pgErr.Code)
	}

	return false
}
