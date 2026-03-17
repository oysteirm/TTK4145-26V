# TTK4145 - Elevator Project:

Distributed elevator control project for Real-time Programming TTK4145 at NTNU

## Overview
This configuration is set for:
- 3 elevators
- 4 floors
- 3 button types: hall up, hall down, cab

These values are defined in theProject/config/config.go

## Architecture
**Modules**
- elevator_IO: communication between elevator hardware and software
- elevatorServer: local elevator logic and wires the subsystems together
- fsm: elevator state machine and event handling
- msgSynsServer: distributed state synchronization and button event propagation
- elevatorStateGuardian: central owner of elevator state and assigned requests
- processPairs: primary-backup process pair logic
- requestAssigner: hall request assigner, algorithm fetched from <https://github.com/TTK4145/Project-recources/tree/master/elev_algo>
- networkDriver: peer discovery and broadcast transport, code provided by <https://github.com/TTK4145/driver-go>

## Prerequisites
- Go 1.25
- An elevator simulator
- or an elevator server that uses TCP endpoints such as localhost:15657

## Repositary Layout
- theProject/: Go module and source code
- SimElevatorServer: prebuilt simulator, fetched from <https://github.com/TTK4145/Simulator-V2>

## Build
Build the main application form the Go module root:
```
cd theProject
go build .
```
this produces the elevator node executable from theProject/main.go

## Run
Start one process per elevator

Example with two elevators:

```
cd theProject
go run main.go -id=0 -addr=localhost:15657
```
```
cd theProject
go run main.go -id=1 -addr=localhost:15658
```
if ```-addr``` is excluded, the program defaults to:
- elevator n: ```localhost:15658 + n```

Available flags:
- ```-id```: local elevator id
- ```-addr```: TCP address for the elevator I/O server

## Startup Sequence
1. Start the elevator simulator or elevator server for each elevator
2. Run one Go process per elevator with a unique ```-id```
3. Verify that peers are discovered

## Configuration
Key constants are defined in theProject/config/config.go:

## Known Limitations

## Suggested Next Improvements