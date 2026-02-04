package elevator

import "time"

func DoorTimer(
	start <-chan time.Duration,
	stop  <-chan struct{},
	timeout chan<- struct{},
) {
	var timer *time.Timer

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
		}
	}
}
