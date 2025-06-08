package retry

import (
	"errors"
	"fmt"
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
		fmt.Println("Detected io.EOF — считаем временной ошибкой")
		return true
	}
	return (errors.As(err, &netErr) && netErr.Timeout()) || errors.As(err, &opErr)
}

func IsPGError(err error) bool {
	fmt.Printf("Type: %T\n", err)
	fmt.Printf("Value: %+v\n", err)

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
