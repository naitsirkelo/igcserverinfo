// The iso8601 package formats time.Duration in the ISO 8601 format.
package iso8601

import (
	"fmt"
	"time"
)

const (
	runePrefix = 'P'
	runeTime   = 'T'
	runeYear   = 'Y'
	runeMonth  = 'M'
	runeWeek   = 'W'
	runeDay    = 'D'
	runeHour   = 'H'
	runeMinute = 'M'
	runeSecond = 'S'

	multMinute = 60
	multHour   = multMinute * 60
	multDay    = multHour * 24
	multWeek   = multDay * 7
	multYear   = multDay * 365.25 // have to get years first to account for leap years
	multMonth  = multYear / 12    // once we have years we use that to get months
)

// Returns a ISO 8601 formatted string from a time.Duration instance.
func Format(d time.Duration) string {
	str := fmt.Sprintf("%c", runePrefix)
	sec := d.Seconds()

	years := int(sec / multYear)
	if years >= 1 {
		str += fmt.Sprintf("%d%c", years, runeYear)
		sec -= float64(years * multYear)
	}

	months := int(sec / multMonth)
	if months >= 1 {
		str += fmt.Sprintf("%d%c", months, runeMonth)
		sec -= float64(months * multMonth)
	}

	weeks := int(sec / multWeek)
	if weeks >= 1 {
		str += fmt.Sprintf("%d%c", weeks, runeWeek)
		sec -= float64(weeks * multWeek)
	}

	days := int(sec / multDay)
	if days >= 1 {
		str += fmt.Sprintf("%d%c", days, runeDay)
		sec -= float64(days * multDay)
	}

	if sec >= 1 {
		// there's still time left, add the time designator
		str += fmt.Sprintf("%c", runeTime)
	}

	hours := int(sec / multHour)
	if hours >= 1 {
		str += fmt.Sprintf("%d%c", hours, runeHour)
		sec -= float64(hours * multHour)
	}

	minutes := int(sec / multMinute)
	if minutes >= 1 {
		str += fmt.Sprintf("%d%c", minutes, runeMinute)
		sec -= float64(minutes * multMinute)
	}

	if sec >= 1 {
		str += fmt.Sprintf("%d%c", int(sec), runeSecond)
	}

	if len(str) == 1 {
		// at least one element must be present, we'll use seconds
		str += fmt.Sprintf("%c%d%c", runeTime, 0, runeSecond)
	}

	return str
}
