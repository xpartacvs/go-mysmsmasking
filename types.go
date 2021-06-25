package mysmsmasking

import (
	"regexp"
	"time"
)

type Status int8

type AccountInfo struct {
	Balance uint32
	Expiry  time.Time
}

type Airwaybill struct {
	Id        string
	Timestamp time.Time
}

const (
	FAILED Status = iota
	SENT
	DELIVERED
	INVALID_ID
	INVALID_MSISDN
	BALANCE_INSUFFICIENT
	BALANCE_EXPIRED
)

var rgxCSVSeparator *regexp.Regexp = regexp.MustCompile(`,\s*`)
