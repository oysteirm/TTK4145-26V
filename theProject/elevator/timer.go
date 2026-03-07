package elevator

import "time"

const inactiveDuration = 10 * time.Second

func DoorTimer(
	start <-chan time.Duration,
	stop  <-chan struct{},
	timeout chan<- struct{},
	obstruction <- chan struct{},
	resetInactive <- chan struct{},
	setFunctional chan<- bool,
) {
	var timer *time.Timer
	var inactivityTimer *time.Timer = time.NewTimer(inactiveDuration)

	for {
		select {
		case d := <-start:
			if timer != nil {
				timer.Stop()
			}
			timer = time.NewTimer(d)

		case <-stop:
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

		case <- resetInactive:
			inactivityTimer.Stop()
			inactivityTimer = time.NewTimer(inactiveDuration)
		
		case <- inactivityTimer.C:
			setFunctional <- false
		}
	}
}
