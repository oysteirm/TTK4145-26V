package config

import "time"

const (
	N_ELEVATORS int = 3
	N_FLOORS int = 4
	N_BUTTONS int = 3
	N_UP_DOWN int = 2
	DOOR_OPEN_DURATION time.Duration = 3 * time.Second
	IS_FUNCTIONAL_TIMER_DURATION time.Duration = 9 * time.Second
	const B_CAST_PORT int = 20014 //MAKE THE RIGHT
	const PEER_UPDATE_PORT = 20015
)

