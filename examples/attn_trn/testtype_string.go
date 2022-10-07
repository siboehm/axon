// Code generated by "stringer -type=TestType"; DO NOT EDIT.

//go:build this_is_broken_we_should_fix_or_delete


package main

import (
	"errors"
	"strconv"
)

var _ = errors.New("dummy error")

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[AttnSize-0]
	_ = x[AttnSizeDebug-1]
	_ = x[AttnSizeC2Up-2]
	_ = x[Popout-3]
	_ = x[TestTypeN-4]
}

const _TestType_name = "AttnSizeAttnSizeDebugAttnSizeC2UpPopoutTestTypeN"

var _TestType_index = [...]uint8{0, 8, 21, 33, 39, 48}

func (i TestType) String() string {
	if i < 0 || i >= TestType(len(_TestType_index)-1) {
		return "TestType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _TestType_name[_TestType_index[i]:_TestType_index[i+1]]
}

func (i *TestType) FromString(s string) error {
	for j := 0; j < len(_TestType_index)-1; j++ {
		if s == _TestType_name[_TestType_index[j]:_TestType_index[j+1]] {
			*i = TestType(j)
			return nil
		}
	}
	return errors.New("String: " + s + " is not a valid option for type: TestType")
}
