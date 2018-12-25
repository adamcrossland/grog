package model

import (
	"time"
)

type NullTime struct {
	Time time.Time
	Null bool
}

func (dbtime NullTime) IsNull() bool {
	return dbtime.Null
}

func (dbtime *NullTime) Set(timeVal time.Time) {
	dbtime.Time = timeVal
	dbtime.Null = false
}

func (dbtime NullTime) Unix() int64 {
	if !dbtime.IsNull() {
		return dbtime.Time.Unix()
	}

	return 0
}

func (dbtime NullTime) String() string {
	return dbtime.Time.String()
}
