package utility

import (
    "time"
    "github.com/go-co-op/gocron"
)

func GocronScheduler(m func(), h func(), d func()) bool {
    if m == nil && h == nil && d == nil {
        return false
    }
    now := time.Now()
    sched := gocron.NewScheduler(time.Local)

    if d != nil {
        _, _ = sched.Every(1).Days().Tag("Days").At("00:00:00").Do(d)
    }
    if h != nil {
        nextHour := now.Truncate(time.Hour)
        _, _ = sched.Every(1).Hours().Tag("Hours").StartAt(nextHour).Do(h)
    }
    if m != nil {
        nextMin := now.Truncate(time.Minute)
        _, _ = sched.Every(1).Minutes().Tag("Minutes").StartAt(nextMin).Do(m)
    }

    sched.StartAsync()

    return true
}

func Scheduler(m func(), h func(), d func()) {
    go SchedulerMin(m)
    go SchedulerHour(h)
    go SchedulerDay(d)
}

func SchedulerMin(min func()) {
    if min == nil {
        return
    }
    first := true
    var now time.Time
    var next time.Time
    var prevMin int
    var currMin int
    for {
        now = time.Now()
        prevMin = now.Minute()
        next = now.Truncate(time.Minute)
        next = next.Add(time.Minute * 1)
        if !first && (prevMin > currMin) {
            goto Run
        }
        first = false
        // fmt.Printf("Min Next: %s, Duration: %v\n", next.String(), next.Sub(now))

        time.Sleep(next.Sub(now))

        now = time.Now()
        currMin = now.Minute()
        if currMin == prevMin {
            continue
        }
        Run:
            min()
    }
}

func SchedulerHour(hour func()) {
    if hour == nil {
        return
    }
    first := true
    var now time.Time
    var next time.Time
    var prevHour int
    var currHour int
    for {
        now = time.Now()
        prevHour = now.Hour()
        next = now.Truncate(time.Hour)
        next = next.Add(time.Hour * 1)
        if !first && (prevHour > currHour) {
            goto Run
        }
        first = false
        // fmt.Printf("Min Next: %s, Duration: %v\n", next.String(), next.Sub(now))

        time.Sleep(next.Sub(now))

        now = time.Now()
        currHour = now.Hour()
        if currHour == prevHour {
            continue
        }
        Run:
            hour()
    }
}

func SchedulerDay(day func()) {
    if day == nil {
        return
    }
    first := true
    var now time.Time
    var next time.Time
    var prevDay int
    var currDay int
    for {
        now = time.Now()
        prevDay = now.Day()
        next = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
        next = next.AddDate(0, 0, 1)
        if !first && (prevDay > currDay) {
            goto Run
        }
        first = false
        // fmt.Printf("Min Next: %s, Duration: %v\n", next.String(), next.Sub(now))

        time.Sleep(next.Sub(now))

        now = time.Now()
        currDay = now.Day()
        if currDay == prevDay {
            continue
        }
        Run:
            day()
    }
}
