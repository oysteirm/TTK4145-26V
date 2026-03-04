package main


import (
	"encoding/json"
	"fmt"
	"TTK4145-26V/message_sync"
	"TTK4145-26V/Request_Assigner"
	"TTK4145-26V/elevator"
	
	
)

func pretty_print(label string, v any) {
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


	// eksempel: 4 floors, 2 hall buttons
	confirmed.Hall_Request_Data = make([][2]message_sync.Request_Cyclic_Counter_t, 4)
	confirmed.Hall_Request_Data[1][0].Value = message_sync.CC_Confirmed // hall up i floor 2 (merk 0 indekserte lister)
	confirmed.Hall_Request_Data[2][1].Value = message_sync.CC_Confirmed // hall down i floor 3
	confirmed.Hall_Request_Data[2][0].Value = message_sync.CC_Confirmed // hall up i floor 3
	confirmed.Hall_Request_Data[0][0].Value = message_sync.CC_Confirmed // hall up i floor 1

	confirmed.Elevator_Data = make([]message_sync.Elevator_Data_t, 3)

	// Elevator 1
	confirmed.Elevator_Data[0].Id = 1
	confirmed.Elevator_Data[0].Is_Alive.Value = true
	confirmed.Elevator_Data[0].Is_Able.Value = true
	confirmed.Elevator_Data[0].Floor.Value = 0
	confirmed.Elevator_Data[0].Elevator_Behaviour.Value = elevator.EB_Idle
	confirmed.Elevator_Data[0].Motor_Direction.Value = elevator.MD_Stop
	confirmed.Elevator_Data[0].Cab_Requests = make([]message_sync.Request_Cyclic_Counter_t, 4)
	confirmed.Elevator_Data[0].Cab_Requests[2].Value = message_sync.CC_Confirmed //cab call i floor 3

	// Elevator 2 (ikke able -> filtreres bort)
	confirmed.Elevator_Data[1].Id = 2
	confirmed.Elevator_Data[1].Is_Alive.Value = true
	confirmed.Elevator_Data[1].Is_Able.Value = false
	confirmed.Elevator_Data[1].Floor.Value = 2
	confirmed.Elevator_Data[1].Cab_Requests = make([]message_sync.Request_Cyclic_Counter_t, 4)

	// Elevator 3
	confirmed.Elevator_Data[2].Id = 3
	confirmed.Elevator_Data[2].Is_Alive.Value = true
	confirmed.Elevator_Data[2].Is_Able.Value = true
	confirmed.Elevator_Data[2].Floor.Value = 2
	confirmed.Elevator_Data[2].Elevator_Behaviour.Value = elevator.EB_Moving
	confirmed.Elevator_Data[2].Motor_Direction.Value = elevator.MD_Down
	confirmed.Elevator_Data[2].Cab_Requests = make([]message_sync.Request_Cyclic_Counter_t, 4)
	confirmed.Elevator_Data[2].Cab_Requests[0].Value = message_sync.CC_Confirmed //cab call i floor 3



	// 2) Kjør heis system generatoren
	ra := requestassigner.Generating_RA_System_Data(confirmed)

	// 3) Print resultatet (det som sendes til assigner)
	pretty_print("RA_System_Data", ra)

	// 4) Kjør assigneren
	requests := requestassigner.Assign_Orders(ra)

	// 5) Print resultat fra assigneren
	pretty_print("RA_Output", requests)


}