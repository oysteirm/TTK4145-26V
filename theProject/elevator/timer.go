package elevator

import "time"

const inactiveDuration = 9 * time.Second

func Timers(
	doorStart <-chan time.Duration,
	doorStop  <-chan struct{},
	doorTimeout chan<- struct{},
	obstruction <- chan struct{},
	inactiveStart <- chan struct{},
	inactiveStop <- chan struct{},
	setFunctional chan<- bool,
) {
	var timer *time.Timer
	var inactivityTimer *time.Timer = time.NewTimer(inactiveDuration)

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
			timeout <- struct{}{}
		
		case <-obstruction:
			setFunctional <- false

		case <-inactiveStart:
			if inactivityTimer != nil {
				inactivityTimer.Stop()
			}
			inactivityTimer = time.NewTimer(inactiveDuration)

		case <- inactiveStop:
			if inactivityTimer != nil {
				inactivityTimer.Stop()
				inactivityTimer = nil
			}

		case <-func() <- chan time.Time {
			if inactivityTimer != nil {
				return inactivityTimer.C
			}
			return nil
		}():
			inactivityTimer = nil
			setFunctional <- false
		}
	}
}
