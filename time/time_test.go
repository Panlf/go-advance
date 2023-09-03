package time

import (
	"testing"
	"time"
)

func TestDate(t *testing.T) {
	//	Year: "2006" "06"
	//	Month: "Jan" "January" "01" "1"
	//	Day of the week: "Mon" "Monday"
	//	Day of the month: "2" "_2" "02"
	//	Day of the year: "__2" "002"
	//	Hour: "15" "3" "03" (PM or AM)
	//	Minute: "4" "04"
	//	Second: "5" "05"
	//	AM/PM mark: "PM"

	date := time.Now().Format("2006-01-02 15:04:05")

	t.Log(date)
}
