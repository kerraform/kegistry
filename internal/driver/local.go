package driver

import "go.uber.org/zap"

type Local struct {
	logger *zap.Logger
}

func newLocalDriver(logger *zap.Logger) (Driver, error) {
	return &Local{
		logger: logger,
	}, nil
}
