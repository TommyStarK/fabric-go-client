package fabclient

import (
	"bytes"
	"testing"
)

func TestConvertArrayOfStringsToArrayOfByteArrays(t *testing.T) {
	witness := [][]byte{
		[]byte("this"),
		[]byte("is"),
		[]byte("working"),
	}

	test := convertArrayOfStringsToArrayOfByteArrays([]string{"this", "is", "working"})

	if len(witness) != len(test) {
		t.Fail()
	}

	for i := range witness {
		if bytes.Compare(witness[i], test[i]) != 0 {
			t.Fatalf("should be %+v but got %+v", witness[i], test[i])
		}
	}
}
