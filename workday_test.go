package cal

import (
	"testing"
	"time"
)

func TestParseClockTime(t *testing.T) {
	tcs := []struct {
		clkTime string
		err     bool
	}{
		{
			clkTime: "11:00:23",
			err:     false,
		},
		{
			clkTime: "25:00:23",
			err:     true,
		},
		{
			clkTime: "11:69:23",
			err:     true,
		},
		{
			clkTime: "11:34:67",
			err:     true,
		},
	}

	for _, tc := range tcs {
		_, err := parseClockTime(tc.clkTime)
		if (err != nil) != tc.err {
			t.Errorf("expected %t while parsing %s", tc.err, tc.clkTime)
		}
	}
}

func TestClockTimeString(t *testing.T) {
	ct := clockTime{
		hh:  8,
		mm:  10,
		sec: 5,
	}
	exp := "08:10:05"
	if exp != ct.String() {
		t.Errorf("got: %s, want: %s", ct.String(), exp)
	}
}

func TestClockTimeAfter(t *testing.T) {
	tcs := []struct {
		ct1 clockTime
		ct2 clockTime
		res bool
	}{
		{
			ct1: clockTime{hh: 10, mm: 10, sec: 10},
			ct2: clockTime{hh: 11, mm: 10, sec: 10},
			res: false,
		},
		{
			ct1: clockTime{hh: 10, mm: 55, sec: 59},
			ct2: clockTime{hh: 10, mm: 45, sec: 59},
			res: true,
		},
		{
			ct1: clockTime{hh: 10, mm: 45, sec: 49},
			ct2: clockTime{hh: 10, mm: 45, sec: 29},
			res: true,
		},
		{
			ct1: clockTime{hh: 11, mm: 47, sec: 26},
			ct2: clockTime{hh: 15, mm: 45, sec: 26},
			res: false,
		},
	}

	for _, tc := range tcs {
		if tc.res != tc.ct1.After(tc.ct2) {
			t.Errorf("got: %t; want: %t", tc.ct1.After(tc.ct2), tc.res)
		}
	}

}

func TestClockTimeToTime(t *testing.T) {
	ct := clockTime{
		hh:  23,
		mm:  55,
		sec: 50,
	}
	t1 := clockTimeToTime(ct)
	hh, mm, sec := t1.Clock()
	if hh != ct.hh {
		t.Errorf("got: %d; want: %d", hh, ct.hh)
	}

	if mm != ct.mm {
		t.Errorf("got: %d; want: %d", mm, ct.mm)
	}

	if sec != ct.sec {
		t.Errorf("got: %d; want: %d", sec, ct.sec)
	}
}
func TestBusinessHoursDuration(t *testing.T) {
	layout := "2006-01-02T15:04:05.000Z"
	str := "2014-11-12T11:45:26.371Z"
	now, _ := time.Parse(layout, str)

	tcs := []struct {
		unit       time.Duration
		multiplier int
		res        time.Duration
	}{
		{unit: time.Hour, multiplier: 4, res: 4 * time.Hour},
		{unit: time.Minute, multiplier: 3, res: 3 * time.Minute},
		{unit: time.Second, multiplier: 1, res: 1 * time.Second},
	}

	for _, tc := range tcs {
		hh, mm, sec := now.Clock()
		now2 := now.Add(tc.unit * time.Duration(tc.multiplier))
		hh2, mm2, sec2 := now2.Clock()
		b := businessHours{
			start: clockTime{hh: hh, mm: mm, sec: sec},
			end:   clockTime{hh: hh2, mm: mm2, sec: sec2},
		}
		dur := b.duration()
		if dur != tc.res {
			t.Errorf("got: %d, exepected %d", dur, tc.res)
		}

	}

}

func TestBusinessHoursRemaining(t *testing.T) {
	layout := "2006-01-02T15:04:05.000Z"
	str := "2014-11-12T11:45:26.371Z"
	now, _ := time.Parse(layout, str)
	hh, mm, sec := now.Clock()
	startClkTime := clockTime{hh: hh, mm: mm, sec: sec}
	hh, mm, sec = now.Add(4 * time.Hour).Clock()
	endClkTime := clockTime{hh: hh, mm: mm, sec: sec}
	b := businessHours{
		start: startClkTime,
		end:   endClkTime,
	}

	tcs := []struct {
		from time.Time
		res  time.Duration
	}{
		{
			from: now.Add(-4 * time.Hour),
			res:  4 * time.Hour,
		},
		{
			from: now.Add(2 * time.Hour),
			res:  2 * time.Hour,
		},
		{
			from: now.Add(120 * time.Second),
			res:  238 * time.Minute,
		},
	}

	for _, tc := range tcs {
		hh, mm, sec := tc.from.Clock()
		ct := clockTime{hh: hh, mm: mm, sec: sec}
		res := b.remainingDuration(ct)
		if res != tc.res {
			t.Errorf("got: %d; want: %d", res, tc.res)
		}
	}
}
func TestWithInBusinessHours(t *testing.T) {
	layout := "2006-01-02T15:04:05.000Z"
	str := "2014-11-12T11:45:26.371Z"
	now, _ := time.Parse(layout, str)
	hh, mm, sec := now.Clock()
	startClkTime := clockTime{hh: hh, mm: mm, sec: sec}
	hh, mm, sec = now.Add(4 * time.Hour).Clock()
	endClkTime := clockTime{hh: hh, mm: mm, sec: sec}

	b := businessHours{
		start: startClkTime,
		end:   endClkTime,
	}

	withIn := now.Add(2 * time.Hour)
	res := b.withInBusinessHours(withIn)
	if !res {
		t.Errorf("%s is not withInBusinessHours for the hours %s and %s ", withIn, startClkTime, endClkTime)
	}
	outSide := now.Add(6 * time.Hour)
	res = b.withInBusinessHours(outSide)
	if res {
		t.Errorf("%s is withInBusinessHours for the hours %s and %s ", outSide, startClkTime, endClkTime)
	}
}

func TestWithInBusinessHours2(t *testing.T) {
	layout := "2006-01-02T15:04:05.000Z"
	str := "2014-11-12T11:45:26.371Z"
	now, _ := time.Parse(layout, str)
	hh, mm, sec := now.Clock()
	startClkTime := clockTime{hh: hh, mm: mm, sec: sec}
	hh, mm, sec = now.Add(4 * time.Hour).Clock()
	endClkTime := clockTime{hh: hh, mm: mm, sec: sec}
	w := Workday{
		working: true,
	}
	w.addBusinessHours(startClkTime, endClkTime)

	hh, mm, sec = now.Add(6 * time.Hour).Clock()
	startClkTime = clockTime{hh: hh, mm: mm, sec: sec}
	hh, mm, sec = now.Add(9 * time.Hour).Clock()
	endClkTime = clockTime{hh: hh, mm: mm, sec: sec}
	w.addBusinessHours(startClkTime, endClkTime)

	res := w.isWorking(now.Add(5 * time.Hour))
	if !res {
		t.Errorf("got: %t; want: %t", res, true)
	}

	res = w.isWorking(now.Add(10 * time.Hour))
	if !res {
		t.Errorf("got: %t; want: %t", res, true)
	}

}

func TestNewWorkday(t *testing.T) {
	tcs := []struct {
		isWorking bool
		startDate string
		endDate   string
		err       bool
	}{
		{
			isWorking: false,
			startDate: "11:34:67",
			endDate:   "11:34:55",
			err:       true,
		},
		{
			isWorking: true,
			startDate: "08:00:00",
			endDate:   "17:00:00",
			err:       false,
		},

		{
			isWorking: true,
			startDate: "18:00:00",
			endDate:   "17:00:00",
			err:       true,
		},
	}

	for _, tc := range tcs {
		wd, err := NewWorkday(tc.isWorking, tc.startDate, tc.endDate)
		if (err != nil) != tc.err {
			t.Errorf("expected error to be %t, but recieved %t", tc.err, err != nil)
		}
		if err != nil {
			continue
		}
		if wd.working != tc.isWorking {
			t.Errorf("expected %t for working, but recieved %t", tc.isWorking, wd.working)
		}

		if len(wd.hrs) != 1 {
			t.Errorf("expected 1 entry for businesshours, but has %d", len(wd.hrs))
		}
	}
}

func TestWorkdayDuration(t *testing.T) {
	w := Workday{}
	w.working = true
	w.AddBusinessHours("08:00:00", "12:00:00")
	w.AddBusinessHours("13:00:00", "18:00:00")
	d := w.duration()
	durationInHrs := d / time.Hour
	if durationInHrs != 9 {
		t.Errorf("duration isn't 10 hrs")
	}
}

func TestWorkdayEmptyDuration(t *testing.T) {
	w := Workday{}
	w.working = true
	d := w.duration()
	durationInHrs := d / time.Hour
	if durationInHrs != 24 {
		t.Errorf("duration isn't 24 hrs")
	}
}

func TestSetWorkDay(t *testing.T) {
	w := Workday{working: true}
	w.AddBusinessHours("08:00:00", "12:00:00")
	w.AddBusinessHours("13:00:00", "18:00:00")
	if len(w.hrs) != 2 {
		t.Errorf("got:%d; want: 2", len(w.hrs))
	}
	w.SetWorkDay(false)
	if w.working != false {
		t.Errorf("got %t; want: false", w.working)
	}
	if len(w.hrs) != 0 {
		t.Errorf("got %d; want: 0, when is working is set to false", len(w.hrs))
	}
}

func TestGetRemainingDuration(t *testing.T) {
	w := Workday{working: true}
	w.AddBusinessHours("08:00:00", "12:00:00")
	w.AddBusinessHours("13:00:00", "17:00:00")
	tcs := []struct {
		ct  string
		dur time.Duration
	}{
		{ct: "11:00:00",
			dur: 5 * time.Hour,
		},
		{
			ct:  "07:00:00",
			dur: 8 * time.Hour,
		},
		{
			ct:  "18:00:00",
			dur: 0 * time.Hour,
		},
	}

	for _, tc := range tcs {
		ct, err := parseClockTime(tc.ct)
		if err != nil {
			t.Errorf("got unexpeced error %s", err)
		}
		d := w.remainingDuration(ct)
		if d != tc.dur {
			t.Errorf("got: %d; want: %d", d, tc.dur)
		}
	}

}
