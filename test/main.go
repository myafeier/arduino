package main

import (
	"flag"
	"fmt"

	"github.com/myafeier/arduino/hjscanner"
)

func main() {
	cmd := flag.String("cmd", "", "input instruction you want!")

	flag.Parse()
	if *cmd == "" {
		panic("no param ")
	}
	err := hjscanner.InitDefaultScanner()
	if err != nil {
		panic(err)
	}
	switch *cmd {
	case "move":
		doMove()
	}
}

func doMove() {

	for {
		var x, y float32
		fmt.Println("input x:")
		fmt.Scanf("%f", &x)
		fmt.Println("input y:")
		fmt.Scanf("%f", &y)
		res, err := hjscanner.DefaultScaner.RunInstruction(hjscanner.InstructionOfMoveXY, x, y)
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Println("resp: ", res)
	}

}
