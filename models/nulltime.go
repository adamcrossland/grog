package model

import (
	"time"
)

// NullTime represents a time.Time object that can be explicitly unset.
type NullTime struct {
	Time time.Time
	Null bool
}

// IsNull returns the value of the Null property.
func (dbtime NullTime) IsNull() bool {
	return dbtime.Null
}

// Set assigns a value to the underlying time.Time and also sets NUll to false
func (dbtime *NullTime) Set(timeVal time.Time) {
	dbtime.Time = timeVal
	dbtime.Null = false
}

// Unix returns the Unix time value of the assigned time.Time.
func (dbtime NullTime) Unix() int64 {
	if !dbtime.IsNull() {
		return dbtime.Time.Unix()
	}

	return 0
}

// Val returns the value of the Time property
func (dbtime NullTime) Val() time.Time {
	return dbtime.Time
}

// String returns a string representation of the Time porperty.
func (dbtime NullTime) String() string {
	return dbtime.Time.String()
}
