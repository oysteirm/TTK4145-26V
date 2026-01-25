package elevator

import "time"

var timer_end_time time.Time
var timer_active bool

func timer_start(duration time.Duration) {
	timer_end_time = time.Now().Add(duration)
	timer_active = true
}

func timer_stop() {
	timer_active = false
}

func timer_timedOut() bool {
	return timer_active && time.Now().After(timer_end_time)
}