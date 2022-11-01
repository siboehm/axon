// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pvlv

import (
	"github.com/emer/axon/axon"
	"github.com/emer/axon/rl"
	"github.com/goki/ki/kit"
)

type PPTgNeuron struct {
	GeSynMax  float32 `desc:"max excitatory synaptic inputs"`
	GeSynPrev float32 `desc:"previous max excitatory synaptic inputs"`
}

// PPTgLayer represents a pedunculopontine nucleus layer
// it subtracts prior trial's excitatory conductance to
// compute the temporal derivative over time, with a positive
// rectification.
// also sets Act to the exact differenence.
type PPTgLayer struct {
	rl.Layer
	PPTgNeurs []PPTgNeuron
}

var KiT_PPTgLayer = kit.Types.AddType(&PPTgLayer{}, LayerProps)

func (ly *PPTgLayer) Defaults() {
	ly.Layer.Defaults()
	ly.Typ = PPTg

	// special inhib params
	ly.Act.Decay.Act = 0
	ly.Act.Decay.Glong = 0
	ly.Act.NMDA.Gbar = 0
	ly.Act.GABAB.Gbar = 0
	ly.Inhib.Layer.On = true
	ly.Inhib.Layer.Gi = 1.0
	ly.Inhib.Pool.On = true
	ly.Inhib.Pool.Gi = 0.5
	// ly.Inhib.Layer.FB = 0
	// ly.Inhib.Pool.FB = 0
	ly.Inhib.ActAvg.Init = 0.1

	for _, pji := range ly.RcvPrjns {
		pj := pji.(axon.AxonPrjn).AsAxon()
		pj.SWt.Init.SPct = 0
		pj.PrjnScale.Abs = 1
		pj.Learn.Learn = false
		pj.SWt.Adapt.SigGain = 1
		pj.SWt.Init.Mean = 0.75
		pj.SWt.Init.Var = 0.0
		pj.SWt.Init.Sym = false
	}
}

func (ly *PPTgLayer) Build() error {
	err := ly.Layer.Build()
	if err != nil {
		return err
	}
	ly.PPTgNeurs = make([]PPTgNeuron, len(ly.Neurons))
	return err
}

func (ly *PPTgLayer) NewState() {
	for ni := range ly.PPTgNeurs {
		nrn := &ly.PPTgNeurs[ni]
		nrn.GeSynPrev = nrn.GeSynMax
		nrn.GeSynMax = 0
	}
	ly.Layer.NewState()
}

func (ly *PPTgLayer) GFmSpike(ltime *axon.Time) {
	ly.GFmSpikePrjn(ltime)
	for ni := range ly.Neurons {
		nrn := &ly.Neurons[ni]
		if nrn.IsOff() {
			continue
		}
		ly.GFmSpikeNeuron(ltime, ni, nrn)
		ly.GFmRawSynNeuron(ltime, ni, nrn)
		pnr := &ly.PPTgNeurs[ni]
		if nrn.GeSyn > pnr.GeSynMax {
			pnr.GeSynMax = nrn.GeSyn
		}
	}
}

func (ly *PPTgLayer) GFmRawSynNeuron(ltime *axon.Time, ni int, nrn *axon.Neuron) {
	pnr := &ly.PPTgNeurs[ni]
	geSyn := (nrn.GeSyn - pnr.GeSynPrev)
	if geSyn < 0 {
		geSyn = 0
	}
	geRawPrev := pnr.GeSynPrev * ly.Act.Dt.GeDt
	geRaw := (nrn.GeRaw - geRawPrev)
	if geRaw < 0 {
		geRaw = 0
	}

	ly.Act.NMDAFmRaw(nrn, geRaw)
	ly.Learn.LrnNMDAFmRaw(nrn, geRaw)
	ly.Act.GvgccFmVm(nrn)

	ly.Act.GeFmSyn(nrn, geSyn, nrn.Gnmda+nrn.Gvgcc)
	nrn.GiSyn = ly.Act.GiFmSyn(nrn, nrn.GiSyn)
}

func (ly *PPTgLayer) ActFmG(ltime *axon.Time) {
	ly.Layer.ActFmG(ltime)
	for ni := range ly.Neurons {
		nrn := &ly.Neurons[ni]
		if nrn.IsOff() {
			continue
		}
		diff := nrn.CaSpkP - nrn.SpkPrv
		if diff < 0 {
			diff = 0
		}
		nrn.Act = diff
		nrn.ActInt = nrn.Act
	}
}
