// Code generated by "stringer -type=TugState"; DO NOT EDIT.

package main

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[TS_WAITING-0]
	_ = x[TS_TUGGING-1]
	_ = x[TS_MOVING-2]
	_ = x[TS_ATTACHING-3]
}

const _TugState_name = "TS_WAITINGTS_TUGGINGTS_MOVINGTS_ATTACHING"

var _TugState_index = [...]uint8{0, 10, 20, 29, 41}

func (i TugState) String() string {
	if i < 0 || i >= TugState(len(_TugState_index)-1) {
		return "TugState(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _TugState_name[_TugState_index[i]:_TugState_index[i+1]]
}