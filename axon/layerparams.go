// Copyright (c) 2023, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package axon

import (
	"encoding/json"

	"github.com/goki/gosl/sltype"
)

//gosl: hlsl layerparams
// #include "layertypes.hlsl"
// #include "act.hlsl"
// #include "inhib.hlsl"
// #include "learn_neur.hlsl"
// #include "deep_layers.hlsl"
// #include "pool.hlsl"
//gosl: end layerparams

//gosl: start layerparams

// global projection param arrays
// var SendPrjns []PrjnParams // [Layer][SendPrjns]
// var RecvPrjns []PrjnParams // [Layer][RecvPrjns]

// LayerIdxs contains index access into global arrays for GPU.
type LayerIdxs struct {
	Pool   uint32 // start of pools for this layer -- first one is always the layer-wide pool
	NeurSt uint32 // start of neurons for this layer in global array (same as Layer.NeurStIdx)
	RecvSt uint32 // start index into RecvPrjns global array
	RecvN  uint32 // number of recv projections
	SendSt uint32 // start index into SendPrjns global array
	SendN  uint32 // number of send projections

	pad, pad1 uint32
}

// LayerParams contains all of the layer parameters.
// These values must remain constant over the course of computation.
// On the GPU, they are loaded into a uniform.
type LayerParams struct {
	LayType LayerTypes `desc:"functional type of layer -- determines functional code path for specialized layer types, and is synchronized with the Layer.Typ value"`

	pad, pad1, pad2 int32

	Act   ActParams       `view:"add-fields" desc:"Activation parameters and methods for computing activations"`
	Inhib InhibParams     `view:"add-fields" desc:"Inhibition parameters and methods for computing layer-level inhibition"`
	Learn LearnNeurParams `view:"add-fields" desc:"Learning parameters and methods that operate at the neuron level"`

	//////////////////////////////////////////
	//  Specialized layer type parameters
	//     each applies to a specific layer type.
	//     use the `viewif` field tag to condition on LayType.

	Burst BurstParams `viewif:"LayType=Super" desc:"BurstParams determine how the 5IB Burst activation is computed from CaSpkP integrated spiking values in Super layers -- thresholded."`
	CT    CTParams    `viewif:"LayType=CT" desc:"params for the CT corticothalamic layer that generates predictions over the Pulvinar using context -- uses the CtxtGe excitatory input plus stronger NMDA channels to maintain context trace"`
	Pulv  PulvParams  `viewif:"LayType=Pulvinar" desc:"provides parameters for how the plus-phase (outcome) state of Pulvinar thalamic relay cell neurons is computed from the corresponding driver neuron Burst activation (or CaSpkP if not Super)"`

	Idxs LayerIdxs `view:"-" desc:"recv and send projection array access info"`
}

func (ly *LayerParams) Update() {
	ly.Act.Update()
	ly.Inhib.Update()
	ly.Learn.Update()
	ly.Burst.Update()
	ly.CT.Update()
	ly.Pulv.Update()
}

func (ly *LayerParams) Defaults() {
	ly.Act.Defaults()
	ly.Inhib.Defaults()
	ly.Learn.Defaults()
	ly.Inhib.Layer.On.SetBool(true)
	ly.Inhib.Layer.Gi = 1.0
	ly.Inhib.Pool.Gi = 1.0
	ly.Burst.Defaults()
	ly.CT.Defaults()
	ly.Pulv.Defaults()
}

// AllParams returns a listing of all parameters in the Layer
func (ly *LayerParams) AllParams() string {
	str := ""
	// todo: replace with a custom reflection crawler that generates
	// the right output directly and filters based on LayType etc.

	b, _ := json.MarshalIndent(&ly.Act, "", " ")
	str += "Act: {\n " + JsonToParams(b)
	b, _ = json.MarshalIndent(&ly.Inhib, "", " ")
	str += "Inhib: {\n " + JsonToParams(b)
	b, _ = json.MarshalIndent(&ly.Learn, "", " ")
	str += "Learn: {\n " + JsonToParams(b)

	switch ly.LayType {
	case Super:
		b, _ = json.MarshalIndent(&ly.Burst, "", " ")
		str += "Burst: {\n " + JsonToParams(b)
	case CT:
		b, _ = json.MarshalIndent(&ly.CT, "", " ")
		str += "CT:   {\n " + JsonToParams(b)
	case Pulvinar:
		b, _ = json.MarshalIndent(&ly.Pulv, "", " ")
		str += "Pulv: {\n " + JsonToParams(b)
	}
	return str
}

//////////////////////////////////////////////////////////////////////////////////////
//  GeExtToPool

// GeExtToPool adds GeExt from each neuron into the Pools
func (ly *LayerParams) GeExtToPool(ni uint32, nrn *Neuron, pl *Pool, lpl *Pool, subPool bool, ctime *Time) {
	pl.Inhib.GeExtRaw += nrn.GeExt // note: from previous cycle..
	if subPool {
		lpl.Inhib.GeExtRaw += nrn.GeExt
	}
}

// LayPoolGiFmSpikes computes inhibition Gi from Spikes for layer-level pool
func (ly *LayerParams) LayPoolGiFmSpikes(lpl *Pool, giMult float32, ctime *Time) {
	lpl.Inhib.SpikesFmRaw(lpl.NNeurons())
	ly.Inhib.Layer.Inhib(&lpl.Inhib, giMult)
}

// SubPoolGiFmSpikes computes inhibition Gi from Spikes within a sub-pool
func (ly *LayerParams) SubPoolGiFmSpikes(pl *Pool, lpl *Pool, lyInhib bool, giMult float32, ctime *Time) {
	pl.Inhib.SpikesFmRaw(pl.NNeurons())
	ly.Inhib.Pool.Inhib(&pl.Inhib, giMult)
	if lyInhib {
		pl.Inhib.LayerMax(lpl.Inhib.Gi) // note: this requires lpl inhib to have been computed before!
	} else {
		lpl.Inhib.PoolMax(pl.Inhib.Gi) // display only
		lpl.Inhib.SaveOrig()           // effective GiOrig
	}
}

//////////////////////////////////////////////////////////////////////////////////////
//  CycleNeuron methods

////////////////////////
//  GInteg

// NeuronGatherSpikesInit initializes G*Raw and G*Syn values for given neuron
// prior to integration
func (ly *LayerParams) NeuronGatherSpikesInit(ni uint32, nrn *Neuron, ctime *Time) {
	nrn.GeRaw = 0
	nrn.GiRaw = 0
	nrn.GeSyn = nrn.GeBase
	nrn.GiSyn = nrn.GiBase
}

// See prjnparams for NeuronGatherSpikesPrjn

// SpecialPreGs is used for special layer types to do things to the
// conductance values prior to doing the standard updates in GFmRawSyn
// drvAct is for Pulvinar layers, activation of driving neuron
func (ly *LayerParams) SpecialPreGs(ni uint32, nrn *Neuron, drvGe, nonDrvPct float32, ctime *Time, randctr *sltype.Uint2) float32 {
	var saveVal float32 // sometimes we need to use a value computed here, for the post Gs step
	switch ly.LayType {
	case CT:
		geCtxt := ly.CT.GeGain * nrn.CtxtGe
		nrn.GeRaw += geCtxt
		if ly.CT.DecayDt > 0 {
			nrn.CtxtGe -= ly.CT.DecayDt * nrn.CtxtGe
			ctxtExt := ly.Act.Dt.GeSynFmRawSteady(geCtxt)
			nrn.GeSyn += ctxtExt
			saveVal = ctxtExt // used In PostGs to set nrn.GeExt
		}
	case Pulvinar:
		if ctime.PlusPhase.IsFalse() {
			break
		}
		nrn.GeRaw = nonDrvPct*nrn.GeRaw + drvGe
		nrn.GeSyn = nonDrvPct*nrn.GeSyn + ly.Act.Dt.GeSynFmRawSteady(drvGe)
	}
	return saveVal
}

// SpecialPostGs is used for special layer types to do things
// after the standard updates in GFmRawSyn.
// It is passed the saveVal from SpecialPreGs
func (ly *LayerParams) SpecialPostGs(ni uint32, nrn *Neuron, ctime *Time, randctr *sltype.Uint2, saveVal float32) {
	switch ly.LayType {
	case CT:
		nrn.GeExt = saveVal // todo: it is not clear if this really does anything?  next time around?
	}
}

// GFmRawSyn computes overall Ge and GiSyn conductances for neuron
// from GeRaw and GeSyn values, including NMDA, VGCC, AMPA, and GABA-A channels.
// drvAct is for Pulvinar layers, activation of driving neuron
func (ly *LayerParams) GFmRawSyn(ni uint32, nrn *Neuron, drvGe, nonDrvPct float32, ctime *Time, randctr *sltype.Uint2) {
	saveVal := ly.SpecialPreGs(ni, nrn, drvGe, nonDrvPct, ctime, randctr)
	ly.Act.NMDAFmRaw(nrn, nrn.GeRaw)
	ly.Learn.LrnNMDAFmRaw(nrn, nrn.GeRaw)
	ly.Act.GvgccFmVm(nrn)
	ly.Act.GeFmSyn(ni, nrn, nrn.GeSyn, nrn.Gnmda+nrn.Gvgcc, randctr) // sets nrn.GeExt too
	ly.Act.GkFmVm(nrn)
	nrn.GiSyn = ly.Act.GiFmSyn(ni, nrn, nrn.GiSyn, randctr)
	ly.SpecialPostGs(ni, nrn, ctime, randctr, saveVal)
}

// GiInteg adds Gi values from all sources including SubPool computed inhib
// and updates GABAB as well
func (ly *LayerParams) GiInteg(ni uint32, nrn *Neuron, pl *Pool, giMult float32, ctime *Time) {
	// pl := &ly.Pools[nrn.SubPool]
	nrn.Gi = giMult*pl.Inhib.Gi + nrn.GiSyn + nrn.GiNoise
	nrn.SSGi = pl.Inhib.SSGi
	nrn.SSGiDend = 0
	if !(ly.Act.Clamp.IsInput.IsTrue() || ly.Act.Clamp.IsTarget.IsTrue()) {
		nrn.SSGiDend = ly.Act.Dend.SSGi * pl.Inhib.SSGi
	}
	ly.Act.GABAB.GABAB(nrn.GABAB, nrn.GABABx, nrn.Gi, &nrn.GABAB, &nrn.GABABx)
	nrn.GgabaB = ly.Act.GABAB.GgabaB(nrn.GABAB, nrn.VmDend)
	nrn.Gk += nrn.GgabaB // Gk was already init
}

////////////////////////
//  SpikeFmG

// SpikeFmG computes Vm from Ge, Gi, Gl conductances and then Spike from that
func (ly *LayerParams) SpikeFmG(ni uint32, nrn *Neuron, ctime *Time) {
	intdt := ly.Act.Dt.IntDt
	if ctime.PlusPhase.IsTrue() {
		intdt *= 3.0
	}
	ly.Act.VmFmG(nrn)
	ly.Act.SpikeFmVm(nrn)
	ly.Learn.CaFmSpike(nrn)
	if ctime.Cycle >= ly.Act.Dt.MaxCycStart {
		nrn.SpkMaxCa += ly.Learn.CaSpk.Dt.PDt * (nrn.CaSpkM - nrn.SpkMaxCa)
		if nrn.SpkMaxCa > nrn.SpkMax {
			nrn.SpkMax = nrn.SpkMaxCa
		}
	}
	nrn.ActInt += intdt * (nrn.Act - nrn.ActInt) // using reg act here now
	if ctime.PlusPhase.IsFalse() {
		nrn.GeM += ly.Act.Dt.IntDt * (nrn.Ge - nrn.GeM)
		nrn.GiM += ly.Act.Dt.IntDt * (nrn.GiSyn - nrn.GiM)
	}
}

// PostSpike does updates at neuron level after spiking has been computed.
// This is where special layer types add extra code.
func (ly *LayerParams) PostSpike(ni uint32, nrn *Neuron, vals *LayerVals, ctime *Time) {
	switch ly.LayType {
	case Super:
		if ctime.PlusPhase.IsTrue() {
			actMax := vals.ActAvg.CaSpkP.Max
			actAvg := vals.ActAvg.CaSpkP.Avg
			thr := ly.Burst.ThrFmAvgMax(actAvg, actMax)
			burst := float32(0)
			if nrn.CaSpkP > thr {
				burst = nrn.CaSpkP
			}
			nrn.Burst = burst
		}
	}
}

//gosl: end layerparams
