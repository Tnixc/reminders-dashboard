package utils

import (
	"fmt"
	"math"
	"strconv"
	"time"
)

const (
	ApproxDaysInYear  = 365
	ApproxDaysInMonth = 28
	DaysInWeek        = 7
)

func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func TimeElapsed(then time.Time) string {
	var parts []string
	var text string

	now := time.Now()
	diff := now.Sub(then)
	day := math.Round(diff.Hours() / 24)
	year := math.Round(day / ApproxDaysInYear)
	month := math.Round(day / ApproxDaysInMonth)
	week := math.Round(day / DaysInWeek)
	hour := math.Round(math.Abs(diff.Hours()))
	minute := math.Round(math.Abs(diff.Minutes()))
	second := math.Round(math.Abs(diff.Seconds()))

	if year > 0 {
		parts = append(parts, strconv.Itoa(int(year))+"y")
	}

	if month > 0 {
		parts = append(parts, strconv.Itoa(int(month))+"mo")
	}

	if week > 0 {
		parts = append(parts, strconv.Itoa(int(week))+"w")
	}

	if day > 0 {
		parts = append(parts, strconv.Itoa(int(day))+"d")
	}

	if hour > 0 {
		parts = append(parts, strconv.Itoa(int(hour))+"h")
	}

	if minute > 0 {
		parts = append(parts, strconv.Itoa(int(minute))+"m")
	}

	if second > 0 {
		parts = append(parts, strconv.Itoa(int(second))+"s")
	}

	if len(parts) == 0 {
		return "now"
	}

	return parts[0] + text
}

func BoolPtr(b bool) *bool { return &b }

func StringPtr(s string) *string { return &s }

func UintPtr(u uint) *uint { return &u }

func IntPtr(u int) *int { return &u }

func TimeUntil(then time.Time) string {
	now := time.Now()
	if then.Before(now) {
		return "overdue"
	}

	diff := then.Sub(now)
	days := int(diff.Hours() / 24)
	hours := int(diff.Hours()) % 24
	minutes := int(diff.Minutes()) % 60

	if days > 0 {
		if hours > 0 {
			return fmt.Sprintf("%dd %dh", days, hours)
		}
		return fmt.Sprintf("%dd", days)
	}

	if hours > 0 {
		if minutes > 0 {
			return fmt.Sprintf("%dh %dm", hours, minutes)
		}
		return fmt.Sprintf("%dh", hours)
	}

	if minutes > 0 {
		return fmt.Sprintf("%dm", minutes)
	}

	return "now"
}

func ShortNumber(n int) string {
	if n < 1000 {
		return strconv.Itoa(n)
	}

	if n < 1000000 {
		return strconv.Itoa(n/1000) + "k"
	}

	return strconv.Itoa(n/1000000) + "m"
}
