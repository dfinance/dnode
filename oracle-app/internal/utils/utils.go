package utils

import (
	"math"
	"strconv"
)

// Standard precision amount
const Precision = 8

// Converts float to floating point int string with precision.
// E.g. 1.632
func FloatToFPString(v float64, p int) string {
	v *= math.Pow(10.0, float64(p))
	return strconv.FormatFloat(v, 'f', 0, 64)
}
