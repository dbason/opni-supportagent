package errors

import (
	"errors"
	"fmt"
)

var (
	ErrQueueDelete      = errors.New("failed to queue delete")
	ErrInvalidDist      = errors.New("distribution must be one of rke, rke2, k3s")
	ErrInvalidArguments = errors.New("invalid arguments")
)

func ErrQueueDeleteWithResp(resp string) error {
	return fmt.Errorf("%s: %w", resp, ErrQueueDelete)
}

func ErrInvalidArgumentNumber(numRequired int) error {
	return fmt.Errorf("command requires %d arguments: %w", numRequired, ErrInvalidArguments)
}
