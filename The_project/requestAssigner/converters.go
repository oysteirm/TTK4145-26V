package requestAssigner
//import something to use System_Data_t??
import (
	"TTK4145-26V/messageSync"
)

func CC_ToBool(cc messageSync.CyclicCounter_t) bool {
	switch cc {
	case messageSync.CC_Confirmed, messageSync.CC_Done:
		return true
	default:
		return false
	}
}