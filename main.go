package main

import (
	"eventim-acc-gen/module"
	module_eventim "eventim-acc-gen/module/eventim"
	"log"
)

func main() {
	run := module_eventim.TestStructure{
		TestStructure: new(module.TestStructure),
	}
	err := run.GenerateEventimAcc()
	log.Println(err)
}
