package util

import (
	"fmt"
	"time"
)

func FormatDuration(d time.Duration) string {
	d = d.Round(time.Minute)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute

	return fmt.Sprintf("%dh%dm", h, m)
}
