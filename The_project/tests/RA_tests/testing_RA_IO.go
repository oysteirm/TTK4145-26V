package main


import (
	
	"TTK4145-26V/requestAssigner"
	"TTK4145-26V/Tests/helperFunctionsForTests"
	
)


func main() {
	// 1) Making FAKE instance of System_Data_t for this test
	confirmed := testHelpers.MakeFakeConfirmedSystemData(4, 3)

	// 2) Generating RA_SystemData on the correct format for the assigner (treated as a "Black Box")
	ra := requestAssigner.Generating_RA_SystemData(confirmed)

	// 3) Printing the generated RA_SystemData to see the format
	testHelpers.MoreReadablePrint_JSON("Generated RA_SystemData", ra)

	// 4) Exectuing the "Black Box" for assigning requests
	requests := requestAssigner.AssignRequests(ra)

	// 5) Printing the result from the assigner (can compare with ra)
	testHelpers.MoreReadablePrint_JSON("RA_Output from assigner", requests)


}