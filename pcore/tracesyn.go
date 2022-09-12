// Copyright (c) 2020, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pcore

import "github.com/goki/mat32"

// TraceSyn holds extra synaptic state for trace projections
type TraceSyn struct {
	NTr float32 `desc:"new trace = send * recv -- drives updates to trace value: sn.ActLrn * rn.ActLrn (subject to derivative too)"`
}

// VarByName returns synapse variable by name
func (sy *TraceSyn) VarByName(varNm string) float32 {
	switch varNm {
	case "NTr":
		return sy.NTr
	}
	return mat32.NaN()
}

// VarByIndex returns synapse variable by index
func (sy *TraceSyn) VarByIndex(varIdx int) float32 {
	switch varIdx {
	case 0:
		return sy.NTr
	}
	return mat32.NaN()
}

var TraceSynVars = []string{"NTr"}
