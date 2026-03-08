package timer

import (
	"time"
	"theProject/config"
)

func Timers(
	doorStart <-chan time.Duration,
	doorStop  <-chan struct{},
	doorTimeout chan<- struct{},
	obstruction <- chan struct{},
	isFunctionalStart <- chan struct{},
	isFunctionalStop <- chan struct{},
	setFunctional chan<- bool,
) {
	var timer *time.Timer
	var isFunctionalTimer *time.Timer = time.NewTimer(IS_FUNCTIONAL_TIMER_DURATION)

	for {
		select {
		case d := <-doorStart:
			if timer != nil {
				timer.Stop()
			}
			timer = time.NewTimer(d)

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
		
		case <-obstruction:
			setFunctional <- false

		case <-isFunctionalStart:
			if isFunctionalTimer != nil {
				isFunctionalTimer.Stop()
			}
			isFunctionalTimer = time.NewTimer(config.IS_FUNCTIONAL_TIMER_DURATION)

		case <- isFunctionalStop:
			if isFunctionalTimer != nil {
				isFunctionalTimer.Stop()
				isFunctionalTimer = nil
			}

		case <-func() <- chan time.Time {
			if isFunctionalTimer != nil {
				return isFunctionalTimer.C
			}
			return nil
		}():
			isFunctionalTimer = nil
			setFunctional <- false
		}
	}
}
