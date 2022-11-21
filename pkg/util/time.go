package util

import (
	"database/sql/driver"
	"fmt"
	"smartest-go/pkg/util/const"
	"time"
)

// JSONTime format json time field by myself
type JSONTime struct {
	time.Time
}

// MarshalJSON on JSONTime format Time field with %Y-%m-%d %H:%M:%S
func (t JSONTime) MarshalJSON() ([]byte, error) {
	formatted := fmt.Sprintf("\"%s\"", t.Format(utilconst.TIME_TEMPLATE_1))
	return []byte(formatted), nil
}

func (t JSONTime) UnmarshalJSON(bytes []byte) error {
	parse, err := time.Parse(fmt.Sprintf("\"%s\"", utilconst.TIME_TEMPLATE_1), string(bytes))
	if err != nil {
		return err
	}

	t.Time = parse

	return nil
}

// Value insert timestamp into mysql need this function.
func (t JSONTime) Value() (driver.Value, error) {
	var zeroTime time.Time
	if t.Time.UnixNano() == zeroTime.UnixNano() {
		return nil, nil
	}
	return t.Time, nil
}

// Scan value of time.Time
func (t *JSONTime) Scan(v interface{}) error {
	value, ok := v.(time.Time)
	if ok {
		*t = JSONTime{Time: value}
		return nil
	}
	return fmt.Errorf("can not convert %v to timestamp", v)
}

func (t JSONTime) ToString() string {
	return t.Time.Format(utilconst.TIME_TEMPLATE_1)
}

func JSONTimeNow() JSONTime {
	return JSONTime{time.Now()}
}

func JSONTimeParse(timeString string) JSONTime {
	parse, _ := time.Parse(utilconst.TIME_TEMPLATE_1, timeString)
	return JSONTime{parse}
}

func YesterdayStartTimeStr() string {
	todayDate := time.Now().AddDate(0, 0, -1).Format(utilconst.TIME_TEMPLATE_3)
	return fmt.Sprintf("%s 00:00:00", todayDate)
}

func TodayStartTimeStr() string {
	todayDate := time.Now().Format(utilconst.TIME_TEMPLATE_3)
	return fmt.Sprintf("%s 00:00:00", todayDate)
}

func TomorrowStartTimeStr() string {
	tomorrowDate := time.Now().AddDate(0, 0, 1).Format(utilconst.TIME_TEMPLATE_3)
	return fmt.Sprintf("%s 00:00:00", tomorrowDate)
}

func YesterdayDateStr() string {
	return time.Now().AddDate(0, 0, -1).Format(utilconst.TIME_TEMPLATE_3)
}
