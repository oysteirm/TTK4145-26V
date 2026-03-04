package test_helpers

import(
	"encoding/json"
	"fmt"
)

func More_readable_print(label string, v any) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Println("Marshal error:", err)
		return
	}
	fmt.Println("====", label, "====")
	fmt.Println(string(b))
}