package kvstore

import (
	"testing"
)

func TestInMemoryStore_Add(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		key   string
		value string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TODO: construct the receiver type.
			var k InMemoryStore
			k.Add(tt.key, tt.value)
		})
	}
}
