// Copyright (c) 2023, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package axon

import (
	"github.com/goki/gosl/slbool"
	"github.com/goki/ki/kit"
	"github.com/goki/mat32"
)

//go:generate stringer -type=DAModTypes
//go:generate stringer -type=ValenceTypes

var KiT_DAModTypes = kit.Enums.AddEnum(DAModTypesN, kit.NotBitFlag, nil)
var KiT_ValenceTypes = kit.Enums.AddEnum(ValenceTypesN, kit.NotBitFlag, nil)

//gosl: start neuromod

// NeuroModVals neuromodulatory values -- they are global to the layer and
// affect learning rate and other neural activity parameters of neurons.
type NeuroModVals struct {
	Rew      float32     `inactive:"+" desc:"reward value -- this is set here in the Context struct, and the RL Rew layer grabs it from there -- must also set HasRew flag when rew is set -- otherwise is ignored."`
	HasRew   slbool.Bool `inactive:"+" desc:"must be set to true when a reward is present -- otherwise Rew is ignored.  Also set during extinction by PVLV.  This drives ACh release in the PVLV model."`
	RewPred  float32     `inactive:"+" desc:"reward prediction -- computed by a special reward prediction layer"`
	PrevPred float32     `inactive:"+" desc:"previous time step reward prediction -- e.g., for TDPredLayer"`
	DA       float32     `inactive:"+" desc:"dopamine -- represents reward prediction error, signaled as phasic increases or decreases in activity relative to a tonic baseline, which is represented by a value of 0.  Released by the VTA -- ventral tegmental area, or SNc -- substantia nigra pars compacta."`
	ACh      float32     `inactive:"+" desc:"acetylcholine -- activated by salient events, particularly at the onset of a reward / punishment outcome (US), or onset of a conditioned stimulus (CS).  Driven by BLA -> PPtg that detects changes in BLA activity, via RSalienceAChLayer type"`
	NE       float32     `inactive:"+" desc:"norepinepherine -- not yet in use"`
	Ser      float32     `inactive:"+" desc:"serotonin -- not yet in use"`

	AChRaw float32 `inactive:"+" desc:"raw ACh value used in updating global ACh value by RSalienceAChLayer"`
	PPTg   float32 `inactive:"+" desc:"raw PPTg value reflecting the positive-rectified delta output of the Amygdala, which drives ACh and DA in the PVLV framework "`

	pad, pad1 float32
}

func (nm *NeuroModVals) Init() {
	nm.Rew = 0
	nm.HasRew.SetBool(false)
	nm.RewPred = 0
	nm.DA = 0
	nm.ACh = 0
	nm.NE = 0
	nm.Ser = 0
	nm.AChRaw = 0
}

// SetRew is a convenience function for setting the external reward
func (nm *NeuroModVals) SetRew(rew float32, hasRew bool) {
	nm.HasRew.SetBool(hasRew)
	if hasRew {
		nm.Rew = rew
	} else {
		nm.Rew = 0
	}
}

// NewState is called by Context.NewState at start of new trial
func (nm *NeuroModVals) NewState() {
	nm.Init()
}

// AChFmRaw updates ACh from AChRaw using given decay time constant.
func (nm *NeuroModVals) AChFmRaw(dt float32) {
	if nm.AChRaw > nm.ACh { // instant up
		nm.ACh = nm.AChRaw
	} else {
		nm.ACh += dt * (nm.AChRaw - nm.ACh)
	}
}

// DAModTypes are types of dopamine modulation of neural activity.
type DAModTypes int32

const (
	// NoDAMod means there is no effect of dopamine on neural activity
	NoDAMod DAModTypes = iota

	// D1Mod is for neurons that primarily express dopamine D1 receptors,
	// which are excitatory from DA bursts, inhibitory from dips.
	// Cortical neurons can generally use this type, while subcortical
	// populations are more diverse in having both D1 and D2 subtypes.
	D1Mod

	// D2Mod is for neurons that primarily express dopamine D2 receptors,
	// which are excitatory from DA dips, inhibitory from bursts.
	D2Mod

	// D1AbsMod is like D1Mod, except the absolute value of DA is used
	// instead of the signed value.
	// There are a subset of DA neurons that send increased DA for
	// both negative and positive outcomes, targeting frontal neurons.
	D1AbsMod

	DAModTypesN
)

// ValenceTypes are types of valence coding: positive or negative.
type ValenceTypes int32

const (
	// Positive valence codes for outcomes aligned with drives / goals.
	Positive ValenceTypes = iota

	// Negative valence codes for harmful or aversive outcomes.
	Negative

	ValenceTypesN
)

// NeuroModParams specifies the effects of neuromodulators on neural
// activity and learning rate.  These can apply to any neuron type,
// and are applied in the core cycle update equations.
type NeuroModParams struct {
	DAMod       DAModTypes   `desc:"dopamine receptor-based effects of dopamine modulation on excitatory and inhibitory conductances: D1 is excitatory, D2 is inhibitory as a function of increasing dopamine"`
	Valence     ValenceTypes `desc:"valence coding of this layer -- may affect specific layer types but does not directly affect neuromodulators currently"`
	DAModGain   float32      `viewif:"DAMod!=NoDAMod" desc:"multiplicative factor on overall DA modulation specified by DAMod -- resulting overall gain factor is: 1 + DAModGain * DA, where DA is appropriate DA-driven factor"`
	DALRateSign slbool.Bool  `desc:"modulate the sign of the learning rate factor according to the DA sign, taking into account the DAMod sign reversal for D2Mod, also using BurstGain and DipGain to modulate DA value -- otherwise, only the magnitude of the learning rate is modulated as a function of raw DA magnitude according to DALRateMod (without additional gain factors)"`
	DALRateMod  float32      `min:"0" max:"1" viewif:"!DALRateSign" desc:"if not using DALRateSign, this is the proportion of maximum learning rate that Abs(DA) magnitude can modulate -- e.g., if 0.2, then DA = 0 = 80% of std learning rate, 1 = 100%"`
	AChLRateMod float32      `min:"0" max:"1" desc:"proportion of maximum learning rate that ACh can modulate -- e.g., if 0.2, then ACh = 0 = 80% of std learning rate, 1 = 100%"`
	AChDisInhib float32      `min:"0" def:"0,5" desc:"amount of extra Gi inhibition added in proportion to 1 - ACh level -- makes ACh disinhibitory"`
	BurstGain   float32      `min:"0" def:"1" desc:"multiplicative gain factor applied to positive dopamine signals -- this operates on the raw dopamine signal prior to any effect of D2 receptors in reversing its sign!"`
	DipGain     float32      `min:"0" def:"1" desc:"multiplicative gain factor applied to negative dopamine signals -- this operates on the raw dopamine signal prior to any effect of D2 receptors in reversing its sign! should be small for acq, but roughly equal to burst for ext"`

	pad, pad1, pad2 float32
}

func (nm *NeuroModParams) Defaults() {
	// nm.DAMod is typically set by BuildConfig -- don't reset here
	nm.DAModGain = 0.5
	nm.DALRateMod = 0
	nm.AChLRateMod = 0
	nm.BurstGain = 1
	nm.DipGain = 1
}

func (nm *NeuroModParams) Update() {
	nm.DALRateMod = mat32.Clamp(nm.DALRateMod, 0, 1)
	nm.AChLRateMod = mat32.Clamp(nm.AChLRateMod, 0, 1)
}

// LRModFact returns learning rate modulation factor for given inputs.
func (nm *NeuroModParams) LRModFact(pct, val float32) float32 {
	val = mat32.Clamp(mat32.Abs(val), 0, 1)
	return 1.0 - pct*(1.0-val)
}

// DAGain returns DA dopamine value with Burst / Dip Gain factors applied
func (nm *NeuroModParams) DAGain(da float32) float32 {
	if da > 0 {
		da *= nm.BurstGain
	} else {
		da *= nm.DipGain
	}
	return da
}

// LRMod returns overall learning rate modulation factor due to neuromodulation
// from given dopamine (DA) and ACh inputs.
// If DALRateMod is true and DAMod == D1Mod or D2Mod, then the sign is a function
// of the DA
func (nm *NeuroModParams) LRMod(da, ach float32) float32 {
	mod := nm.LRModFact(nm.AChLRateMod, ach)
	if nm.DALRateSign.IsTrue() {
		da := nm.DAGain(da)
		if nm.DAMod == D1Mod {
			mod *= da
		} else if nm.DAMod == D2Mod {
			mod *= -da
		}
	} else {
		mod *= nm.LRModFact(nm.DALRateMod, da)
	}
	return mod
}

// GGain returns effective Ge and Gi gain factor given
// dopamine (DA) +/- burst / dip value (0 = tonic level).
// factor is 1 for no modulation, otherwise higher or lower.
func (nm *NeuroModParams) GGain(da float32) float32 {
	if da > 0 {
		da *= nm.BurstGain
	} else {
		da *= nm.DipGain
	}
	gain := float32(1)
	switch nm.DAMod {
	case NoDAMod:
	case D1Mod:
		gain += nm.DAModGain * da
	case D2Mod:
		gain -= nm.DAModGain * da
	case D1AbsMod:
		gain += nm.DAModGain * mat32.Abs(da)
	}
	return gain
}

// GIFmACh returns amount of extra inhibition to add based on disinhibitory
// effects of ACh -- no inhibition when ACh = 1, extra when < 1.
func (nm *NeuroModParams) GiFmACh(ach float32) float32 {
	ai := 1 - ach
	if ai < 0 {
		ai = 0
	}
	return nm.AChDisInhib * ai
}

//gosl: end neuromod
