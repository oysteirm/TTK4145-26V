package elevator

import "time"

// InitTimers creates and initializes all timers for the elevator system
// Drains initial channels to ensure clean state
func InitTimers() *time.Timer {
	doorTimer := time.NewTimer(0 * time.Second)
	<-doorTimer.C // Drain channel
	return doorTimer
}

// StartTimer safely starts a timer with the given duration
// Drains any pending signal before creating new timer
func StartTimer(timer *time.Timer, duration time.Duration) *time.Timer {
	if !timer.Stop() {
		<-timer.C
	}
	return time.NewTimer(duration)
}

// StopTimer safely stops a timer
// Drains any pending signal
func StopTimer(timer *time.Timer) *time.Timer {
	if !timer.Stop() {
		<-timer.C
	}
	return timer
}
