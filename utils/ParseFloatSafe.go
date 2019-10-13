package util

import (
	"bean/logger"
	"encoding/json"
	"fmt"
	"github.com/shopspring/decimal"
	"math"
	"reflect"
	"strconv"
)

func ParseFloatSafe64(s string) float64 {
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		logger.Error().Err(err).Msg("parsing string to float64")
		//		panic("parsing string to float64")
		return 0
	}
	return v
}

func SafeFloat64(d interface{}) float64 {
	var t1 float64
	switch d.(type) {
	case json.Number:
		t1, _ = d.(json.Number).Float64()
	case string:
		t1 = ParseFloatSafe64(d.(string))
	case float64:
		t1 = d.(float64)
	case decimal.Decimal:
		t1, _ = d.(decimal.Decimal).Float64()
	default:
		logger.Warn().Msg("SafeFloat64 conversion: do not know how to convert interfact to float64, " + fmt.Sprint(reflect.TypeOf(d)))
		t1 = math.NaN() //
	}
	return t1
}

func JsonNumberToFloatSafe64(j json.Number) float64 {
	v, err := j.Float64()
	if err != nil {
		logger.Error().Err(err).Msg("")
		//		panic("parsing Json number to float64")
		return 0
	}
	return v
}
