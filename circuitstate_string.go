// Code generated by "stringer -type CircuitState --linecomment"; DO NOT EDIT.

package conch

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[CircuitUndefined-0]
	_ = x[CircuitOpen-1]
	_ = x[CircuitClosed-2]
	_ = x[CircuitHalfOpen-3]
}

const _CircuitState_name = "UNDEFINEDOPENCLOSEDHALF_OPEN"

var _CircuitState_index = [...]uint8{0, 9, 13, 19, 28}

func (i CircuitState) String() string {
	if i >= CircuitState(len(_CircuitState_index)-1) {
		return "CircuitState(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _CircuitState_name[_CircuitState_index[i]:_CircuitState_index[i+1]]
}
