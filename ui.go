package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// styles

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Padding(5, 8)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Padding(2, 8)

	buttonStyle = lipgloss.NewStyle().Padding(0, 5)
	// less padding on focused style to avoid jumping when changing selection
	focusedStyle = lipgloss.NewStyle().Padding(0, 3).Bold(true)

	listTitleStyle = lipgloss.NewStyle().Padding(1, 5).Bold(true)
	listItemStyle  = lipgloss.NewStyle().Padding(1, 3)
	listStyle      = lipgloss.NewStyle().Margin(1).BorderStyle(lipgloss.NormalBorder())
)

type ModelState int

const (
	MENU ModelState = iota
	MAIN
	SETTINGS
	SHIP
	TUG
	BERTH
)

type model struct {
	state ModelState

	// menu state
	MenuChoice  int
	MenuButtons []string

	// main state
	Sim          Simulation
	SteppedCount int
}

func initModel() model {
	return model{
		state:        MENU,
		MenuChoice:   0,
		MenuButtons:  []string{"Start", "Configure"},
		Sim:          Simulation{},
		SteppedCount: 0,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		k := msg.String()
		if k == "q" || k == "esc" || k == "ctrl+c" {
			return m, tea.Quit
		}
	}

	switch m.state {
	case MENU:
		return updateMenu(m, msg)
	case MAIN:
		return updateMain(m, msg)
	}
	return m, nil
}

func (m model) View() string {
	switch m.state {
	case MENU:
		return viewMenu(m)
	case MAIN:
		return viewMain(m)
	}
	return ""
}

// menu update & subview

func updateMenu(m model, message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "left", "h":
			if m.MenuChoice == 1 {
				m.MenuChoice = 0
				return m, nil
			}
		case "right", "l":
			if m.MenuChoice == 0 {
				m.MenuChoice = 1
				return m, nil
			}
		case "enter":
			if m.MenuChoice == 0 {
				return m, StartDefault
			}
		}
	case StartMsg:
		m.Sim = Simulation(msg)
		m.state = MAIN
		return m, nil
	}
	return m, nil
}

func viewMenu(m model) string {
	s := titleStyle.Render("Port Simulator")

	s += "\n\n"

	for i := 0; i < 2; i++ {
		if m.MenuChoice == i {
			s += button(true, m.MenuButtons[i])
		} else {
			s += button(false, m.MenuButtons[i])
		}
	}

	return s
}

// main screen update & subview

func updateMain(m model, message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			return m, Step(m.Sim)
		}
	case StepMsg:
		if msg.Stepped {
			m.Sim = msg.Sim
			m.SteppedCount++
			return m, nil
		}
	}
	return m, nil
}

func viewMain(m model) string {
	s := headerStyle.Render("Port Simulator") + "\n\n"

	s += listItemStyle.Render(fmt.Sprintf("Enqueued Events: %d Step Count: %d", m.Sim.EventCount(), m.SteppedCount))

	shipsBlock := listTitleStyle.Render(fmt.Sprintf("Active Ships: %d", m.Sim.ActiveShipCount()))
	shipsBlock += listItemStyle.Render(
		fmt.Sprintf("WAITING: %d BERTHING: %d BERTHED: %d LAUNCHING: %d GONE: %d",
			m.Sim.ShipsByState(SS_WAITING),
			m.Sim.ShipsByState(SS_BERTHING),
			m.Sim.ShipsByState(SS_BERTHED),
			m.Sim.ShipsByState(SS_LAUNCHING),
			m.Sim.ShipsByState(SS_GONE),
		))

	s += listStyle.Render(shipsBlock)

	tugsBlock := listTitleStyle.Render(fmt.Sprintf("Available Tugs: %d", len(m.Sim.dispatcher.AvailableTugs)))
	tugsBlock += listItemStyle.Render(
		fmt.Sprintf("WAITING: %d ATTACHING: %d MOVING %d",
			m.Sim.TugsByState(TS_WAITING),
			m.Sim.TugsByState(TS_ATTACHING),
			m.Sim.TugsByState(TS_MOVING),
		))

	s += listStyle.Render(tugsBlock)

	berthsBlock := listTitleStyle.Render(fmt.Sprintf("Available Berths: %d", len(m.Sim.dispatcher.AvailableBerths)))
	berthsBlock += listItemStyle.Render(
		fmt.Sprintf("OPEN: %d RESERVED: %d OCCUPIED: %d",
			m.Sim.BerthsByState(BS_OPEN),
			m.Sim.BerthsByState(BS_RESERVED),
			m.Sim.BerthsByState(BS_OCCUPIED),
		))

	s += listStyle.Render(berthsBlock)
	return s
}

// Cmds and Msgs

type StartMsg Simulation
type StepMsg struct {
	Sim     Simulation
	Stepped bool
}

// start the simulation with default arguments

func StartDefault() tea.Msg {
	dispatcher := NewDispatcher()
	sim := NewSimulation(
		3,
		5,
		15,
		50,
		100,
		1000,
		2,
		100,
		3,
		123456789,
		[]int{100, 100, 300},
		[]int{25, 25, 50},
		&dispatcher,
	)

	sim.ConcretizeAll()
	return StartMsg(sim)
}

func Step(sim Simulation) tea.Cmd {
	return func() tea.Msg {
		stepped := false
		if e := sim.EventCount(); e > 0 {
			sim.ProcessNextEvent()
			stepped = true
		}

		return StepMsg{
			Sim:     sim,
			Stepped: stepped,
		}
	}
}

// utils

func button(focused bool, label string) string {
	if focused {
		return focusedStyle.Render(fmt.Sprintf("[ %s ]", label))
	} else {
		return buttonStyle.Render(label)
	}
}
