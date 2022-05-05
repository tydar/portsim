package main

import (
	"errors"
	"fmt"
	"log"
)

// DONE: implement berth event processing in dispatcher, including passing distance as tug event payload instead of time
// 	 and distance payload for tug detach commands

// TODO: ensure all ProcessEvents functions bubble up errors due to bad state transitions

func main() {
	dispatch := NewDispatcher()
	b1 := NewBerth(1, 2500, 100, &dispatch)
	b2 := NewBerth(2, 2500, 1000, &dispatch)
	s1 := NewShip(1, 1, 100, &dispatch)
	s2 := NewShip(2, 1, 100, &dispatch)

	dispatch.AddBerth(b1, 0)
	dispatch.AddBerth(b2, 0)
	dispatch.AddShip(s1, 10)

	t1 := NewTug(1, 5, &dispatch)
	dispatch.AddTug(t1, 100)

	dispatch.AddShip(s2, 200)

	for e := dispatch.Events; len(e) != 0; e = dispatch.Events {
		dispatch.ProcessNextEvent()
	}

	fmt.Printf("Ship berthed? %t \n", dispatch.Ships[1].State == SS_BERTHED)
	fmt.Printf("Tug moving? %t \n", dispatch.Tugs[1].State == TS_MOVING)
	fmt.Printf("arrival queue: %+v\n", dispatch.BerthQueue)
}

// TODO: implement simulation type and constructor

// shared types

type EventType int

type Event struct {
	Time       int
	Type       EventType
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
	ArrivalsQueue   []int
	BerthQueue      []int // queue of ship IDs waiting to be berthed
	DeparturesQueue []int
	AvailableTugs   []int
	AvailableBerths []int
	LastEventTime   int
	EventCount      int
}

// dispatcher events
//go:generate stringer -type=EventType
const (
	// dispatcher
	D_SHIP_ARRIVAL EventType = iota
	D_SHIP_BERTHING
	D_SHIP_LAUNCHING
	D_TUG_AVAILABLE
	D_BERTH_OCCUPIED
	D_BERTH_EMPTY
	D_SHIP_ASSIGNED
	D_UNLOADING_DONE

	//ship
	S_ATTACH
	S_DETACH
	S_BERTH
	S_LAUNCH
	S_ASSIGN

	//tug
	T_SHIP
	T_DETACH
	T_GO

	// berth
	B_DOCK
	B_UNDOCK
	B_RESERVE
)

func NewDispatcher() Dispatcher {
	return Dispatcher{
		Ships:           make(map[int]*Ship),
		Tugs:            make(map[int]*Tug),
		Berths:          make(map[int]*Berth),
		Events:          make([]Event, 0),
		ArrivalsQueue:   make([]int, 0),
		BerthQueue:      make([]int, 0),
		DeparturesQueue: make([]int, 0),
		LastEventTime:   0,
		EventCount:      0,
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

func (d *Dispatcher) AddBerth(b Berth, openTime int) error {
	_, prs := d.Berths[b.ID]
	if prs {
		return fmt.Errorf("berth with that ID already exists: %d", b.ID)
	}

	d.Berths[b.ID] = &b
	d.AddEvent(Event{
		Type:       D_BERTH_EMPTY,
		TargetType: TGT_DISPATCHER,
		TargetID:   1,
		Payload:    b.ID,
		Time:       openTime,
	})

	return nil
}

func (d *Dispatcher) AddEvent(e Event) {
	log.Printf("adding event: %+v", e)
	for i := range d.Events {
		// if the event's time is prior to the event at the current index
		// we place this event in that index
		if e.Time < d.Events[i].Time {
			d.Events = append(d.Events[:i+1], d.Events[i:]...)
			d.Events[i] = e
			return
		}
	}

	// if we didn't find a place for this already, add it to the end
	// also hit this branch if the queue is empty
	d.Events = append(d.Events, e)
}

func (d *Dispatcher) ProcessNextEvent() error {
	if len(d.Events) == 0 {
		// no event to process; just return
		return errors.New("no events in queue")
	}

	e := d.Events[0]
	d.Events = d.Events[1:]
	d.LastEventTime = e.Time
	d.EventCount++

	log.Printf("State: %+v\n", d)
	log.Printf("Event: %+v\n", e)

	switch e.TargetType {
	case TGT_DISPATCHER:
		d.ProcessEvent(e)
	case TGT_SHIP:
		// process ship events
		s, prs := d.Ships[e.TargetID]
		if !prs {
			return fmt.Errorf("no such ship %d. full event: %+v", e.TargetID, e)
		}

		s.ProcessEvent(e)
	case TGT_TUG:
		// process tug events
		t, prs := d.Tugs[e.TargetID]
		if !prs {
			return fmt.Errorf("no such tug %d. full event %+v", e.TargetID, e)
		}

		t.ProcessEvent(e)
	case TGT_BERTH:
		// process berth events
		b, prs := d.Berths[e.TargetID]
		if !prs {
			return fmt.Errorf("no such berth %d. full event %+v", e.TargetID, e)
		}

		return b.ProcessEvent(e)
	}

	return nil
}

func (d *Dispatcher) ProcessEvent(e Event) {
	switch e.Type {
	case D_SHIP_ARRIVAL:
		// DONE: rewrite to assign ship to berth & move to BerthQueue and fire D_SHIP_ASSIGNED event
		//       that event will then contain the tug assignment. If no Berth available instead put in
		//       ArrivalQueue, which will then be processed with D_BERTH_OPEN events.
		berthsCount := len(d.AvailableBerths)
		if berthsCount > 0 {
			d.AddEvent(Event{
				Type:       B_RESERVE,
				TargetType: TGT_BERTH,
				TargetID:   d.AvailableBerths[0],
				Payload:    e.Payload,
				Time:       e.Time - 1, // ensure this goes before any other attempts to reserve berths
			})

			d.AddEvent(Event{
				Type:       D_SHIP_ASSIGNED,
				TargetType: TGT_DISPATCHER,
				TargetID:   1,
				Time:       e.Time,
				Payload:    e.Payload,
			})

			d.AddEvent(Event{
				Type:       S_ASSIGN,
				TargetType: TGT_SHIP,
				TargetID:   e.Payload,
				Time:       e.Time,
				Payload:    d.AvailableBerths[0],
			})

			d.AvailableBerths = d.AvailableBerths[1:]
		} else {
			d.ArrivalsQueue = append(d.ArrivalsQueue, e.Payload)
		}

	case D_SHIP_ASSIGNED:
		d.BerthQueue = append(d.BerthQueue, e.Payload)
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
		d.AvailableTugs = d.AvailableTugs[index:]

	case D_SHIP_LAUNCHING:
		index := -1
		for i := range d.DeparturesQueue {
			if d.DeparturesQueue[i] == e.Payload {
				index = i
				break
			}
		}

		ship := d.Ships[e.Payload]
		tugs := ship.AttachedTugs
		speed := 0
		for i := range tugs {
			speed += d.Tugs[tugs[i]].Speed
		}

		distance := d.Berths[ship.Berth].Distance
		speed = speed / len(tugs)
		timeUnderway := distance / speed // DONE: update to calculate based on distances btwn tugs & berths / ocean

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
			Type:       S_LAUNCH,
			TargetType: TGT_SHIP,
			TargetID:   e.Payload,
			Payload:    ship.Berth,
			Time:       e.Time + timeUnderway,
		})

		d.AddEvent(Event{
			Type:       B_UNDOCK,
			TargetType: TGT_BERTH,
			TargetID:   ship.Berth,
			Payload:    0,
			Time:       e.Time + timeUnderway, // this makes assumptiont that berth isn't available until ship fully gone
		})

		if index+1 < len(d.DeparturesQueue) {
			d.DeparturesQueue = append(d.DeparturesQueue[:index], d.DeparturesQueue[index+1:]...)
		} else {
			d.DeparturesQueue = d.DeparturesQueue[:index]
		}

	case D_SHIP_BERTHING:
		index := -1
		for i := range d.BerthQueue {
			if d.BerthQueue[i] == e.Payload {
				index = i
				break
			}
		}

		tugs := d.Ships[e.Payload].AttachedTugs
		speed := 0
		for i := range tugs {
			speed += d.Tugs[tugs[i]].Speed
		}

		ship := d.Ships[e.Payload]
		distance := d.Berths[ship.Berth].Distance
		speed = speed / len(tugs)
		timeUnderway := distance / speed // DONE: update to calculate based on distances btwn tugs & berths / ocean

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

		if index+1 < len(d.BerthQueue) {
			d.BerthQueue = append(d.BerthQueue[:index], d.BerthQueue[index+1:]...)
		} else {
			d.BerthQueue = d.BerthQueue[:index]
		}

	case D_TUG_AVAILABLE:
		// check departure queue to free a berth first
		depCount, arrCount := len(d.DeparturesQueue), len(d.BerthQueue)

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
			// verify there's a berth
			s := d.Ships[d.BerthQueue[0]]
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

	case D_BERTH_EMPTY:
		arrCount := len(d.ArrivalsQueue)
		if arrCount > 0 {
			s := d.Ships[d.ArrivalsQueue[0]]

			d.AddEvent(Event{
				Type:       S_ASSIGN,
				TargetType: TGT_SHIP,
				TargetID:   s.ID,
				Payload:    e.Payload,
				Time:       e.Time,
			})

			d.AddEvent(Event{
				Type:       B_RESERVE,
				TargetType: TGT_BERTH,
				TargetID:   e.Payload,
				Payload:    s.ID,
				Time:       e.Time - 1,
			})

			d.AddEvent(Event{
				Type:       D_SHIP_ASSIGNED,
				TargetType: TGT_DISPATCHER,
				TargetID:   1,
				Payload:    s.ID,
				Time:       e.Time,
			})

			d.ArrivalsQueue = d.ArrivalsQueue[1:]
		} else {
			d.AvailableBerths = append(d.AvailableBerths, e.Payload)
		}
	case D_UNLOADING_DONE:
		d.DeparturesQueue = append(d.DeparturesQueue, e.Payload)
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
		d.AvailableTugs = d.AvailableTugs[index:]
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
	Berth          int
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
		Berth:          -1,
	}
}

// ship events
const ()

// ship states
const (
	SS_BERTHING = iota
	SS_BERTHED
	SS_WAITING
	SS_GONE
)

func (s *Ship) ProcessEvent(e Event) {
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
		s.State = SS_BERTHED
		s.Dispatcher.AddEvent(Event{
			Type:       B_DOCK,
			TargetType: TGT_BERTH,
			TargetID:   s.Berth,
			Payload:    s.ID,
			Time:       e.Time,
		})
	case S_ASSIGN:
		s.Berth = e.Payload
	case S_LAUNCH:
		s.State = SS_GONE
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
const ()

// tug states
const (
	TS_WAITING = iota
	TS_TUGGING
	TS_MOVING
	TS_ATTACHING
)

func (t *Tug) ProcessEvent(e Event) {
	switch e.Type {
	case T_SHIP:
		t.State = TS_ATTACHING
		t.Ship = e.Payload

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
			return
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
			return
		}
		t.State = TS_WAITING
		t.Dispatcher.AddEvent(Event{
			Time:       e.Time + (100 / t.Speed), // TODO: update this with calc based on distance
			Type:       D_TUG_AVAILABLE,
			TargetType: TGT_DISPATCHER,
			TargetID:   1,
			Payload:    t.ID,
		})
		t.Dispatcher.AddEvent(Event{
			Time:       e.Time,
			Type:       S_DETACH,
			TargetID:   t.Ship,
			TargetType: TGT_SHIP,
			Payload:    0,
		})
	default:
		log.Println("not implemented yet")
	}
}

// berth
type Berth struct {
	ID                int
	Distance          int // abstracted linear distance from entrance to port
	State             int
	ContainersPerUnit int // unloading / loading speed abstraction
	Dispatcher        *Dispatcher
}

func NewBerth(id, distance, cpu int, dispatcher *Dispatcher) Berth {
	return Berth{
		ID:                id,
		Distance:          distance,
		State:             BS_OPEN,
		Dispatcher:        dispatcher,
		ContainersPerUnit: cpu,
	}
}

// berth states
const (
	BS_OPEN = iota
	BS_CLOSED
	BS_OCCUPIED
	BS_RESERVED
)

func (b *Berth) ProcessEvent(e Event) error {
	switch e.Type {
	case B_DOCK:
		if b.State != BS_OPEN && b.State != BS_RESERVED {
			return fmt.Errorf("dock msg received by occupied dock ID %d", b.ID)
		}

		b.State = BS_OCCUPIED

		log.Println("about to add unloading done event")
		b.Dispatcher.AddEvent(Event{
			TargetType: TGT_DISPATCHER,
			TargetID:   1,
			Type:       D_UNLOADING_DONE,
			Payload:    e.Payload,
			Time:       e.Time + (b.Dispatcher.Ships[e.Payload].ContainerCount / b.ContainersPerUnit),
		})
	case B_UNDOCK:
		if b.State != BS_OCCUPIED {
			return fmt.Errorf("undock message received by unoccupied dock ID %d", b.ID)
		}

		b.State = BS_OPEN
		b.Dispatcher.AddEvent(Event{
			TargetType: TGT_DISPATCHER,
			TargetID:   1,
			Type:       D_BERTH_EMPTY,
			Payload:    b.ID,
			Time:       e.Time,
		})
	case B_RESERVE:
		if b.State != BS_OPEN {
			return fmt.Errorf("reserve message received by non-open dock ID %d", b.ID)
		}
		b.State = BS_RESERVED
	}
	return nil
}
