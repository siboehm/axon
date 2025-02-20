// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package chans provides standard neural conductance channels for computing
a point-neuron approximation based on the standard equivalent RC circuit
model of a neuron (i.e., basic Ohms law equations).
Includes excitatory, leak, inhibition, and dynamic potassium channels.
*/
package chans

//gosl: hlsl chans
// #include "fastexp.hlsl"
//gosl: end chans

//gosl: start chans

// Chans are ion channels used in computing point-neuron activation function
type Chans struct {
	E float32 `desc:"excitatory sodium (Na) AMPA channels activated by synaptic glutamate"`
	L float32 `desc:"constant leak (potassium, K+) channels -- determines resting potential (typically higher than resting potential of K)"`
	I float32 `desc:"inhibitory chloride (Cl-) channels activated by synaptic GABA"`
	K float32 `desc:"gated / active potassium channels -- typically hyperpolarizing relative to leak / rest"`
}

// SetAll sets all the values
func (ch *Chans) SetAll(e, l, i, k float32) {
	ch.E, ch.L, ch.I, ch.K = e, l, i, k
}

// VToBio returns biological mV voltage from normalized 0-1 voltage
// where 0 = -100mV and 1 = 0mV
func VToBio(vm float32) float32 {
	return vm*100 - 100
}

// VFmBio returns normalized 0-1 voltage from biological mV voltage
// where 0 = -100mV and 1 = 0mV
func VFmBio(vm float32) float32 {
	return (vm + 100) / 100
}

//gosl: end chans

// SetFmOtherMinus sets all the values from other Chans minus given value
func (ch *Chans) SetFmOtherMinus(oth Chans, minus float32) {
	ch.E, ch.L, ch.I, ch.K = oth.E-minus, oth.L-minus, oth.I-minus, oth.K-minus
}

// SetFmMinusOther sets all the values from given value minus other Chans
func (ch *Chans) SetFmMinusOther(minus float32, oth Chans) {
	ch.E, ch.L, ch.I, ch.K = minus-oth.E, minus-oth.L, minus-oth.I, minus-oth.K
}
