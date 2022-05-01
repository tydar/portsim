package main

import (
	"errors"
	"fmt"
	"log"
)

func main() {
	dispatch := NewDispatcher()
	s1 := NewShip(1, 1, 100, &dispatch)
	s2 := NewShip(2, 1, 100, &dispatch)
	dispatch.AddShip(s1, 0)
	fmt.Println(dispatch.Ships)

	t1 := NewTug(1, 5, &dispatch)
	dispatch.AddTug(t1, 15)

	dispatch.AddShip(s2, 200)

	fmt.Println(dispatch.Tugs)
	fmt.Println(dispatch.Events)

	for e := dispatch.Events; len(e) != 0; e = dispatch.Events {
		dispatch.ProcessNextEvent()
	}

	fmt.Printf("Ship berthing? %t \n", dispatch.Ships[1].State == SS_BERTHING)
	fmt.Printf("Tug moving? %t \n", dispatch.Tugs[1].State == TS_MOVING)
	fmt.Printf("arrival queue: %+v\n", dispatch.ArrivalsQueue)
}

// shared types
type Event struct {
	Time       int
	Type       int
	Payload    int
	TargetType int
	TargetID   int
}

// target types
const (
	TGT_SHIP = iota
	TGT_TUG
	TGT_BERTH
	TGT_DISPATCHER
)

type Entity interface {
	ProcessEvent(e Event)
}

// dispatcher
type Dispatcher struct {
	Ships           map[int]*Ship  // map of all ships keyed by ID
	Tugs            map[int]*Tug   // map of all tugs keyed by ID
	Berths          map[int]*Berth // map of all berths keyed by ID
	Events          []Event        // main event queue
	ArrivalsQueue   []int          // queue of ship IDs waiting to be berthed
	DeparturesQueue []int
	AvailableTugs   []int
	LastEventTime   int
}

// dispatcher events
const (
	D_SHIP_ARRIVAL = iota
	D_SHIP_BERTHING
	D_SHIP_LAUNCHING
	D_TUG_AVAILABLE
)

func NewDispatcher() Dispatcher {
	return Dispatcher{
		Ships:           make(map[int]*Ship),
		Tugs:            make(map[int]*Tug),
		Berths:          make(map[int]*Berth),
		Events:          make([]Event, 0),
		ArrivalsQueue:   make([]int, 0),
		DeparturesQueue: make([]int, 0),
		LastEventTime:   0,
	}
}

func (d *Dispatcher) AddShip(s Ship, arrivalTime int) error {
	_, prs := d.Ships[s.ID]
	if prs {
		return fmt.Errorf("ship with ID already exists: %d", s.ID)
	}

	d.Ships[s.ID] = &s

	d.AddEvent(Event{
		TargetType: TGT_DISPATCHER,
		TargetID:   1,
		Payload:    s.ID,
		Type:       D_SHIP_ARRIVAL,
		Time:       arrivalTime,
	})

	return nil
}

func (d *Dispatcher) AddTug(t Tug, arrivalTime int) error {
	_, prs := d.Tugs[t.ID]
	if prs {
		return fmt.Errorf("tug with that ID already exists: %d", t.ID)
	}

	d.Tugs[t.ID] = &t
	d.AddEvent(Event{
		TargetType: TGT_DISPATCHER,
		TargetID:   1,
		Payload:    t.ID,
		Type:       D_TUG_AVAILABLE,
		Time:       arrivalTime,
	})

	return nil
}

func (d *Dispatcher) AddEvent(e Event) {
	for i := range d.Events {
		// if the event's time is prior to the event at the current index
		// we place this event in that index
		log.Printf("iteration %d of AddEvent\n", i)
		if e.Time < d.Events[i].Time {
			d.Events = append(d.Events[:i+1], d.Events[i:]...)
			d.Events[i] = e
			log.Printf("Event added at index %d: %+v", i, e)
			return
		}
	}

	// if we didn't find a place for this already, add it to the end
	// also hit this branch if the queue is empty
	d.Events = append(d.Events, e)
	log.Printf("Event added to end: %+v", e)
}

func (d *Dispatcher) ProcessNextEvent() error {
	log.Println("processing event")
	if len(d.Events) == 0 {
		// no event to process; just return
		return errors.New("no events in queue")
	}

	e := d.Events[0]
	d.Events = d.Events[1:]
	d.LastEventTime = e.Time

	switch e.TargetType {
	case TGT_DISPATCHER:
		log.Println("dispatch event")
		d.ProcessEvent(e)
	case TGT_SHIP:
		log.Println("ship event")
		// process ship events
		s, prs := d.Ships[e.TargetID]
		if !prs {
			return fmt.Errorf("no such ship %d. full event: %+v", e.TargetID, e)
		}

		s.ProcessEvent(e)
	case TGT_TUG:
		// process tug events
		log.Printf("tug event: %+v", e)
		t, prs := d.Tugs[e.TargetID]
		if !prs {
			return fmt.Errorf("no such tug %d. full event %+v", e.TargetID, e)
		}

		t.ProcessEvent(e)
	}

	return nil
}

func (d *Dispatcher) ProcessEvent(e Event) {
	switch e.Type {
	case D_SHIP_ARRIVAL:
		d.ArrivalsQueue = append(d.ArrivalsQueue, e.Payload)
		tugsCount := len(d.AvailableTugs)
		index := 0
		if tugsCount > 0 {
			// dispatch appropriate number of tugs based on the ship's needs & capacity
			for i := 0; i < tugsCount && i < d.Ships[e.Payload].TugSize; i++ {
				d.AddEvent(Event{
					Type:       T_SHIP,
					TargetType: TGT_TUG,
					TargetID:   d.AvailableTugs[i],
					Payload:    e.Payload,
					Time:       e.Time,
				})
				index = i
			}
		}
		d.AvailableTugs = d.AvailableTugs[index+1:]

		log.Printf("D_SHIP_ARRIVAL processed: ship %d", e.Payload)

	case D_SHIP_LAUNCHING:
		index := -1
		for i := range d.DeparturesQueue {
			if d.DeparturesQueue[i] == e.Payload {
				index = i
				break
			}
		}

		tugs := d.Ships[e.Payload].AttachedTugs
		speed := 0
		for i := range tugs {
			speed += d.Tugs[tugs[i]].Speed
		}

		speed = speed / len(tugs)
		timeUnderway := 1000 / speed // TODO: update to calculate based on distances btwn tugs & berths / ocean

		for i := range tugs {
			d.AddEvent(Event{
				Type:       T_GO,
				TargetType: TGT_TUG,
				TargetID:   tugs[i],
				Payload:    timeUnderway,
				Time:       e.Time,
			})
		}

		d.AddEvent(Event{
			Type:       S_BERTH,
			TargetType: TGT_SHIP,
			TargetID:   e.Payload,
			Payload:    0,
			Time:       e.Time + timeUnderway,
		})

		if index+1 < len(d.DeparturesQueue) {
			d.DeparturesQueue = append(d.DeparturesQueue[:index], d.DeparturesQueue[index+1:]...)
		} else {
			d.DeparturesQueue = d.DeparturesQueue[:index]
		}

	case D_SHIP_BERTHING:
		index := -1
		for i := range d.ArrivalsQueue {
			if d.ArrivalsQueue[i] == e.Payload {
				index = i
				break
			}
		}

		tugs := d.Ships[e.Payload].AttachedTugs
		speed := 0
		for i := range tugs {
			speed += d.Tugs[tugs[i]].Speed
		}

		speed = speed / len(tugs)
		timeUnderway := 1000 / speed // TODO: update to calculate based on distances btwn tugs & berths / ocean

		for i := range tugs {
			d.AddEvent(Event{
				Type:       T_GO,
				TargetType: TGT_TUG,
				TargetID:   tugs[i],
				Payload:    timeUnderway,
				Time:       e.Time,
			})
		}

		d.AddEvent(Event{
			Type:       S_BERTH,
			TargetType: TGT_SHIP,
			TargetID:   e.Payload,
			Payload:    0,
			Time:       e.Time + timeUnderway,
		})

		if index+1 < len(d.ArrivalsQueue) {
			d.ArrivalsQueue = append(d.ArrivalsQueue[:index], d.ArrivalsQueue[index+1:]...)
		} else {
			d.ArrivalsQueue = d.ArrivalsQueue[:index]
		}

		log.Printf("D_SHIP_BERTHING processed: ship %d. New arrivals queue: %v\n", e.Payload, d.ArrivalsQueue)

	case D_TUG_AVAILABLE:
		// check departure queue to free a berth first
		depCount, arrCount := len(d.DeparturesQueue), len(d.ArrivalsQueue)

		if depCount > 0 {
			s := d.Ships[d.DeparturesQueue[0]]
			d.AddEvent(Event{
				Type:       T_SHIP,
				TargetType: TGT_TUG,
				TargetID:   e.Payload,
				Payload:    s.ID,
				Time:       e.Time,
			})
		} else if arrCount > 0 {
			s := d.Ships[d.ArrivalsQueue[0]]
			d.AddEvent(Event{
				Type:       T_SHIP,
				TargetType: TGT_TUG,
				TargetID:   e.Payload,
				Payload:    s.ID,
				Time:       e.Time,
			})
		} else {
			d.AvailableTugs = append(d.AvailableTugs, e.Payload)
		}

		log.Printf("D_TUG_AVAILABLE processed: tug %d", e.Payload)
	}
}

// ship
type Ship struct {
	ID             int
	TugSize        int // number of tugs needed
	TugCount       int // number of tugs attached
	ContainerCount int // number of containers
	State          int
	AttachedTugs   []int
	Dispatcher     *Dispatcher
}

// default NewShip function
// State: waiting
func NewShip(id, tugSize, containerCount int, dispatcher *Dispatcher) Ship {
	return Ship{
		ID:             id,
		TugSize:        tugSize,
		TugCount:       0,
		ContainerCount: containerCount,
		State:          SS_WAITING,
		Dispatcher:     dispatcher,
		AttachedTugs:   make([]int, tugSize),
	}
}

// ship events
const (
	S_ATTACH = iota
	S_DETACH
	S_BERTH
	S_LAUNCH
)

// ship states
const (
	SS_BERTHING = iota
	SS_BERTHED
	SS_WAITING
)

func (s *Ship) ProcessEvent(e Event) {
	log.Printf("ship state at start of process: %d\n", s.State)
	switch e.Type {
	case S_ATTACH:
		s.AttachedTugs[s.TugCount] = e.Payload
		s.TugCount++
		if s.TugCount == s.TugSize {
			if s.State == SS_WAITING {
				s.Dispatcher.AddEvent(Event{
					TargetType: TGT_DISPATCHER,
					TargetID:   1,
					Type:       D_SHIP_BERTHING,
					Payload:    s.ID,
					Time:       e.Time,
				})
			} else if s.State == SS_BERTHED {
				s.Dispatcher.AddEvent(Event{
					TargetType: TGT_DISPATCHER,
					TargetID:   1,
					Type:       D_SHIP_LAUNCHING,
					Payload:    s.ID,
					Time:       e.Time,
				})
			}
			s.State = SS_BERTHING
		}

	case S_DETACH:
		s.TugCount--

	case S_BERTH:
		log.Println("S_BERTH event made it to ship")
		s.State = SS_BERTHED
	}
}

// tug
type Tug struct {
	ID         int
	Speed      int
	State      int
	Ship       int
	Dispatcher *Dispatcher
}

func NewTug(id, speed int, dispatcher *Dispatcher) Tug {
	return Tug{
		ID:         id,
		Speed:      speed,
		State:      TS_WAITING,
		Ship:       -1,
		Dispatcher: dispatcher,
	}
}

// tug events
const (
	T_SHIP = iota
	T_DETACH
	T_GO // payload must be time offset
)

// tug states
const (
	TS_WAITING = iota
	TS_TUGGING
	TS_MOVING
	TS_ATTACHING
)

func (t *Tug) ProcessEvent(e Event) {
	log.Printf("Tug state at start of process: %d\n", t.State)
	switch e.Type {
	case T_SHIP:
		log.Println("tug T_SHIP event")
		t.State = TS_ATTACHING
		t.Ship = e.Payload
		log.Printf("state after T_SHIP: %d", t.State)

		t.Dispatcher.AddEvent(Event{
			Time:       e.Time + (100 / t.Speed), // TODO: update this with duration calc based on distance
			Type:       S_ATTACH,
			TargetID:   e.Payload,
			TargetType: TGT_SHIP,
			Payload:    t.ID,
		})

	case T_GO:
		if t.State != TS_ATTACHING {
			// should never happen
			log.Printf("tug ordered to go when not attached: %+v", e)
		}

		t.State = TS_MOVING
		t.Dispatcher.AddEvent(Event{
			Time:       e.Time + e.Payload,
			Type:       T_DETACH,
			TargetID:   t.ID,
			TargetType: TGT_TUG,
		})
	case T_DETACH:
		if t.State != TS_MOVING {
			log.Println("tug ordered to detach when not attached")
		}
		t.State = TS_WAITING
		t.Dispatcher.AddEvent(Event{
			Time:       e.Time + (100 / t.Speed), // TODO: update this with calc based on distance
			Type:       D_TUG_AVAILABLE,
			TargetType: TGT_DISPATCHER,
			TargetID:   1,
			Payload:    t.ID,
		})
	default:
		log.Println("not implemented yet")
	}
}

// berth
type Berth struct{}
