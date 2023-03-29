package main

import (
	"fmt"

	"github.com/myafeier/arduino/hjscanner"
	"github.com/myafeier/log"
)

func init() {
	log.SetLogLevel(log.DEBUG)
}

func main() {
	sn, err := hjscanner.InitDefaultScanner()
	if err != nil {
		panic(err)
	}
	fmt.Printf("device connected: %s \n", sn)

	var cmd string
	for {
		fmt.Println("input your cmd:")
		if _, err := fmt.Scanln(&cmd); err != nil {
			panic(err)
		}
		switch cmd {
		case "move":
			doMove()
		case "zoom":
			doZoom()
		case "pop":
			doPop()
		case "push":
			doPush()
		case "off":
			doLaseroff()
		}
	}

}
func doLaseroff() {
	hjscanner.DefaultScaner.RunInstruction(hjscanner.InstructionOfCloseLaser, "green")
}
func doPop() {
	hjscanner.DefaultScaner.RunInstruction(hjscanner.InstructionOfMoveOut)
}
func doPush() {

	hjscanner.DefaultScaner.RunInstruction(hjscanner.InstructionOfMoveIn)
}
func doZoom() {
	for {
		var x float32
		fmt.Println("input zoom in/out value:")
		fmt.Scanf("%f", &x)
		res, err := hjscanner.DefaultScaner.RunInstruction(hjscanner.InstructionOfMoveZ, x)
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Println("resp: ", res)
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
