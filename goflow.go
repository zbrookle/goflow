package main

import "fmt"
import "goflow/cron"

func main() {
	var orchestrator = new(cron.Orchestrator)
	fmt.Println(orchestrator)
}
