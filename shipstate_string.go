// Code generated by "stringer -type=ShipState"; DO NOT EDIT.

package main

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[SS_BERTHING-0]
	_ = x[SS_BERTHED-1]
	_ = x[SS_WAITING-2]
	_ = x[SS_LAUNCHING-3]
	_ = x[SS_GONE-4]
}

const _ShipState_name = "SS_BERTHINGSS_BERTHEDSS_WAITINGSS_LAUNCHINGSS_GONE"

var _ShipState_index = [...]uint8{0, 11, 21, 31, 43, 50}

func (i ShipState) String() string {
	if i < 0 || i >= ShipState(len(_ShipState_index)-1) {
		return "ShipState(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _ShipState_name[_ShipState_index[i]:_ShipState_index[i+1]]
}