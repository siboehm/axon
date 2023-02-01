// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "context.hlsl"
#include "layerparams.hlsl"

// calls ApplyExt on neurons

// note: binding is var, set

// Set 0: uniforms -- these are constant
[[vk::binding(0, 0)]] uniform LayerParams Layers[]; // [Layer]
// [[vk::binding(1, 0)]] uniform PrjnParams Prjns[]; // [Layer][RecvPrjns]

// Set 1: main network structs and vals
[[vk::binding(0, 1)]] StructuredBuffer<Context> Ctxt; // [0]
[[vk::binding(1, 1)]] RWStructuredBuffer<Neuron> Neurons; // [Layer][Neuron]
// [[vk::binding(2, 1)]] RWStructuredBuffer<Pool> Pools; // [Layer][Pools]
// [[vk::binding(3, 1)]] RWStructuredBuffer<LayerVals> LayVals; // [Layer]
// [[vk::binding(4, 1)]] RWStructuredBuffer<Synapse> Synapses;  // [Layer][RecvPrjns][RecvNeurons][Syns]
// [[vk::binding(5, 1)]] RWStructuredBuffer<float> GBuf;  // [Layer][RecvPrjns][RecvNeurons][MaxDel+1]
// [[vk::binding(6, 1)]] RWStructuredBuffer<float> GSyns;  // [Layer][RecvPrjns][RecvNeurons]

// Set 2: prjn, synapse level indexes and buffer values
// [[vk::binding(0, 2)]] StructuredBuffer<StartN> RecvCon; // [Layer][RecvPrjns][RecvNeurons]

// Set 3: external inputs
[[vk::binding(0, 3)]] RWStructuredBuffer<float> Exts;  // [In / Out Layers][Neurons]

void ApplyExt2(in Context ctx, in LayerParams ly, uint nin, inout Neuron nrn) {
	uint ni = nin - ly.Idxs.NeurSt; // layer-based as in Go
	ly.InitExt(ni, nrn);
	if (IsExtLayerType(ly.LayType)) {
		ly.ApplyExtVal(ni, nrn, Exts[ly.Idxs.ExtsSt + ni]);
	}
}

void ApplyExt(in Context ctx, uint nin, inout Neuron nrn) {
	ApplyExt2(ctx, Layers[nrn.LayIdx], nin, nrn);
}

[numthreads(64, 1, 1)]
void main(uint3 idx : SV_DispatchThreadID) { // over Neurons
	uint ns;
	uint st;
	Neurons.GetDimensions(ns, st);
	if(idx.x < ns) {
		ApplyExt(Ctxt[0], idx.x, Neurons[idx.x]);
	}
}

