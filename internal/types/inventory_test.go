package types

import (
	"fmt"
	"testing"
)

func TestGet(t *testing.T) {
	inv := &ElaTargetInventory{}
	inv.Get("http://localhost:9200", "inventory-latest")

	data, _ := inv.JSON()
	fmt.Println(string(data))
}