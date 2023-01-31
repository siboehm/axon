// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package axon

import (
	"embed"
	"unsafe"

	"github.com/goki/gi/oswin"
	"github.com/goki/ki/ints"
	"github.com/goki/vgpu/vgpu"
)

//go:embed shaders/*.spv
var content embed.FS

//go:generate gosl -exclude=Update,UpdateParams,Defaults,AllParams github.com/goki/mat32/fastexp.go github.com/emer/etable/minmax ../chans/chans.go ../chans ../kinase ../fsfffb/inhib.go ../fsfffb github.com/emer/emergent/etime github.com/emer/emergent/ringidx neuromod.go context.go neuron.go synapse.go pool.go layervals.go act.go act_prjn.go inhib.go learn.go layertypes.go layerparams.go deep_layers.go rl_layers.go pvlv_layers.go pcore_layers.go prjntypes.go prjnparams.go deep_prjns.go rl_prjns.go pvlv_prjns.go pcore_prjns.go gpu_gather.hlsl gpu_poolgemax.hlsl gpu_poolgi.hlsl gpu_cycle.hlsl gpu_synca.hlsl gpu_dwt.hlsl gpu_wtfmdwt.hlsl

// Full vars code -- each gpu_*.hlsl uses a subset

/*

// note: binding is var, set

// Set 0: uniforms -- these are constant
[[vk::binding(0, 0)]] uniform LayerParams Layers[]; // [Layer]
[[vk::binding(1, 0)]] uniform PrjnParams Prjns[]; // [Layer][RecvPrjns]

// Set 1: main network structs and vals -- all are writable
[[vk::binding(0, 1)]] StructuredBuffer<Context> Ctxt; // [0]
[[vk::binding(1, 1)]] RWStructuredBuffer<Neuron> Neurons; // [Layer][Neuron]
[[vk::binding(2, 1)]] RWStructuredBuffer<Pool> Pools; // [Layer][Pools]
[[vk::binding(3, 1)]] RWStructuredBuffer<LayerVals> LayVals; // [Layer]
[[vk::binding(4, 1)]] RWStructuredBuffer<Synapse> Synapses;  // [Layer][RecvPrjns][RecvNeurons][Syns]
[[vk::binding(5, 1)]] RWStructuredBuffer<float> GBuf;  // [Layer][RecvPrjns][RecvNeurons][MaxDel+1]
[[vk::binding(6, 1)]] RWStructuredBuffer<float> GSyns;  // [Layer][RecvPrjns][RecvNeurons]

// Set 2: prjn, synapse level indexes -- read only but too big for uniform probably?
[[vk::binding(0, 2)]] StructuredBuffer<StartN> RecvCon; // [Layer][RecvPrjns][RecvNeurons]

*/

/*
cycle update:

GatherSpikes & RecvSpikes:
gpu_gather.hlsl 	[Neurons]

GiFmSpikes:
gpu_poolgemax.hlsl [Pools] -- does GeToPool and AvgMax Update, calls LayPoolGiFmSpikes
gpu_poolgi.hlsl 	  [Pools] -- only operates on sub-Pools, does SubPoolGiFmSpikes

CycleNeuron:
gpu_cycle.hlsl 	[Neurons]

gpu_synca.hlsl	[Synapses]

DWt:
gpu_dwt.hlsl		[Synapses]
todo: submean
gpu_wtfmdwt.hlsl	[Synapses]

todo: plus phase!

todo: need apply inputs tensors!

*/

// GPU manages all of the GPU-based computation
type GPU struct {
	On           bool           `desc:"if true, actually use the GPU"`
	GPU          *vgpu.GPU      `desc:"the vgpu Vulkan GPU"`
	Sys          *vgpu.System   `desc:"the vgpu compute system"`
	Params       *vgpu.VarSet   `desc:"VarSet = 0: the uniform LayerParams, PrjnParams "`
	Structs      *vgpu.VarSet   `desc:"VarSet = 1: the Storage buffer for RW state structs "`
	Idxs         *vgpu.VarSet   `desc:"Varset = 2: the Storage buffer for large read-only index vals"`
	GatherSpikes *vgpu.Pipeline `desc:"GatherSpikes pipeline"`
	PoolGeMax    *vgpu.Pipeline `desc:"PoolGeMax pipeline"`
	PoolGi       *vgpu.Pipeline `desc:"PoolG pipeline"`
	Cycle        *vgpu.Pipeline `desc:"Cycle pipeline"`
	SynCa        *vgpu.Pipeline `desc:"SynCa pipeline"`
	DWt          *vgpu.Pipeline `desc:"DWt pipeline"`
	WtFmDWt      *vgpu.Pipeline `desc:"WtFmDWt pipeline"`
	NThreads     int            `desc:"number of warp threads -- typically 64 -- must update all hlsl files if changed!"`
}

// GPUOnGUI turns on GPU mode in context of GUI active, configures the GPU -- call after all built,
// initialized, params are set, and ready to run
func (nt *Network) GPUOnGUI(ctx *Context) {
	oswin.TheApp.RunOnMain(func() {
		nt.GPU.Config(ctx, nt)
	})
	nt.GPU.On = true
}

// GPUOnNoGUI turns on GPU mode in context of NO GUI active,
// configures the GPU -- call after all built,
// initialized, params are set, and ready to run
func (nt *Network) GPUOnNoGUI(ctx *Context) {
	if vgpu.InitNoDisplay() != nil {
		return
	}
	nt.GPU.Config(ctx, nt)
	nt.GPU.On = true
}

// Config configures the network -- must call on an already-built network
func (gp *GPU) Config(ctx *Context, net *Network) {
	gp.NThreads = 64
	gp.GPU = vgpu.NewComputeGPU()
	// vgpu.Debug = true
	gp.GPU.Config("axon")

	gp.Sys = gp.GPU.NewComputeSystem("axon")

	vars := gp.Sys.Vars()
	gp.Params = vars.AddSet()
	gp.Structs = vars.AddSet()
	gp.Idxs = vars.AddSet()

	gp.Params.AddStruct("Layers", int(unsafe.Sizeof(LayerParams{})), len(net.LayParams), vgpu.Uniform, vgpu.ComputeShader)
	gp.Params.AddStruct("Prjns", int(unsafe.Sizeof(PrjnParams{})), len(net.PrjnParams), vgpu.Uniform, vgpu.ComputeShader)

	gp.Structs.AddStruct("Ctxt", int(unsafe.Sizeof(Context{})), 1, vgpu.Storage, vgpu.ComputeShader)
	gp.Structs.AddStruct("Neurons", int(unsafe.Sizeof(Neuron{})), len(net.Neurons), vgpu.Storage, vgpu.ComputeShader)
	gp.Structs.AddStruct("Pools", int(unsafe.Sizeof(Pool{})), len(net.Pools), vgpu.Storage, vgpu.ComputeShader)
	gp.Structs.AddStruct("LayVals", int(unsafe.Sizeof(LayerVals{})), len(net.LayVals), vgpu.Storage, vgpu.ComputeShader)
	gp.Structs.AddStruct("Synapses", int(unsafe.Sizeof(Synapse{})), len(net.Synapses), vgpu.Storage, vgpu.ComputeShader)
	gp.Structs.Add("GBuf", vgpu.Float32, len(net.PrjnGBuf), vgpu.Storage, vgpu.ComputeShader)
	gp.Structs.Add("GSyns", vgpu.Float32, len(net.PrjnGSyns), vgpu.Storage, vgpu.ComputeShader)

	gp.Idxs.AddStruct("RecvCon", int(unsafe.Sizeof(StartN{})), len(net.PrjnRecvCon), vgpu.Storage, vgpu.ComputeShader)

	gp.Params.ConfigVals(1)
	gp.Structs.ConfigVals(1)
	gp.Idxs.ConfigVals(1)

	// todo: embed .spv into binary

	// pipelines
	gp.GatherSpikes = gp.Sys.NewPipeline("GatherSpikes")
	cb, _ := content.ReadFile("shaders/gpu_gather.spv")
	gp.GatherSpikes.AddShaderCode("GatherSpikes", vgpu.ComputeShader, cb)

	gp.PoolGeMax = gp.Sys.NewPipeline("PoolGeMax")
	cb, _ = content.ReadFile("shaders/gpu_poolgemax.spv")
	gp.PoolGeMax.AddShaderCode("PoolGeMax", vgpu.ComputeShader, cb)

	gp.PoolGi = gp.Sys.NewPipeline("PoolGi")
	cb, _ = content.ReadFile("shaders/gpu_poolgi.spv")
	gp.PoolGi.AddShaderCode("PoolGi", vgpu.ComputeShader, cb)

	gp.Cycle = gp.Sys.NewPipeline("Cycle")
	cb, _ = content.ReadFile("shaders/gpu_cycle.spv")
	gp.Cycle.AddShaderCode("Cycle", vgpu.ComputeShader, cb)

	gp.SynCa = gp.Sys.NewPipeline("SynCa")
	cb, _ = content.ReadFile("shaders/gpu_synca.spv")
	gp.SynCa.AddShaderCode("SynCa", vgpu.ComputeShader, cb)

	gp.DWt = gp.Sys.NewPipeline("DWt")
	cb, _ = content.ReadFile("shaders/gpu_dwt.spv")
	gp.DWt.AddShaderCode("DWt", vgpu.ComputeShader, cb)

	gp.WtFmDWt = gp.Sys.NewPipeline("WtFmDWt")
	cb, _ = content.ReadFile("shaders/gpu_wtfmdwt.spv")
	gp.WtFmDWt.AddShaderCode("WtFmDWt", vgpu.ComputeShader, cb)

	gp.Sys.Config()

	gp.CopyParamsToGPU(ctx, net)
	gp.CopyIdxsToGPU(ctx, net)
	gp.CopyContextToGPU(ctx, net)
	gp.CopyNeuronsToGPU(ctx, net)
	gp.CopyStateToGPU(ctx, net)

	gp.Sys.Mem.SyncToGPU()

	// todo: add a convenience method to vgpu to do this for everything
	vars.BindDynValIdx(0, "Layers", 0)
	vars.BindDynValIdx(0, "Prjns", 0)

	vars.BindDynValIdx(1, "Ctxt", 0)
	vars.BindDynValIdx(1, "Neurons", 0)
	vars.BindDynValIdx(1, "Pools", 0)
	vars.BindDynValIdx(1, "LayVals", 0)
	vars.BindDynValIdx(1, "Synapses", 0)
	vars.BindDynValIdx(1, "GBuf", 0)
	vars.BindDynValIdx(1, "GSyns", 0)

	vars.BindDynValIdx(2, "RecvCon", 0)
}

func (gp *GPU) CopyParamsToGPU(ctx *Context, net *Network) {
	_, layv, _ := gp.Params.ValByIdxTry("Layers", 0)
	layv.CopyToBytes(unsafe.Pointer(&net.LayParams[0]))

	_, pjnv, _ := gp.Params.ValByIdxTry("Prjns", 0)
	pjnv.CopyToBytes(unsafe.Pointer(&net.PrjnParams[0]))
}

func (gp *GPU) CopyIdxsToGPU(ctx *Context, net *Network) {
	_, rconv, _ := gp.Idxs.ValByIdxTry("RecvCon", 0)
	rconv.CopyToBytes(unsafe.Pointer(&net.PrjnRecvCon[0]))
}

func (gp *GPU) CopyContextToGPU(ctx *Context, net *Network) {
	_, ctxv, _ := gp.Structs.ValByIdxTry("Ctxt", 0)
	ctxv.CopyToBytes(unsafe.Pointer(ctx))
}

func (gp *GPU) CopyNeuronsToGPU(ctx *Context, net *Network) {
	_, neurv, _ := gp.Structs.ValByIdxTry("Neurons", 0)
	neurv.CopyToBytes(unsafe.Pointer(&net.Neurons[0]))
}

func (gp *GPU) CopyStateToGPU(ctx *Context, net *Network) {
	_, poolv, _ := gp.Structs.ValByIdxTry("Pools", 0)
	poolv.CopyToBytes(unsafe.Pointer(&net.Pools[0]))

	_, layv, _ := gp.Structs.ValByIdxTry("LayVals", 0)
	layv.CopyToBytes(unsafe.Pointer(&net.LayVals[0]))

	_, synv, _ := gp.Structs.ValByIdxTry("Synapses", 0)
	synv.CopyToBytes(unsafe.Pointer(&net.Synapses[0]))
}

func (gp *GPU) SynMemToGPU() {
	gp.Sys.Mem.SyncToGPU()
}

func (gp *GPU) NSub(n int) int {
	return ints.IntMultiple(n, gp.NThreads) / gp.NThreads
}

func (gp *GPU) RunPipeline(net *Network, name string, pl *vgpu.Pipeline, n int) {
	// todo: need to bind vars again?
	net.FunTimerStart(name)
	// gp.Sys.CmdResetBindVars(gp.Sys.CmdPool.Buff, 0)
	gp.Sys.ComputeResetBegin()
	pl.ComputeCommand(gp.NSub(n), 1, 1)
	gp.Sys.ComputeSubmitWait()
	net.FunTimerStop(name)
}

func (gp *GPU) RunCycle(ctx *Context, net *Network) {
	gp.RunPipeline(net, "GPU:GatherSpikes", gp.GatherSpikes, len(net.Neurons))

	// todo: use semaphors for all of these instead of waits
	// todo: need to bind vars again?
	gp.RunPipeline(net, "GPU:PoolGeMax", gp.PoolGeMax, len(net.Pools))

	gp.RunPipeline(net, "GPU:PoolGi", gp.PoolGi, len(net.Pools))

	gp.RunPipeline(net, "GPU:Cycle", gp.Cycle, len(net.Neurons))

	if ctx.Testing.IsFalse() {
		gp.RunPipeline(net, "GPU:SynCa", gp.SynCa, len(net.Synapses))
	}
}

func (gp *GPU) RunDWt(ctx *Context, net *Network) {
	gp.RunPipeline(net, "GPU:DWt", gp.DWt, len(net.Synapses))
}

func (gp *GPU) RunWtFmDWt(ctx *Context, net *Network) {
	gp.RunPipeline(net, "GPU:WtFmDWt", gp.WtFmDWt, len(net.Synapses))
}
