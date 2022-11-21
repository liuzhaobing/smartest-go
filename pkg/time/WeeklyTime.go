package time

import (
	"fmt"
	"math"
	"time"
)

func WeeklyTime(m int) (weekData []WeekDateString) {
	l, _ := time.LoadLocation("Asia/Shanghai")

	now := time.Now()
	nowTime := now.Format("2006-01-02")
	lastMonthNow := now.AddDate(0, -m, 0)
	lastMonthNowTime := lastMonthNow.Format("2006-01-02")
	startTime, _ := time.ParseInLocation("2006-01-02", lastMonthNowTime, l)
	endTime, _ := time.ParseInLocation("2006-01-02", nowTime, l)

	data := GroupByWeekDate(startTime, endTime)
	for _, d := range data {
		weekData = append(weekData, WeekDateString{
			WeekTh:    d.WeekTh,
			StartTime: d.StartTime.Format("2006-01-02 15:04:05"),
			EndTime:   d.EndTime.Format("2006-01-02 15:04:05"),
		})
	}
	return
}

type WeekDateString struct {
	WeekTh    string
	StartTime string
	EndTime   string
}

//判断时间是当年的第几周

func WeekByDate(t time.Time) string {
	yearDay := t.YearDay()
	yearFirstDay := t.AddDate(0, 0, -yearDay+1)
	firstDayInWeek := int(yearFirstDay.Weekday())

	//今年第一周有几天
	firstWeekDays := 1
	if firstDayInWeek != 0 {
		firstWeekDays = 7 - firstDayInWeek + 1
	}
	var week int
	if yearDay <= firstWeekDays {
		week = 1
	} else {
		week = (yearDay-firstWeekDays)/7 + 2
	}
	return fmt.Sprintf("%d第%d周", t.Year(), week)
}

type WeekDate struct {
	WeekTh    string
	StartTime time.Time
	EndTime   time.Time
}

// GroupByWeekDate 将开始时间和结束时间分割为周为单位
func GroupByWeekDate(startTime, endTime time.Time) []WeekDate {
	weekDate := make([]WeekDate, 0)
	diffDuration := endTime.Sub(startTime)
	days := int(math.Ceil(float64(diffDuration/(time.Hour*24)))) + 1
	currentWeekDate := WeekDate{}
	currentWeekDate.WeekTh = WeekByDate(endTime)
	currentWeekDate.EndTime = endTime
	currentWeekDay := int(endTime.Weekday())
	if currentWeekDay == 0 {
		currentWeekDay = 7
	}
	currentWeekDate.StartTime = endTime.AddDate(0, 0, -currentWeekDay+1)
	nextWeekEndTime := currentWeekDate.StartTime
	weekDate = append(weekDate, currentWeekDate)
	for i := 0; i < (days-currentWeekDay)/7+1; i++ {
		weekData := WeekDate{}
		weekData.EndTime = nextWeekEndTime
		weekData.StartTime = nextWeekEndTime.AddDate(0, 0, -7)
		weekData.WeekTh = WeekByDate(weekData.StartTime)
		nextWeekEndTime = weekData.StartTime
		weekDate = append(weekDate, weekData)

	}
	return weekDate

}
