// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package axon

// prjn_compute.go has the core computational methods, for the CPU.
// On GPU, this same functionality is implemented in corresponding gpu_*.hlsl
// files, which correspond to different shaders for each different function.

//////////////////////////////////////////////////////////////////////////////////////
//  Act methods

// RecvSpikes receives spikes from the sending neurons at index sendIdx
// into the GBuf buffer on the receiver side. The buffer on the receiver side
// is a ring buffer, which is used for modelling the time delay between
// sending and receiving spikes.
// THIS IS NOT USED BY DEFAULT -- VERY SLOW!
func (pj *Prjn) RecvSpikes(ctx *Context, recvIdx int) {
	return
	/*
		scale := pj.Params.GScale.Scale
		slay := pj.Send
		pjcom := &pj.Params.Com
		bi := pjcom.WriteIdx(uint32(recvIdx), ctx.CyclesTotal-1, pj.Params.Idxs.RecvNeurN)
		// note: -1 because this is logically done on prior timestep
		syns := pj.RecvSyns(recvIdx)
		if pj.PrjnType() == CTCtxtPrjn {
			if ctx.Cycle != ctx.ThetaCycles-1-int32(pj.Params.Com.DelLen) {
				return
			}
			for ci := range syns {
				sy := &syns[ci]
				sendIdx := pj.Params.SynSendLayIdx(sy)
				sn := &slay.Neurons[sendIdx]
				sv := sn.Burst * scale * sy.Wt
				pj.GBuf[bi] += sv
			}
		} else {
			for ci := range syns {
				sy := &syns[ci]
				sendIdx := pj.Params.SynSendLayIdx(sy)
				sn := &slay.Neurons[sendIdx]
				sv := sn.Spike * scale * sy.Wt
				pj.GBuf[bi] += sv
			}
		}
	*/
}

// SendSpike sends a spike from the sending neuron at index sendIdx
// into the GBuf buffer on the receiver side. The buffer on the receiver side
// is a ring buffer, which is used for modelling the time delay between
// sending and receiving spikes.
func (pj *Prjn) SendSpike(ctx *Context, sendIdx int, nrn *Neuron) {
	scale := pj.Params.GScale.Scale * pj.Params.Com.FloatToIntFactor() // pre-bake in conversion to uint factor
	if pj.PrjnType() == CTCtxtPrjn {
		if ctx.Cycle != ctx.ThetaCycles-1-int32(pj.Params.Com.DelLen) {
			return
		}
		scale *= nrn.Burst // Burst is regular CaSpkP for all non-SuperLayer neurons
	} else {
		if nrn.Spike == 0 {
			return
		}
	}
	pjcom := &pj.Params.Com
	wrOff := pjcom.WriteOff(ctx.CyclesTotal)
	sidxs := pj.SendSynIdxs(sendIdx)
	for _, ssi := range sidxs {
		sy := &pj.Syns[ssi]
		recvIdx := pj.Params.SynRecvLayIdx(sy)
		sv := int32(scale * sy.Wt)
		bi := pjcom.WriteIdxOff(recvIdx, wrOff, pj.Params.Idxs.RecvNeurN)
		pj.GBuf[bi] += sv
	}
}

//////////////////////////////////////////////////////////////////////////////////////
//  SynCa methods

// todo: exclude doing SynCa for types that don't use it!

// SynCaSend updates synaptic calcium based on spiking, for SynSpkTheta mode.
// Optimized version only updates at point of spiking.
// This pass goes through in sending order, filtering on sending spike.
// Threading: Can be called concurrently for all prjns, since it updates synapses
// (which are local to a single prjn).
func (pj *Prjn) SynCaSend(ctx *Context, ni uint32, sn *Neuron, updtThr float32) {
	if pj.Params.Learn.Learn.IsFalse() {
		return
	}
	rlay := pj.Recv
	snCaSyn := pj.Params.Learn.KinaseCa.SpikeG * sn.CaSyn
	sidxs := pj.SendSynIdxs(int(ni))
	for _, ssi := range sidxs {
		sy := &pj.Syns[ssi]
		ri := pj.Params.SynRecvLayIdx(sy)
		rn := &rlay.Neurons[ri]
		pj.Params.SynCaSendSyn(ctx, sy, rn, snCaSyn, updtThr)
	}
}

// SynCaRecv updates synaptic calcium based on spiking, for SynSpkTheta mode.
// Optimized version only updates at point of spiking.
// This pass goes through in recv order, filtering on recv spike.
// Threading: Can be called concurrently for all prjns, since it updates synapses
// (which are local to a single prjn).
func (pj *Prjn) SynCaRecv(ctx *Context, ni uint32, rn *Neuron, updtThr float32) {
	if pj.Params.Learn.Learn.IsFalse() {
		return
	}
	slay := pj.Send
	rnCaSyn := pj.Params.Learn.KinaseCa.SpikeG * rn.CaSyn
	syns := pj.RecvSyns(int(ni))
	for ci := range syns {
		sy := &syns[ci]
		si := pj.Params.SynSendLayIdx(sy)
		sn := &slay.Neurons[si]
		pj.Params.SynCaRecvSyn(ctx, sy, sn, rnCaSyn, updtThr)
	}
}

//////////////////////////////////////////////////////////////////////////////////////
//  Learn methods

// DWt computes the weight change (learning), based on
// synaptically-integrated spiking, computed at the Theta cycle interval.
// This is the trace version for hidden units, and uses syn CaP - CaD for targets.
func (pj *Prjn) DWt(ctx *Context) {
	if pj.Params.Learn.Learn.IsFalse() {
		return
	}
	slay := pj.Send
	rlay := pj.Recv
	layPool := &rlay.Pools[0]
	isTarget := rlay.Params.Act.Clamp.IsTarget.IsTrue()
	for ri := range rlay.Neurons {
		rn := &rlay.Neurons[ri]
		// note: UpdtThr doesn't make sense here b/c Tr needs to be updated
		syns := pj.RecvSyns(ri)
		for ci := range syns {
			sy := &syns[ci]
			si := pj.Params.SynSendLayIdx(sy)
			sn := &slay.Neurons[si]
			subPool := &rlay.Pools[rn.SubPool]
			pj.Params.DWtSyn(ctx, sy, sn, rn, layPool, subPool, isTarget)
		}
	}
}

// DWtSubMean subtracts the mean from any projections that have SubMean > 0.
// This is called on *receiving* projections, prior to WtFmDwt.
func (pj *Prjn) DWtSubMean(ctx *Context) {
	rlay := pj.Recv
	sm := pj.Params.Learn.Trace.SubMean
	if sm == 0 { // || rlay.AxonLay.IsTarget() { // sm default is now 0, so don't exclude
		return
	}
	for ri := range rlay.Neurons {
		syns := pj.RecvSyns(ri)
		if len(syns) < 1 {
			continue
		}
		sumDWt := float32(0)
		nnz := 0 // non-zero
		for ci := range syns {
			sy := &syns[ci]
			dw := sy.DWt
			if dw != 0 {
				sumDWt += dw
				nnz++
			}
		}
		if nnz <= 1 {
			continue
		}
		sumDWt /= float32(nnz)
		for ci := range syns {
			sy := &syns[ci]
			if sy.DWt != 0 {
				sy.DWt -= sm * sumDWt
			}
		}
	}
}

// WtFmDWt updates the synaptic weight values from delta-weight changes.
// called on the *receiving* projections.
func (pj *Prjn) WtFmDWt(ctx *Context) {
	rlay := pj.Recv
	for ri := range rlay.Neurons {
		syns := pj.RecvSyns(ri)
		for ci := range syns {
			sy := &syns[ci]
			pj.Params.WtFmDWtSyn(ctx, sy)
		}
	}
}

// SlowAdapt does the slow adaptation: SWt learning and SynScale
func (pj *Prjn) SlowAdapt(ctx *Context) {
	pj.SWtFmWt()
	pj.SynScale()
}

// SWtFmWt updates structural, slowly-adapting SWt value based on
// accumulated DSWt values, which are zero-summed with additional soft bounding
// relative to SWt limits.
func (pj *Prjn) SWtFmWt() {
	if pj.Params.Learn.Learn.IsFalse() || pj.Params.SWt.Adapt.On.IsFalse() {
		return
	}
	rlay := pj.Recv
	if rlay.Params.IsTarget() {
		return
	}
	max := pj.Params.SWt.Limit.Max
	min := pj.Params.SWt.Limit.Min
	lr := pj.Params.SWt.Adapt.LRate
	dvar := pj.Params.SWt.Adapt.DreamVar
	for ri := range rlay.Neurons {
		syns := pj.RecvSyns(ri)
		nCons := len(syns)
		if nCons < 1 {
			continue
		}
		avgDWt := float32(0)
		for ci := range syns {
			sy := &syns[ci]
			if sy.DSWt >= 0 { // softbound for SWt
				sy.DSWt *= (max - sy.SWt)
			} else {
				sy.DSWt *= (sy.SWt - min)
			}
			avgDWt += sy.DSWt
		}
		avgDWt /= float32(nCons)
		avgDWt *= pj.Params.SWt.Adapt.SubMean
		if dvar > 0 {
			for ci := range syns {
				sy := &syns[ci]
				sy.SWt += lr * (sy.DSWt - avgDWt)
				sy.DSWt = 0
				if sy.Wt == 0 { // restore failed wts
					sy.Wt = pj.Params.SWt.WtVal(sy.SWt, sy.LWt)
				}
				sy.LWt = pj.Params.SWt.LWtFmWts(sy.Wt, sy.SWt) // + pj.Params.SWt.Adapt.RndVar()
				sy.Wt = pj.Params.SWt.WtVal(sy.SWt, sy.LWt)
			}
		} else {
			for ci := range syns {
				sy := &syns[ci]
				sy.SWt += lr * (sy.DSWt - avgDWt)
				sy.DSWt = 0
				if sy.Wt == 0 { // restore failed wts
					sy.Wt = pj.Params.SWt.WtVal(sy.SWt, sy.LWt)
				}
				sy.LWt = pj.Params.SWt.LWtFmWts(sy.Wt, sy.SWt)
				sy.Wt = pj.Params.SWt.WtVal(sy.SWt, sy.LWt)
			}
		}
	}
}

// SynScale performs synaptic scaling based on running average activation vs. targets.
// Layer-level AvgDifFmTrgAvg function must be called first.
func (pj *Prjn) SynScale() {
	if pj.Params.Learn.Learn.IsFalse() || pj.Params.IsInhib() {
		return
	}
	rlay := pj.Recv
	if !rlay.Params.IsLearnTrgAvg() {
		return
	}
	tp := &rlay.Params.Learn.TrgAvgAct
	lr := tp.SynScaleRate
	for ri := range rlay.Neurons {
		nrn := &rlay.Neurons[ri]
		if nrn.IsOff() {
			continue
		}
		adif := -lr * nrn.AvgDif
		syns := pj.RecvSyns(ri)
		for ci := range syns {
			sy := &syns[ci]
			if adif >= 0 { // key to have soft bounding on lwt here!
				sy.LWt += (1 - sy.LWt) * adif * sy.SWt
			} else {
				sy.LWt += sy.LWt * adif * sy.SWt
			}
			sy.Wt = pj.Params.SWt.WtVal(sy.SWt, sy.LWt)
		}
	}
}

// SynFail updates synaptic weight failure only -- normally done as part of DWt
// and WtFmDWt, but this call can be used during testing to update failing synapses.
func (pj *Prjn) SynFail(ctx *Context) {
	rlay := pj.Recv
	for ri := range rlay.Neurons {
		syns := pj.RecvSyns(ri)
		for ci := range syns {
			sy := &syns[ci]
			if sy.Wt == 0 { // restore failed wts
				sy.Wt = pj.Params.SWt.WtVal(sy.SWt, sy.LWt)
			}
			pj.Params.Com.Fail(ctx, &sy.Wt, sy.SWt)
		}
	}
}

// LRateMod sets the LRate modulation parameter for Prjns, which is
// for dynamic modulation of learning rate (see also LRateSched).
// Updates the effective learning rate factor accordingly.
func (pj *Prjn) LRateMod(mod float32) {
	pj.Params.Learn.LRate.Mod = mod
	pj.Params.Learn.LRate.Update()
}

// LRateSched sets the schedule-based learning rate multiplier.
// See also LRateMod.
// Updates the effective learning rate factor accordingly.
func (pj *Prjn) LRateSched(sched float32) {
	pj.Params.Learn.LRate.Sched = sched
	pj.Params.Learn.LRate.Update()
}
