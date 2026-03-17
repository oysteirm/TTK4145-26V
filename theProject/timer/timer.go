package timer

import (
	"theProject/config"
	"time"
)

/*
-----------------------------------
Functionality:
	- Timers multiplexes two independent timers behind start/stop channels.
 		1. Door timer: started by doorStart, cancelled by doorStop, emits on doorTimeout.
  		2. Functional watchdog: started by isFunctionalStart, cancelled by isFunctionalStop,
		   emits on isFunctionalTimeout.
	- The function runs forever in one goroutine and serializes all timer state changes,
	  which avoids shared-memory races around timer pointers.
-----------------------------------
*/

func Timers(
	doorStart <-chan struct{},
	doorStop <-chan struct{},
	doorTimeout chan<- struct{},
	isFunctionalStart <-chan struct{},
	isFunctionalStop <-chan struct{},
	isFunctionalTimeout chan<- struct{},
) {
	
	var timer *time.Timer
	var isFunctionalTimer *time.Timer
	
	// Nil means "inactive" for that timer.
	for {
		select {
		case <-doorStart:
			if timer != nil {
				timer.Stop()
			}
			timer = time.NewTimer(config.DOOR_OPEN_DURATION)

		case <-doorStop:
			// Explicitly disable the door timer.
			if timer != nil {
				timer.Stop()
				timer = nil
			}

		case <-func() <-chan time.Time {
			// If timer is nil, this case is skipped.
			if timer != nil {
				return timer.C
			}
			return nil
		}():
			// Timer expired: clear it and notify timeout.
			timer = nil
			doorTimeout <- struct{}{}

		case <-isFunctionalStart:
			if isFunctionalTimer != nil {
				isFunctionalTimer.Stop()
			}
			isFunctionalTimer = time.NewTimer(config.IS_FUNCTIONAL_TIMER_DURATION)

		case <-isFunctionalStop:
			if isFunctionalTimer != nil {
				isFunctionalTimer.Stop()
				isFunctionalTimer = nil
			}

		case <-func() <-chan time.Time {
			// If timer is nil, this case is skipped.
			if isFunctionalTimer != nil {
				return isFunctionalTimer.C
			}
			return nil
		}():
			// Timer expired: clear it and notify timeout.
			isFunctionalTimer = nil
			isFunctionalTimeout <- struct{}{}
		}
	}
}
