package main

import (
	"bytes"
	"fmt"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"testing"
	"testing/quick"
)

func TestPack(t *testing.T) {
	dest := []byte{'d', 'e', 'f'}
	full := pack([]byte{'a', 'b', 'c'}, &dest)
	if full {
		t.Fatal("Should not be full")
	}
	expected := []byte{'d', 'e', 'f', '\n', 'a', 'b', 'c'}
	for i, val := range expected {
		if dest[i] != val {
			t.Fatal("Bad packing: Expected %s but was actually %s", val, dest[i])
		}
	}
}

func TestPackEmptyDest(t *testing.T) {
	dest := []byte{}
	full := pack([]byte{'a', 'b', 'c'}, &dest)
	if full {
		t.Error("Buffer full when it shouldn't be")
	}
	if !bytes.Equal(dest, []byte{'a', 'b', 'c'}) {
		t.Error("Bad packing of empty dest")
	}
}

func TestPackQuick(t *testing.T) {
	f := func(src, dest []byte) bool {
		expected := make([]byte, len(dest))
		copy(expected, dest)
		full := pack(src, &dest)
		if full {
			return true
		}
		if len(expected) == 0 {
			expected = src
		} else {
			expected = append(expected, '\n')
			expected = append(expected, src...)
		}
		return bytes.Equal(dest, expected)
	}

	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

func TestRecords(t *testing.T) {
	x := make([]*kinesis.PutRecordsRequestEntry, 10)
	fmt.Printf("%s", x)
}
