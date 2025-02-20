// Copyright (c) 2020, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package axon

import (
	"github.com/emer/etable/minmax"
	"github.com/goki/gosl/slbool"
	"github.com/goki/mat32"
)

//gosl: start rl_layers

// RSalAChParams compute reward salience as ACh global neuromodulatory signal
// as a function of the MAX activation of its inputs.
type RSalAChParams struct {
	RewThr     float32     `desc:"threshold per input source, on absolute value (magnitude), to count as a significant reward event, which then drives maximal ACh -- set to 0 to disable this nonlinear behavior"`
	Rew        slbool.Bool `desc:"use the global Context.NeuroMod.HasRew flag -- if there is some kind of external reward being given, then ACh goes to 1, else 0 for this component"`
	RewPred    slbool.Bool `desc:"use the global Context.NeuroMod.RewPred value"`
	PPTg       slbool.Bool `desc:"use the global Context.NeuroMod.PPTg value"`
	SrcLay1Idx int32       `inactive:"+" desc:"idx of Layer to get max activity from -- set during Build from BuildConfig SrcLay1Name if present -- -1 if not used"`
	SrcLay2Idx int32       `inactive:"+" desc:"idx of Layer to get max activity from -- set during Build from BuildConfig SrcLay2Name if present -- -1 if not used"`
	SrcLay3Idx int32       `inactive:"+" desc:"idx of Layer to get max activity from -- set during Build from BuildConfig SrcLay3Name if present -- -1 if not used"`
	SrcLay4Idx int32       `inactive:"+" desc:"idx of Layer to get max activity from -- set during Build from BuildConfig SrcLay4Name if present -- -1 if not used"`
}

func (rp *RSalAChParams) Defaults() {
	rp.RewThr = 0.1
	rp.Rew.SetBool(true)
	rp.PPTg.SetBool(true)
}

func (rp *RSalAChParams) Update() {
}

// Thr applies
func (rp *RSalAChParams) Thr(val float32) float32 {
	if rp.RewThr <= 0 {
		return val
	}
	if mat32.Abs(val) < rp.RewThr {
		return 0
	}
	return 1
}

// RWPredParams parameterizes reward prediction for a simple Rescorla-Wagner
// learning dynamic (i.e., PV learning in the PVLV framework).
type RWPredParams struct {
	PredRange minmax.F32 `desc:"default 0.1..0.99 range of predictions that can be represented -- having a truncated range preserves some sensitivity in dopamine at the extremes of good or poor performance"`
}

func (rp *RWPredParams) Defaults() {
	rp.PredRange.Set(0.01, 0.99)
}

func (rp *RWPredParams) Update() {
}

// RWDaParams computes a dopamine (DA) signal using simple Rescorla-Wagner
// learning dynamic (i.e., PV learning in the PVLV framework).
type RWDaParams struct {
	TonicGe      float32 `desc:"tonic baseline Ge level for DA = 0 -- +/- are between 0 and 2*TonicGe -- just for spiking display of computed DA value"`
	RWPredLayIdx int32   `inactive:"+" desc:"idx of RWPredLayer to get reward prediction from -- set during Build from BuildConfig RWPredLayName"`

	pad, pad1 uint32
}

func (rp *RWDaParams) Defaults() {
	rp.TonicGe = 0.2
}

func (rp *RWDaParams) Update() {
}

// GeFmDA returns excitatory conductance from DA dopamine value
func (rp *RWDaParams) GeFmDA(da float32) float32 {
	ge := rp.TonicGe * (1.0 + da)
	if ge < 0 {
		ge = 0
	}
	return ge
}

// TDIntegParams are params for reward integrator layer
type TDIntegParams struct {
	Discount     float32 `desc:"discount factor -- how much to discount the future prediction from TDPred"`
	PredGain     float32 `desc:"gain factor on TD rew pred activations"`
	TDPredLayIdx int32   `inactive:"+" desc:"idx of TDPredLayer to get reward prediction from -- set during Build from BuildConfig TDPredLayName"`

	pad uint32
}

func (tp *TDIntegParams) Defaults() {
	tp.Discount = 0.9
	tp.PredGain = 1
}

func (tp *TDIntegParams) Update() {
}

// TDDaParams are params for dopamine (DA) signal as the temporal difference (TD)
// between the TDIntegLayer activations in the minus and plus phase.
type TDDaParams struct {
	TonicGe       float32 `desc:"tonic baseline Ge level for DA = 0 -- +/- are between 0 and 2*TonicGe -- just for spiking display of computed DA value"`
	TDIntegLayIdx int32   `inactive:"+" desc:"idx of TDIntegLayer to get reward prediction from -- set during Build from BuildConfig TDIntegLayName"`

	pad, pad1 uint32
}

func (tp *TDDaParams) Defaults() {
	tp.TonicGe = 0.3
}

func (tp *TDDaParams) Update() {
}

// GeFmDA returns excitatory conductance from DA dopamine value
func (tp *TDDaParams) GeFmDA(da float32) float32 {
	return tp.TonicGe * (1.0 + da)
}

//gosl: end rl_layers

// note: Defaults not called on GPU

func (ly *LayerParams) RWDefaults() {
	ly.Inhib.ActAvg.Nominal = .5
}

func (ly *LayerParams) RWPredDefaults() {
	ly.Act.Decay.Act = 1
	ly.Act.Decay.Glong = 1
	ly.Act.Dt.GeTau = 40
}

// RWDaPostBuild does post-Build config
func (ly *Layer) RWDaPostBuild() {
	ly.Params.RWDa.RWPredLayIdx = ly.BuildConfigFindLayer("RWPredLayName", true)
}

func (ly *LayerParams) TDDefaults() {
	ly.Inhib.ActAvg.Nominal = .5
}

func (ly *LayerParams) TDPredDefaults() {
	ly.Act.Decay.Act = 1
	ly.Act.Decay.Glong = 1
	ly.Act.Dt.GeTau = 40
}

func (ly *Layer) RSalAChPostBuild() {
	ly.Params.RSalACh.SrcLay1Idx = ly.BuildConfigFindLayer("SrcLay1Name", false) // optional
	ly.Params.RSalACh.SrcLay2Idx = ly.BuildConfigFindLayer("SrcLay2Name", false) // optional
	ly.Params.RSalACh.SrcLay3Idx = ly.BuildConfigFindLayer("SrcLay3Name", false) // optional
	ly.Params.RSalACh.SrcLay4Idx = ly.BuildConfigFindLayer("SrcLay4Name", false) // optional
}

// TDIntegPostBuild does post-Build config
func (ly *Layer) TDIntegPostBuild() {
	ly.Params.TDInteg.TDPredLayIdx = ly.BuildConfigFindLayer("TDPredLayName", true)
}

// TDDaPostBuild does post-Build config
func (ly *Layer) TDDaPostBuild() {
	ly.Params.TDDa.TDIntegLayIdx = ly.BuildConfigFindLayer("TDIntegLayName", true)
}
