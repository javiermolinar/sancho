package view

import (
	"strconv"
	"time"

	"github.com/javiermolinar/sancho/internal/task"
)

// HeaderLabels builds column labels and marks today's column.
func HeaderLabels(weekStart time.Time, today time.Time) ([]string, map[int]bool) {
	labels := make([]string, 0, 8)
	todayCols := make(map[int]bool)

	yearSuffix := weekStart.Year() % 100
	monthLabel := weekStart.Format("Jan") + " " + strconv.Itoa(yearSuffix/10) + strconv.Itoa(yearSuffix%10)
	labels = append(labels, monthLabel)

	for i := 0; i < 7; i++ {
		dayDate := weekStart.AddDate(0, 0, i)
		dayName := task.WeekdayShortName(i)
		dayNum := dayDate.Day()
		label := dayName + " " + strconv.Itoa(dayNum)
		if sameDay(dayDate, today) {
			label = "*" + label + "*"
			todayCols[i+1] = true
		}
		labels = append(labels, label)
	}

	return labels, todayCols
}

func sameDay(a, b time.Time) bool {
	ya, ma, da := a.Date()
	yb, mb, db := b.Date()
	return ya == yb && ma == mb && da == db
}
