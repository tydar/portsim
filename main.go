package main

import (
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

// DONE: implement berth event processing in dispatcher, including passing distance as tug event payload instead of time
// 	 and distance payload for tug detach commands

// TODO: ensure all ProcessEvents functions bubble up errors due to bad state transitions
// TODO: fix race condition: tugs can all be assigned without ships moving

func main() {
	f, err := os.OpenFile("logfile", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v\n", err)
	}
	defer f.Close()

	log.SetOutput(f)

	m := initModel()
	p := tea.NewProgram(m, tea.WithAltScreen())

	if err := p.Start(); err != nil {
		log.Fatal("could not start program:", err)
	}
	/*
		dispatch := NewDispatcher()
		berthDistances := []int{1000, 5000, 2500}
		berthCpus := []int{20, 20, 20}
		sim := NewSimulation(15, 5, 10, 5000, 100, 500, 2, 1000, 3, 834758239475, berthDistances, berthCpus, &dispatch)
		sim.ConcretizeAll()

		for e := sim.EventCount(); e != 0; e = sim.EventCount() {
			sim.ProcessNextEvent()
		}
	*/
}
