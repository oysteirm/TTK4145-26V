package requestassigner
//import something to use System_Data_t??

func CC_To_Bool(CC Cyclic_Counter_t)bool{
	if CC == CC_Uninit || CC == CC_No || CC = CC_Unconfirmed {
		return false
	}
	if CC == CC_Confirmed || CC == CC_Done {
		return true
	}
	else {
		return nil
	}
}
