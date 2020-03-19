package utils

import (
	"fmt"
	"strconv"
	"strings"
)

// Standard precision amount
const Precision = 8

// Converts float to floating point int string with precision.
// E.g. 1.632
func FloatToFPString(a float64, prec int) (string, error) {
	str := strconv.FormatFloat(a, 'f', -1, 64)
	parts := strings.Split(str, ".")

	if len(parts) == 1 {
		for i := 0; i < prec; i++ {
			parts[0] += "0"
		}

		return parts[0], nil
	}

	if len(parts) > 2 {
		return "", fmt.Errorf("wrong floating point number %q", str)
	}

	res := parts[0]
	precpart := parts[1]

	if len(precpart) < prec {
		missedZeros := prec - len(precpart)
		for i := 0; i < missedZeros; i++ {
			precpart += "0"
		}
	} else if len(precpart) > prec {
		// cut
		precpart = precpart[:prec]
	}

	if res == "0" || res == "-0" {
		if res[0] == '-' {
			res = "-"
		} else {
			res = ""
		}

		// remove zeros
		for i := range precpart {
			if precpart[i] != '0' {
				precpart = precpart[i:]
				break
			}
		}
	}

	return res + precpart, nil
}
