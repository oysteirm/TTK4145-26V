package theproject

import(
	"theProject/fsm"
	"theProject/messageSync"
	"theProject/timer"
	"theProject/config"
)

func main(){
	//TODO: make process pairs with primary-backup topology

	//TODO: get the local ID from terminal command
	localID := //hjelp henning

	//TODO: initialize all variables and constants
	//TODO: initialize all the channels
	//elevatorServer channels
	elevatorDataToMsgSync := make(chan messageSync.ElevatorData_t)
    requestToMsgSync := make(chan messageSync.RequestCyclicCounter_t)
	systemDataFromMsgSync := make(chan messageSync.SystemData_t)
	//messageSync channels
	elevatorDataFromFSM := make(chan ElevatorData_t) 
	hallRequests_CC_FromFSM := make(chan RequestCyclicCounter_t)
	dataToFSM := make(SystemData_t)		
	peersReciever := make(chan peers.PeerUpdate)

	

	//TODO: launch the go routines 

	//TODO: forever loop?
}