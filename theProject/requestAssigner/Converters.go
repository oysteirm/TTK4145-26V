package requestassigner
//import something to use System_Data_t??
import (
	"TTK4145-26V/message_sync"
)

func CC_To_Bool(cc message_sync.Cyclic_Counter_t) bool {
	switch cc {
	case message_sync.CC_Confirmed, message_sync.CC_Done:
		return true
	default:
		return false
	}
}