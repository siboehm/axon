// Copyright (c) 2023, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package axon

import "github.com/goki/ki/kit"

//gosl: start prjntypes

// PrjnTypes is an axon-specific prjn type enum,
// that encompasses all the different algorithm types supported.
// Class parameter styles automatically key off of these types.
// The first entries must be kept synchronized with the emer.PrjnType.
type PrjnTypes int32

// The projection types
const (
	// Forward is a feedforward, bottom-up projection from sensory inputs to higher layers
	Forward PrjnTypes = iota

	// Back is a feedback, top-down projection from higher layers back to lower layers
	Back

	// Lateral is a lateral projection within the same layer / area
	Lateral

	// Inhibitory is an inhibitory projection that drives inhibitory
	// synaptic conductances instead of the default excitatory ones.
	Inhibitory

	// CTCtxt are projections from Superficial layers to CT layers that
	// send Burst activations drive updating of CtxtGe excitatory conductance,
	// at end of plus (51B Bursting) phase.  Biologically, this projection
	// comes from the PT layer 5IB neurons, but it is simpler to use the
	// Super neurons directly, and PT are optional for most network types.
	// These projections also use a special learning rule that
	// takes into account the temporal delays in the activation states.
	// Can also add self context from CT for deeper temporal context.
	CTCtxt

	PrjnTypesN
)

//gosl: end prjntypes

//go:generate stringer -type=PrjnTypes

var KiT_PrjnTypes = kit.Enums.AddEnum(PrjnTypesN, kit.NotBitFlag, nil)

func (ev PrjnTypes) MarshalJSON() ([]byte, error)  { return kit.EnumMarshalJSON(ev) }
func (ev *PrjnTypes) UnmarshalJSON(b []byte) error { return kit.EnumUnmarshalJSON(ev, b) }
