package main


import (
	"encoding/json"
	"fmt"
	"TTK4145-26V/message_sync"
	"TTK4145-26V/Request_Assigner"
	"TTK4145-26V/elevator"
	
	
)

func prettyPrint(label string, v any) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Println("Marshal error:", err)
		return
	}
	fmt.Println("====", label, "====")
	fmt.Println(string(b))
}

func main() {
	// 1) Lag et fake confirmed_system_data
	var confirmed message_sync.System_Data_t

	// TODO: Fyll inn confirmed med noe meningsfullt.
	// Denne delen må matches med hvordan System_Data_t ser ut hos dere.
	// Jeg legger inn en "template" under.

	// eksempel: 4 floors, 2 hall buttons
	confirmed.Hall_Request_Data = make([][2]message_sync.Request_Cyclic_Counter_t, 4)
	confirmed.Hall_Request_Data[1][0].Value = message_sync.CC_Confirmed // hall up i floor 1
	confirmed.Hall_Request_Data[3][1].Value = message_sync.CC_Confirmed // hall down i floor 3

	confirmed.Elevator_Data = make([]message_sync.Elevator_Data_t, 2)

		// Elevator 1
	confirmed.Elevator_Data[0].Id = 1
	confirmed.Elevator_Data[0].Is_Alive.Value = true
	confirmed.Elevator_Data[0].Is_Able.Value = true
	confirmed.Elevator_Data[0].Floor.Value = 0
	confirmed.Elevator_Data[0].Elevator_Behaviour.Value = elevator.EB_Idle
	confirmed.Elevator_Data[0].Motor_Direction.Value = elevator.MD_Stop
	confirmed.Elevator_Data[0].Cab_Requests = make([]message_sync.Request_Cyclic_Counter_t, 4)
	confirmed.Elevator_Data[0].Cab_Requests[2].Value = message_sync.CC_Confirmed

// Elevator 2 (ikke able -> filtreres bort)
	confirmed.Elevator_Data[1].Id = 2
	confirmed.Elevator_Data[1].Is_Alive.Value = true
	confirmed.Elevator_Data[1].Is_Able.Value = false
	confirmed.Elevator_Data[1].Floor.Value = 2
	confirmed.Elevator_Data[1].Cab_Requests = make([]message_sync.Request_Cyclic_Counter_t, 4)

	// 2) Kjør generatoren
	ra := requestassigner.Generating_RA_System_Data(confirmed)

	// 3) Print resultatet (det som sendes til assigner)
	prettyPrint("RA_System_Data", ra)
}