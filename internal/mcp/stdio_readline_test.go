package mcp

import (
	"bufio"
	"io"
	"strings"
	"testing"
)

func TestReadLine_OversizedLineIsSkippedNotFatal(t *testing.T) {
	input := "short\n" + strings.Repeat("x", 100) + "\nnext\r\nlast"
	r := bufio.NewReaderSize(strings.NewReader(input), 16)

	type res struct {
		line    string
		tooLong bool
		err     error
	}
	var got []res
	for {
		line, tooLong, err := readLine(r, 50)
		got = append(got, res{string(line), tooLong, err})
		if err != nil {
			break
		}
	}

	want := []res{
		{"short", false, nil},
		{"", true, nil},
		{"next", false, nil},
		{"last", false, nil},
		{"", false, io.EOF},
	}
	if len(got) != len(want) {
		t.Fatalf("got %d lines %+v, want %d", len(got), got, len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("line %d = %+v, want %+v", i, got[i], want[i])
		}
	}
}
