package utils

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type timeEncoding struct {
	year   int
	month  int
	day    int
	hour   int
	minute int
	second int
	milli  int
}

func DateTimeStringToUnixMilli(s string) (int, error) {
	n := len(s)
	enc := timeEncoding{
		year:  1970,
		month: 1,
		day:   1,
	}

	_, err := regexp.Match("\\d{4}-(0\\d|1[0-2])-(0[1-9]|[1-3]\\d)(T(0\\d|[1-2][0-4]):(0\\d|[1-5]\\d):(0\\d|[1-5]\\d))?(.\\d)?", []byte(s))
	if err != nil {
		return -1, errors.New("encountered invalid time format. Expected YYYY-MM-DDTHH:mm:ss.d, YYYY-MM-DDTHH:mm:ss or YYYY-MM-DD. Ensure that only numerical characters are given")
	}

	if n >= 10 {
		enc.year, _ = strconv.Atoi(s[0:4])
		enc.month, _ = strconv.Atoi(s[5:7])
		enc.day, _ = strconv.Atoi(s[8:10])
	}

	if n >= 19 {
		enc.hour, _ = strconv.Atoi(s[11:13])
		enc.minute, _ = strconv.Atoi(s[14:16])
		enc.second, _ = strconv.Atoi(s[17:19])
	}

	if n >= 21 {
		enc.milli, _ = strconv.Atoi(string([]byte{s[20]}))
		enc.milli *= 100
	}

	t := time.Date(
		enc.year,
		time.Month(enc.month),
		enc.day,
		enc.hour,
		enc.minute,
		enc.second,
		enc.milli*1000*1000,
		time.UTC,
	)

	return int(t.UnixMilli()), nil
}

func UnixMilliToDateTimeString(t int) string {
	dt := time.UnixMilli(int64(t)).UTC()

	year := strconv.FormatInt(int64(dt.Year()), 10)

	month := strconv.FormatInt(int64(dt.Month()), 10)
	if len(month) == 1 {
		month = "0" + month
	}

	day := strconv.FormatInt(int64(dt.Day()), 10)
	if len(day) == 1 {
		day = "0" + day
	}

	hour := strconv.FormatInt(int64(dt.Hour()), 10)
	if len(hour) == 1 {
		hour = "0" + hour
	}

	minute := strconv.FormatInt(int64(dt.Minute()), 10)
	if len(minute) == 1 {
		minute = "0" + minute
	}

	second := strconv.FormatInt(int64(dt.Second()), 10)
	if len(second) == 1 {
		second = "0" + second
	}

	var deciSeconds int = dt.Nanosecond() / 100000000

	dtString := []string{
		year,
		"-",
		month,
		"-",
		day,
		"T",
		hour,
		":",
		minute,
		":",
		second,
		".",
		strconv.FormatInt(int64(deciSeconds), 10),
	}

	return strings.Join(dtString, "")

}
