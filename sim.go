package main

import (
	"log"
	"math/rand"
)

type Simulation struct {
	TugCount     int
	TugSpeedLow  int
	TugSpeedHigh int

	ShipCount         int
	ShipContainerLow  int
	ShipContainerHigh int
	ShipTugsLow       int
	ShipTugsHigh      int
	ShipInterval      int

	BerthCount     int
	BerthDistances []int
	BerthCpus      []int

	Seed       int64
	dispatcher *Dispatcher
}

func NewSimulation(tugCount,
	tugSpeedLow,
	tugSpeedHigh,
	shipCount,
	shipContainerLow,
	shipContainerHigh,
	shipTugsHigh,
	shipInterval,
	berthCount int,
	seed int64,
	berthDistances,
	berthCpus []int,
	dispatcher *Dispatcher,
) Simulation {
	rand.Seed(seed)

	return Simulation{
		TugCount:          tugCount,
		TugSpeedLow:       tugSpeedLow,
		TugSpeedHigh:      tugSpeedHigh,
		ShipCount:         shipCount,
		ShipContainerLow:  shipContainerLow,
		ShipContainerHigh: shipContainerHigh,
		ShipTugsLow:       1,
		ShipTugsHigh:      shipTugsHigh,
		ShipInterval:      shipInterval,
		BerthCount:        berthCount,
		BerthDistances:    berthDistances,
		BerthCpus:         berthCpus,
		Seed:              seed,
		dispatcher:        dispatcher,
	}
}

// Concretize functions will generate the entities and events based on simulation parameters
func (s *Simulation) ConcretizeShips() {
	lastTime := 0
	for i := 0; i < s.ShipCount; i++ {
		interval := rand.Intn(s.ShipInterval)
		thisTime := lastTime + interval

		tc := rand.Intn(s.ShipTugsHigh) + 1
		cc := rand.Intn(s.ShipContainerHigh-s.ShipContainerLow+1) + s.ShipContainerLow
		sh := NewShip(i, tc, cc, s.dispatcher)
		s.dispatcher.AddShip(sh, thisTime)

		log.Printf("new ship: %+v\n", sh)

		lastTime = thisTime
	}
}

func (s *Simulation) ConcretizeTugs() {
	for i := 0; i < s.TugCount; i++ {
		ts := rand.Intn(s.TugSpeedHigh-s.TugSpeedLow) + s.TugSpeedLow
		t := NewTug(i, ts, s.dispatcher)
		s.dispatcher.AddTug(t, 0)
	}
}

func (s *Simulation) ConcretizeBerths() {
	for i := 0; i < s.BerthCount; i++ {
		b := NewBerth(i, s.BerthDistances[i], s.BerthCpus[i], s.dispatcher)
		s.dispatcher.AddBerth(b, 0)
	}
}

func (s *Simulation) ConcretizeAll() {
	s.ConcretizeShips()
	s.ConcretizeBerths()
	s.ConcretizeTugs()
}

// passthrough functions
func (s *Simulation) ProcessNextEvent() error {
	return s.dispatcher.ProcessNextEvent()
}

func (s *Simulation) EventCount() int {
	return len(s.dispatcher.Events)
}

// queries for UI
// TODO: for queries that may be against a very large list (i.e. ships)
//       calc and cache once per event in ProcessNextEvent

func (s Simulation) ShipsByState(state ShipState) int {
	count := 0
	for i := range s.dispatcher.Ships {
		if s.dispatcher.Ships[i].State == state {
			count++
		}
	}
	return count
}

func (s Simulation) ActiveShipCount() int {
	return s.ShipCount - s.ShipsByState(SS_GONE)
}

func (s Simulation) TugsByState(state TugState) int {
	count := 0
	for i := range s.dispatcher.Tugs {
		if s.dispatcher.Tugs[i].State == state {
			count++
		}
	}
	return count
}

func (s Simulation) BerthsByState(state BerthState) int {
	count := 0
	for i := range s.dispatcher.Berths {
		if s.dispatcher.Berths[i].State == state {
			count++
		}
	}
	return count
}
