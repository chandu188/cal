package cal

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type clockTime struct {
	hh  int
	mm  int
	sec int
}

func (c *clockTime) After(ct clockTime) bool {
	ct1 := clockTimeToTime(*c)
	ct2 := clockTimeToTime(ct)
	return ct1.After(ct2)
}

func (c *clockTime) Before(ct clockTime) bool {
	ct1 := clockTimeToTime(*c)
	ct2 := clockTimeToTime(ct)
	return ct1.Before(ct2)
}

func clockTimeToTime(c clockTime) time.Time {
	layout := "2006-01-02T15:04:05.000Z"
	str := "2014-11-12T11:45:26.371Z"
	now, _ := time.Parse(layout, str)
	return time.Date(now.Year(), now.Month(), now.Day(),
		c.hh, c.mm, c.sec, 0, now.Location())
}

func (c clockTime) String() string {
	return fmt.Sprintf("%02d:%02d:%02d", c.hh, c.mm, c.sec)
}

func parseClockTime(clkTime string) (clockTime, error) {
	tokens := strings.Split(clkTime, ":")
	if len(tokens) != 3 {
		return clockTime{}, fmt.Errorf("%s is not a valid clock time", clkTime)
	}
	hh, err := strconv.ParseUint(tokens[0], 10, 0)
	if err != nil || hh < 0 || hh > 23 {
		return clockTime{}, fmt.Errorf("%s has an invalid hour", clkTime)
	}
	mm, err := strconv.ParseUint(tokens[1], 10, 0)
	if err != nil || mm < 0 || mm > 59 {
		return clockTime{}, fmt.Errorf("%s has an invalid minute", clkTime)
	}
	sec, err := strconv.ParseUint(tokens[2], 10, 0)
	if err != nil || sec < 0 || sec > 59 {
		return clockTime{}, fmt.Errorf("%s has an invalid minute", clkTime)
	}

	return clockTime{
		hh:  int(hh),
		mm:  int(mm),
		sec: int(sec),
	}, nil
}

type businessHours struct {
	start clockTime
	end   clockTime
}

func (b *businessHours) duration() time.Duration {
	return b.remainingDuration(b.start)
}

func (b *businessHours) remainingDuration(ct clockTime) time.Duration {
	var duration time.Duration

	if ct.After(b.end) {
		return duration
	}

	if b.start.After(ct) {
		ct = b.start
	}

	start := clockTimeToTime(ct)
	end := clockTimeToTime(b.end)
	return end.Sub(start)

}

func (b *businessHours) durationTo(ct clockTime) time.Duration {
	var duration time.Duration

	if ct.Before(b.start) {
		return duration
	}

	if ct.After(b.end) {
		ct = b.end
	}
	end := clockTimeToTime(ct)
	start := clockTimeToTime(b.start)
	return end.Sub(start)

}

func (b *businessHours) withInBusinessHours(date time.Time) bool {
	ctStart := time.Date(date.Year(), date.Month(), date.Day(), b.start.hh, b.start.mm, b.start.sec, 0, date.Location())
	ctEnd := time.Date(date.Year(), date.Month(), date.Day(), b.end.hh, b.end.mm, b.end.sec, 0, date.Location())
	return date.After(ctStart) && date.Before(ctEnd)
}

type Workday struct {
	working bool
	hrs     []businessHours
}

//NewWorkday returns workday
func NewWorkday(working bool) *Workday {

	w := &Workday{}
	w.working = working
	return w

}

// AddBusinessHours adds business hours to a workday
func (w *Workday) AddBusinessHours(start, end string) error {
	if !w.working {
		return errors.New("cannot add business hours on a non working day")
	}
	startClkTime, err := parseClockTime(start)
	if err != nil {
		return err
	}

	endClkTime, err := parseClockTime(end)
	if err != nil {
		return err
	}
	err = w.addBusinessHours(startClkTime, endClkTime)
	if err != nil {
		return err
	}
	return nil
}

func (w *Workday) addBusinessHours(start, end clockTime) error {

	startTime := time.Date(0, 0, 0, start.hh, start.mm, start.sec, 0, time.UTC)
	endTime := time.Date(0, 0, 0, end.hh, end.mm, end.sec, 0, time.UTC)

	if !startTime.Before(endTime) {
		return fmt.Errorf("start time %s is after end time %s", start, end)
	}

	if w.hrs == nil {
		w.hrs = make([]businessHours, 0)
	}
	w.hrs = append(w.hrs, businessHours{start: start, end: end})
	return nil
}

func (w *Workday) isWorking(date time.Time) bool {
	// if !w.working {
	// 	return false
	// }

	// min := clockTime{hh: 24, mm: 59, sec: 59}
	// max := clockTime{}
	// for _, b := range w.hrs {
	// 	if b.start.Before(min) {
	// 		min = b.start
	// 	}
	// 	if b.end.After(max) {
	// 		max = b.end
	// 	}
	// 	if b.withInBusinessHours(date) {
	// 		return true
	// 	}
	// }

	// if len(w.hrs) > 0 {
	// 	b := businessHours{start: min, end: max}
	// 	return b.withInBusinessHours(date)
	// }

	return w.working
}

func (w *Workday) duration() time.Duration {

	if len(w.hrs) == 0 && w.working {
		return 24 * time.Hour
	}
	var dur time.Duration
	for _, bhr := range w.hrs {
		dur += bhr.duration()
	}
	return dur
}

func (w *Workday) durationTo(ct clockTime) time.Duration {
	duration := time.Duration(0)
	for _, bhr := range w.hrs {
		duration += bhr.durationTo(ct)
	}
	return duration
}

func (w *Workday) SetWorkDay(working bool) {
	w.working = working
	if !working {
		w.hrs = nil
	}
}

func (w *Workday) remainingDuration(ct clockTime) time.Duration {
	duration := time.Duration(0)
	for _, bhr := range w.hrs {
		duration += bhr.remainingDuration(ct)
	}
	return duration
}
