package main


import (
	
	"TTK4145-26V/Request_Assigner"
	"TTK4145-26V/Tests/Helper_functions_for_tests"
	
)


func main() {
	// 1) Making FAKE instance of System_Data_t for this test
	confirmed := test_helpers.Make_Fake_Confirmed_System_Data_t(4, 3)

	// 2) Generating RA_System_Data on the correct format for the assigner (treated as a "Black Box")
	ra := requestassigner.Generating_RA_System_Data(confirmed)

	// 3) Printing the generated RA_System_Data to see the format
	test_helpers.More_readable_print("Generated RA_System_Data", ra)

	// 4) Exectuing the "Black Box" for assigning requests
	requests := requestassigner.Assign_Requests(ra)

	// 5) Printing the result from the assigner (can compare with ra)
	test_helpers.More_readable_print("RA_Output from assigner", requests)


}