package testHelpers

import(
	"encoding/json"
	"fmt"
)

func MoreReadablePrint_JSON(label string, v any) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Println("Marshal error:", err)
		return
	}
	fmt.Println("====", label, "====")
	fmt.Println(string(b))
}