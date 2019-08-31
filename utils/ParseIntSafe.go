package util

import (
	"bean/logger"
	"strconv"
)

func ParseIntSafe64(s string) int64 {
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		logger.Error().Err(err).Msg("parsing string to float64")
		return 0
	}
	return v
}
