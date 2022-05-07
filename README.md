# PortSim

A discrete event simulation model for port traffic in a containerized shipping port.

Each entity is modeled with a FSM to manage its internal state. An event bus in the top-level dispatcher allows processing events in order in a next-event time progression.

TODO:

1) Add logic for availability of berths [DONE]
2) Add logic for calculating timing based on the speed of tugs & distance to/from berth to staging areas. [DONE]
3) Add logic for loading/unloading. [NEXT]
4) Implement logic for individual container tracking.
5) Provide a terminal UI for monitoring and interacting with the simulation.
6) Implement Stringer interface generation for enums; factor states and message type codes into their own types to facilitate pretty printing. [DONE]
