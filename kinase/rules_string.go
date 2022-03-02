// Code generated by "stringer -type=Rules"; DO NOT EDIT.

package kinase

import (
	"errors"
	"strconv"
)

var _ = errors.New("dummy error")

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[NeurSpkCa-0]
	_ = x[SynSpkCa-1]
	_ = x[SynNMDACa-2]
	_ = x[RulesN-3]
}

const _Rules_name = "NeurSpkCaSynSpkCaSynNMDACaRulesN"

var _Rules_index = [...]uint8{0, 9, 17, 26, 32}

func (i Rules) String() string {
	if i < 0 || i >= Rules(len(_Rules_index)-1) {
		return "Rules(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _Rules_name[_Rules_index[i]:_Rules_index[i+1]]
}

func (i *Rules) FromString(s string) error {
	for j := 0; j < len(_Rules_index)-1; j++ {
		if s == _Rules_name[_Rules_index[j]:_Rules_index[j+1]] {
			*i = Rules(j)
			return nil
		}
	}
	return errors.New("String: " + s + " is not a valid option for type: Rules")
}