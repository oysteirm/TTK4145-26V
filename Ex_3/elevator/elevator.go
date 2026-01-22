package elevator

import (
	"fmt"
	"Driver-go/elevio"
)

type Elevator struct {
	id int
	floor int
	motor_direction MotorDirection
	obstruction bool
	door bool 
	next_floor int
}

// -1:unknown	0: no order	1:uncomfirmed	2:confirmed
type Order int

var 

