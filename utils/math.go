package util

// return 1.0, 0, or -1.0
func Sign(v float64) float64 {
	if v > 0 {
		return 1.0
	} else if v < 0 {
		return -1.0
	} else {
		return 0
	}
}
