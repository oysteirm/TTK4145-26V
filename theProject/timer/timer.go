package timer

import (
	"theProject/config"
	"time"
)

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

	for {
		select {
		case <-doorStart:
			if timer != nil {
				timer.Stop()
			}
			timer = time.NewTimer(config.DOOR_OPEN_DURATION)

		case <-doorStop:
			if timer != nil {
				timer.Stop()
				timer = nil
			}

		case <-func() <-chan time.Time {
			if timer != nil {
				return timer.C
			}
			return nil
		}():
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
			if isFunctionalTimer != nil {
				return isFunctionalTimer.C
			}
			return nil
		}():
			isFunctionalTimeout <- struct{}{}
		}
	}
}
