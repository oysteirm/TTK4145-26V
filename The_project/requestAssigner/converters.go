package requestAssigner
//import something to use System_Data_t??
import (
	"theProject/messageSync"
)

func CC_ToBool(cc messageSync.CyclicCounter_t) bool {
	switch cc {
	case messageSync.CC_Confirmed, messageSync.CC_Done:
		return true
	default:
		return false
	}
}