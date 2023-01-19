// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// performs the DWt function on all sending projections

#include "time.hlsl"
#include "layerparams.hlsl"
#include "prjnparams.hlsl"

// note: binding is var, set
[[vk::binding(0, 0)]] uniform LayerParams Layers[]; // [Layer]
[[vk::binding(1, 0)]] uniform PrjnParams SendPrjns[]; // [Layer][SendPrjns]
[[vk::binding(2, 0)]] uniform PrjnParams RecvPrjns[]; // [Layer][RecvPrjns]

[[vk::binding(0, 1)]] StructuredBuffer<NeurSynIdx> SendNeurSynIdxs; // [Layer][SendPrjns][SendNeurs]
// [[vk::binding(1, 1)]] StructuredBuffer<NeurSynIdx> RecvNeurSynIdxs; // [Layer][RecvPrjns][RecvNeurs]
// [[vk::binding(2, 1)]] StructuredBuffer<SynIdx> RecvSynIdxs; // [Layer][RecvPrjns][Recv Neurs][Syns]

[[vk::binding(0, 2)]] StructuredBuffer<Time> CTime; // [0]
[[vk::binding(1, 2)]] RWStructuredBuffer<Neuron> Neurons; // [Layer][Neuron]
[[vk::binding(2, 2)]] RWStructuredBuffer<Synapse> Synapses;  // [Layer][SendPrjns][SendNeurs][Syns]
// [[vk::binding(3, 2)]] RWStructuredBuffer<Pool> Pools; // [Layer][Pools]
// [[vk::binding(4, 2)]] RWStructuredBuffer<LayerVals> LayVals; // [Layer]
// [[vk::binding(5, 2)]] RWStructuredBuffer<PrjnVals> PrjnVals; // [Layer][SendPrjns]

void DWtSyn(uint si, in PrjnParams pj, inout Synapse sy, in Neuron sn, in Neuron rn, bool isTarget, in Time ctime) {
	pj.DWtSyn(sy, sn, rn, isTarget, ctime);
}

void DWtSendNeurSyn2(uint snsi, in NeurSynIdx nsi, in LayerParams ly, in PrjnParams pj, in Time ctime) {
	if(pj.Learn.Learn == 0) {
		return;
	}
	bool isTarget = (ly.Act.Clamp.IsTarget == 1);
	uint nc = nsi.SynN;
	uint st = nsi.SynSt;
	for(uint si = 0; si < nc; si++) {
		uint sia = si + st;
		DWtSyn(sia, pj, Synapses[sia], Neurons[nsi.NeurIdx], Neurons[Synapses[sia].RecvNeurIdx], isTarget, ctime);
	}
}

void DWtSendNeurSyn(uint snsi, in NeurSynIdx nsi, in Time ctime) {
	DWtSendNeurSyn2(snsi, nsi, Layers[SendPrjns[nsi.PrjnIdx].Idxs.RecvLay], SendPrjns[nsi.PrjnIdx], ctime);
}


[numthreads(64, 1, 1)]
void main(uint3 idx : SV_DispatchThreadID) { // over SendNeurSynIdxs
	uint ns;
	uint st;
	SendNeurSynIdxs.GetDimensions(ns, st);
	if(idx.x < ns) {
		DWtSendNeurSyn(idx.x, SendNeurSynIdxs[idx.x], CTime[0]);
	}
}



