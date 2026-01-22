package main

import "Driver-go/elevio"
import "fmt"

func main(){

    num_floors := 4
    num_elevators := 1

    e Elevator = elevator_uninitialized()

    elevio.Init("localhost:15657", numFloors)
    
    var d elevio.MotorDirection = elevio.MD_Up
    //elevio.SetMotorDirection(d)
    
    drv_buttons := make(chan elevio.ButtonEvent)
    drv_floors  := make(chan int)
    drv_obstr   := make(chan bool)
    drv_stop    := make(chan bool)    
    
    go elevio.PollButtons(drv_buttons)
    go elevio.PollFloorSensor(drv_floors)
    go elevio.PollObstructionSwitch(drv_obstr)
    go elevio.PollStopButton(drv_stop)
    
    
    for {
        select {
        case btn := <- drv_buttons:
            fsm_on_request_button_press(&e, btn)
            
        case floor := <- drv_floors:
            fsm_on_floor_arrival(&e, floor)
            
        /*    ????
        case obstr := <- drv_obstr:
            fsm_
        */    
        case stop := <- drv_stop:
            fmt.Printf("%+v\n", a)
            for f := 0; f < numFloors; f++ {
                for b := elevio.ButtonType(0); b < 3; b++ {
                    elevio.SetButtonLamp(b, f, false)
                }
            }
        }
    }    
}