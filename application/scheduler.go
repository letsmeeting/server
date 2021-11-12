package appmmain

import (
    . "github.com/jinuopti/lilpop-server/log"
)

// SchedulerMin 매분 00초 마다 호출된다
func SchedulerMin() {
    Logd("Timeout Minute scheduler")

    // Do something...
}

// SchedulerHour 매시 00분00초 마다 호출된다
func SchedulerHour() {
    Logd("Timeout Hour scheduler")

    // Do something...
}

// SchedulerDay 매일 00시00분00초 마다 호출된다
func SchedulerDay() {
    Logd("Timeout Day scheduler")

    // Log file rotate
    LogRotate()
}
