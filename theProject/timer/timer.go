package timer

import (
	"time"
	"theProject/config"
)

/*
-----------------------------------
Functionality:
	- Timers multiplexes two independent timers behind start/stop channels.
 		1. Door timer: started by doorStart, cancelled by doorStop, emits on doorTimeout.
  		2. Functional watchdog: started by isFunctionalStart, cancelled by isFunctionalStop,
		   emits on isFunctionalTimeout.
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
	
	var doorTimer *time.Timer
	var doorC <-chan time.Time
	
	var isFunctionalTimer *time.Timer
	var isFunctionalC <-chan time.Time
	
	for {
		select {
		case <-doorStart:
			if doorTimer != nil {
				doorTimer.Stop()
			}
			doorTimer = time.NewTimer(config.DOOR_OPEN_DURATION)
			doorC = doorTimer.C

		case <-doorStop:
			if doorTimer != nil {
				doorTimer.Stop()
				doorTimer = nil
				doorC = nil
			}

		case <-doorC:
			// Timer expired: clear it and notify timeout
			doorTimer = nil
			doorC = nil
			doorTimeout <- struct{}{}

		case <-isFunctionalStart:
			if isFunctionalTimer != nil {
				isFunctionalTimer.Stop()
			}
			isFunctionalTimer = time.NewTimer(config.IS_FUNCTIONAL_TIMER_DURATION)
			isFunctionalC = isFunctionalTimer.C

		case <-isFunctionalStop:
			if isFunctionalTimer != nil {
				isFunctionalTimer.Stop()
				isFunctionalTimer = nil
				isFunctionalC = nil
			}

		case <-isFunctionalC:
			// Timer expired: clear it and notify timeout.
			isFunctionalTimer = nil
			isFunctionalC = nil
			isFunctionalTimeout <- struct{}{}
		}
	}
}
