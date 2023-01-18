// Code generated by "stringer -type=PrjnTypes"; DO NOT EDIT.

package axon

import (
	"errors"
	"strconv"
)

var _ = errors.New("dummy error")

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[Forward-0]
	_ = x[Back-1]
	_ = x[Lateral-2]
	_ = x[Inhibitory-3]
	_ = x[CTCtxt-4]
	_ = x[PrjnTypesN-5]
}

const _PrjnTypes_name = "ForwardBackLateralInhibitoryCTCtxtPrjnTypesN"

var _PrjnTypes_index = [...]uint8{0, 7, 11, 18, 28, 34, 44}

func (i PrjnTypes) String() string {
	if i < 0 || i >= PrjnTypes(len(_PrjnTypes_index)-1) {
		return "PrjnTypes(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _PrjnTypes_name[_PrjnTypes_index[i]:_PrjnTypes_index[i+1]]
}

func (i *PrjnTypes) FromString(s string) error {
	for j := 0; j < len(_PrjnTypes_index)-1; j++ {
		if s == _PrjnTypes_name[_PrjnTypes_index[j]:_PrjnTypes_index[j+1]] {
			*i = PrjnTypes(j)
			return nil
		}
	}
	return errors.New("String: " + s + " is not a valid option for type: PrjnTypes")
}
