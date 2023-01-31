package hjscanner

import "testing"

func TestInstructioncompile(t *testing.T) {
	if data, err := InstructionOfMoveXY.compile(13.0, 23.0); err != nil {
		t.Fatal(err)
	} else if string(data) != "#*,move_xy,13.00,23.00,0,0,0,*#" {
		t.Fatalf("data:%s", data)
	}
	if data, err := InstructionOfMoveZ.compile(12.0); err != nil {
		t.Fatal(err)
	} else if string(data) != "#*,move_s,12.00,0,0,0,0,*#" {
		t.Fatalf("data:%s", data)
	}
}
