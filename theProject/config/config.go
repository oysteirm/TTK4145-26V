package config

import "time"

const (
	N_ELEVATORS int = 2
	N_FLOORS int = 4
	N_BUTTONS int = 3
	N_UP_DOWN int = 2
	DOOR_OPEN_DURATION time.Duration = 3 * time.Second
	IS_FUNCTIONAL_TIMER_DURATION time.Duration = 9 * time.Second

	B_CAST_PORT int = 20014
	PEER_UPDATE_PORT int = 20015


	B_CAST_PERIOD time.Duration = 1000 * time.Millisecond

	PP_PORT = 30000
	PP_SERVER_IP = "127.0.0.1"
	PP_TIMEOUT = 100 * time.Millisecond
	PP_INTERVAL = 10 * time.Millisecond
)