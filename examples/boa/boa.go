// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
boa: This project tests BG, OFC & ACC learning in a CS-driven approach task.
*/
package main

import (
	"fmt"
	"log"
	"math"
	"os"

	"github.com/emer/axon/axon"
	"github.com/emer/emergent/ecmd"
	"github.com/emer/emergent/egui"
	"github.com/emer/emergent/elog"
	"github.com/emer/emergent/emer"
	"github.com/emer/emergent/env"
	"github.com/emer/emergent/erand"
	"github.com/emer/emergent/estats"
	"github.com/emer/emergent/etime"
	"github.com/emer/emergent/looper"
	"github.com/emer/emergent/netview"
	"github.com/emer/emergent/prjn"
	"github.com/emer/emergent/relpos"
	"github.com/emer/empi/mpi"
	"github.com/emer/etable/agg"
	"github.com/emer/etable/etable"
	"github.com/emer/etable/etensor"
	"github.com/emer/etable/minmax"
	"github.com/emer/etable/split"
	"github.com/goki/gi/gi"
	"github.com/goki/gi/gimain"
	"github.com/goki/ki/bools"
	"github.com/goki/mat32"
)

var (
	// Debug triggers various messages etc
	Debug = false
	// GPU runs with the GPU (for demo, testing -- not useful for such a small network)
	GPU = false
)

func main() {
	TheSim.New()
	TheSim.Config()
	if len(os.Args) > 1 {
		TheSim.CmdArgs() // simple assumption is that any args = no gui -- could add explicit arg if you want
	} else {
		gimain.Main(func() { // this starts gui -- requires valid OpenGL display connection (e.g., X11)
			guirun()
		})
	}
}

func guirun() {
	TheSim.Init()
	win := TheSim.ConfigGui()
	win.StartEventLoop()
}

// see params.go for network params

// SimParams has all the custom params for this sim
type SimParams struct {
	PctCortex         float32 `desc:"proportion of behavioral approach sequences driven by the cortex vs. hard-coded reflexive subcortical"`
	PctCortexMax      float32 `desc:"maximum PctCortex, when running on the schedule"`
	PctCortexStEpc    int     `desc:"epoch when PctCortex starts increasing"`
	PctCortexNEpc     int     `desc:"number of epochs over which PctCortexMax is reached"`
	PctCortexInterval int     `desc:"how often to update PctCortex"`
	PCAInterval       int     `desc:"how frequently (in epochs) to compute PCA on hidden representations to measure variance?"`
	CortexDriving     bool    `desc:"true if cortex is driving this behavioral approach sequence"`
}

// Defaults sets default params
func (ss *SimParams) Defaults() {
	ss.PctCortexMax = 1.0
	ss.PctCortexStEpc = 10
	ss.PctCortexNEpc = 10
	ss.PctCortexInterval = 1
	ss.PCAInterval = 10
}

// Sim encapsulates the entire simulation model, and we define all the
// functionality as methods on this struct.  This structure keeps all relevant
// state information organized and available without having to pass everything around
// as arguments to methods, and provides the core GUI interface (note the view tags
// for the fields which provide hints to how things should be displayed).
type Sim struct {
	Net          *axon.Network    `view:"no-inline" desc:"the network -- click to view / edit parameters for layers, prjns, etc"`
	Sim          SimParams        `view:"no-inline" desc:"sim params"`
	StopOnErr    bool             `desc:"if true, stop running when an error programmed into the code occurs"`
	Params       emer.Params      `view:"inline" desc:"all parameter management"`
	Loops        *looper.Manager  `view:"no-inline" desc:"contains looper control loops for running sim"`
	Stats        estats.Stats     `desc:"contains computed statistic values"`
	Logs         elog.Logs        `desc:"Contains all the logs and information about the logs.'"`
	Pats         *etable.Table    `view:"no-inline" desc:"the training patterns to use"`
	Envs         env.Envs         `view:"no-inline" desc:"Environments"`
	Context      axon.Context     `desc:"axon timing parameters and state"`
	ViewUpdt     netview.ViewUpdt `view:"inline" desc:"netview update parameters"`
	TestInterval int              `desc:"how often to run through all the test patterns, in terms of training epochs -- can use 0 or -1 for no testing"`

	GUI      egui.GUI    `view:"-" desc:"manages all the gui elements"`
	Args     ecmd.Args   `view:"no-inline" desc:"command line args"`
	RndSeeds erand.Seeds `view:"-" desc:"a list of random seeds to use for each run"`
}

// TheSim is the overall state for this simulation
var TheSim Sim

// New creates new blank elements and initializes defaults
func (ss *Sim) New() {
	ss.Net = &axon.Network{}
	ss.Sim.Defaults()
	ss.Params.Params = ParamSets
	// ss.Params.ExtraSets = "WtScales"
	ss.Params.AddNetwork(ss.Net)
	ss.Params.AddSim(ss)
	ss.Params.AddNetSize()
	ss.Pats = &etable.Table{}
	ss.Stats.Init()
	ss.RndSeeds.Init(100) // max 100 runs
	ss.TestInterval = 500
	ss.Context.Defaults()
	ss.Context.PVLV.Effort.Gain = 0.05
	ss.ConfigArgs() // do this first, has key defaults
}

////////////////////////////////////////////////////////////////////////////////////////////
// 		Configs

// Config configures all the elements using the standard functions
func (ss *Sim) Config() {
	ss.ConfigEnv()
	ss.ConfigNet(ss.Net)
	ss.ConfigLogs()
	ss.ConfigLoops()
}

func (ss *Sim) ConfigEnv() {
	// Can be called multiple times -- don't re-create
	var trn, tst *Approach
	if len(ss.Envs) == 0 {
		trn = &Approach{}
		tst = &Approach{}
	} else {
		trn = ss.Envs.ByMode(etime.Train).(*Approach)
		tst = ss.Envs.ByMode(etime.Test).(*Approach)
	}

	// note: names must be standard here!
	trn.Nm = etime.Train.String()
	trn.Defaults()
	trn.Config()
	trn.Validate()

	ss.Context.PVLV.Drive.NActive = int32(trn.NDrives)

	tst.Nm = etime.Test.String()
	tst.Defaults()
	tst.Config()
	tst.Validate()

	trn.Init(0)
	tst.Init(0)

	// note: names must be in place when adding
	ss.Envs.Add(trn, tst)
}

func (ss *Sim) ConfigNet(net *axon.Network) {
	ev := ss.Envs["Train"].(*Approach)
	net.InitName(net, "Boa")

	nuBgY := 5
	nuBgX := 5
	nuCtxY := 6
	nuCtxX := 6
	nAct := ev.NActs
	popY := 4
	popX := 4
	space := float32(2)

	pone2one := prjn.NewPoolOneToOne()
	one2one := prjn.NewOneToOne()
	full := prjn.NewFull()
	mtxRndPrjn := prjn.NewPoolUnifRnd()
	mtxRndPrjn.PCon = 0.75
	_ = mtxRndPrjn
	_ = pone2one

	ny := ev.NYReps
	nloc := ev.Locations

	vta, lhb, ach := net.AddVTALHbAChLayers(relpos.Behind, space)
	_ = lhb
	_ = ach

	vPmtxGo, vPmtxNo, _, _, vPgpeTA, vPstnp, vPstns, vPgpi := net.AddBG("Vp", 1, ev.NDrives, nuBgY, nuBgX, nuBgY, nuBgX, space)
	vsGated := net.AddVSGatedLayer("", ny)
	vsPatch := net.AddVSPatchLayer("", ev.NDrives, nuBgY, nuBgX)

	drives, drivesP, effort, effortP, usPos, usNeg, usPosP, usNegP, pvPos, pvNeg, pvPosP, pvNegP := net.AddDrivePVLVPulvLayers(&ss.Context, ev.NDrives, ny, popY, popX, space)
	_ = usNegP
	_ = usPos
	_ = usNeg
	_ = pvPos
	_ = pvNeg
	_ = pvNegP
	_ = effort
	_ = effortP
	_ = drivesP

	cs := net.AddLayer2D("CS", ny, ev.NDrives, axon.InputLayer)
	dist, distP := net.AddInputPulv2D("Dist", ny, ev.DistMax, space)
	pos := net.AddLayer2D("Pos", ny, nloc, axon.InputLayer) // irrelevant here

	m1 := net.AddLayer2D("M1", nuCtxY, nuCtxX, axon.SuperLayer)
	act := net.AddLayer2D("Act", ny, nAct, axon.InputLayer) // Action
	vl := net.AddPulvLayer2D("VL", ny, nAct)                // VL predicts brainstem Action: either its own or instinct
	vl.SetBuildConfig("DriveLayName", act.Name())

	m1P := net.AddPulvLayer2D("M1P", nuCtxY, nuCtxX)
	m1P.SetBuildConfig("DriveLayName", m1.Name())
	_ = vl
	_ = act

	blaPosA, blaPosE, _, _, cemPos, _, pptg := net.AddAmygdala("", false, ev.NDrives, nuCtxY, nuCtxX, space)
	_ = cemPos
	_ = pptg
	// blaPosA.SetBuildConfig("LayInhib1Name", blaPosE.Name()) // just for testing
	// blaPosE.SetBuildConfig("LayInhib1Name", blaPosA.Name())

	ofc, ofcCT := net.AddSuperCT4D("OFC", 1, ev.NDrives, nuCtxY, nuCtxX, space, one2one)
	// prjns are: super->PT, PT self, CT-> thal
	ofcPT, ofcMD := net.AddPTMaintThalForSuper(ofc, ofcCT, "MD", one2one, pone2one, pone2one, space)
	_ = ofcPT
	ofcCT.SetClass("OFC CTCopy")
	ofcPTPred := net.AddPTPredLayer(ofcPT, ofcCT, ofcMD, pone2one, pone2one, pone2one, space)
	ofcPTPred.SetClass("OFC")
	notMaint := net.AddPTNotMaintLayer(ofcPT, ny, 1, space)
	notMaint.Nm = "NotMaint"

	// net.ConnectCTSelf(ofcCT, pone2one) // much better for ofc not to have self prjns..
	// net.ConnectToPulv(ofc, ofcCT, csp, full, full) // todo: test
	net.ConnectToPulv(ofc, ofcCT, usPosP, pone2one, pone2one)
	net.ConnectToPulv(ofc, ofcCT, pvPosP, pone2one, pone2one)
	net.ConnectToPulv(ofc, ofcCT, drivesP, pone2one, pone2one)

	net.ConnectPTPredToPulv(ofcPTPred, usPosP, pone2one, pone2one)
	net.ConnectPTPredToPulv(ofcPTPred, pvPosP, pone2one, pone2one)
	net.ConnectPTPredToPulv(ofcPTPred, drivesP, pone2one, pone2one)

	// Drives -> OFC then activates OFC -> VS -- OFC needs to be strongly BLA dependent
	// to reflect either current CS or maintained CS but not just echoing drive state.
	net.ConnectLayers(drives, ofc, pone2one, axon.ForwardPrjn).SetClass("DrivesToOFC")

	// net.ConnectLayers(drives, ofcCT, pone2one, axon.ForwardPrjn).SetClass("DrivesToOFC")
	net.ConnectLayers(vPgpi, ofcMD, full, axon.InhibPrjn).SetClass("BgFixed")
	// net.ConnectLayers(cs, ofc, full, axon.ForwardPrjn) // let BLA handle it
	net.ConnectLayers(usPos, ofc, pone2one, axon.BackPrjn)
	net.ConnectLayers(ofcPT, ofcCT, pone2one, axon.ForwardPrjn) // good?

	net.ConnectLayers(usPos, ofcPTPred, pone2one, axon.ForwardPrjn)
	net.ConnectLayers(pvPos, ofcPTPred, pone2one, axon.ForwardPrjn)
	net.ConnectLayers(drives, ofcPTPred, pone2one, axon.ForwardPrjn)

	acc, accCT := net.AddSuperCT2D("ACC", nuCtxY+2, nuCtxX+2, space, one2one)
	// prjns are: super->PT, PT self, CT->thal
	accPT, accMD := net.AddPTMaintThalForSuper(acc, accCT, "MD", one2one, full, full, space)
	_ = accPT
	accCT.SetClass("ACC CTCopy")
	accPTPred := net.AddPTPredLayer(accPT, accCT, accMD, full, full, full, space)
	accPTPred.SetClass("ACC")
	net.ConnectPTNotMaint(accPT, notMaint, full)

	net.ConnectCTSelf(accCT, full)
	net.ConnectToPulv(acc, accCT, distP, full, full)
	net.ConnectToPulv(acc, accCT, effortP, full, full)
	net.ConnectLayers(vPgpi, accMD, full, axon.InhibPrjn).SetClass("BgFixed")

	net.ConnectPTPredToPulv(accPTPred, distP, full, full)
	net.ConnectPTPredToPulv(accPTPred, effortP, full, full)

	net.ConnectLayers(dist, acc, full, axon.ForwardPrjn)
	net.ConnectLayers(effort, acc, full, axon.ForwardPrjn)
	net.ConnectLayers(accPT, accCT, full, axon.ForwardPrjn) // good?

	net.ConnectLayers(dist, accPTPred, full, axon.ForwardPrjn)
	net.ConnectLayers(effort, accPTPred, full, axon.ForwardPrjn)

	net.ConnectLayers(acc, ofc, full, axon.BackPrjn)
	net.ConnectLayers(ofc, acc, full, axon.BackPrjn)

	vPmtxGo.SetBuildConfig("ThalLay1Name", ofcMD.Name())
	vPmtxNo.SetBuildConfig("ThalLay1Name", ofcMD.Name())
	vPmtxGo.SetBuildConfig("ThalLay2Name", accMD.Name())
	vPmtxNo.SetBuildConfig("ThalLay2Name", accMD.Name())

	// m1P plus phase has action, Ctxt -> CT allows CT now to use that prev action

	alm, almCT := net.AddSuperCT2D("ALM", nuCtxY, nuCtxX, space, one2one)
	// note: not doing full alm / dlPFC yet
	// almpt, almthal := net.AddPTThalForSuper(alm, almCT, "MD", one2one, full, full, space)
	almCT.SetClass("ALM CTCopy")
	// net.ConnectCTSelf(almCT, full) // todo: test
	net.ConnectToPulv(alm, almCT, m1P, full, full)
	// net.ConnectToPulv(alm, almCT, distP, full, full) // todo: test -- should be used

	// todo: explore contextualization based on action
	// net.BidirConnectLayers(ofc, alm, full)
	// net.BidirConnectLayers(acc, alm, full)
	// net.ConnectLayers(ofcPT, alm, full, axon.ForwardPrjn)
	// net.ConnectLayers(accPT, alm, full, axon.ForwardPrjn)

	// todo: blaPosE is not connected properly at all yet

	// BLA
	net.ConnectToBLAAcq(cs, blaPosA, full)
	net.ConnectToBLAAcq(usPos, blaPosA, pone2one).SetClass("USToBLA")
	net.ConnectLayers(blaPosA, ofc, pone2one, axon.ForwardPrjn)
	// todo: from deep maint layer:
	// net.ConnectLayers(ofcPT, blaPosE, pone2one, axon.ForwardPrjn)
	net.ConnectLayers(blaPosE, blaPosA, pone2one, axon.InhibPrjn).SetClass("BgFixed")
	// net.ConnectToBLA(drives, blaPosA, pone2one).SetClass("USToBLA") // bla is not strongly drive mod!
	// net.ConnectLayers(drives, blaPosE, pone2one, axon.ForwardPrjn)
	net.ConnectToBLAExt(cs, blaPosE, full)
	net.ConnectToBLAExt(ofcPT, blaPosE, pone2one)

	net.ConnectLayers(dist, alm, full, axon.ForwardPrjn)
	net.ConnectLayers(effort, alm, full, axon.ForwardPrjn)
	// net.ConnectLayers(ofcPT, alm, full, axon.ForwardPrjn) // todo: test
	// net.ConnectLayers(accPT, alm, full, axon.ForwardPrjn)
	net.ConnectLayers(ofcPTPred, alm, full, axon.ForwardPrjn)
	net.ConnectLayers(accPTPred, alm, full, axon.ForwardPrjn)
	net.ConnectLayers(notMaint, alm, full, axon.ForwardPrjn)
	// net.ConnectLayers(pos, alm, full, axon.ForwardPrjn) // not relevant for action here.
	net.ConnectLayers(dist, m1, full, axon.ForwardPrjn).SetClass("ToM1")
	net.ConnectLayers(effort, m1, full, axon.ForwardPrjn).SetClass("ToM1")

	// net.ConnectLayers(ofcPT, m1, full, axon.ForwardPrjn) // todo: test
	// net.ConnectLayers(accPT, m1, full, axon.ForwardPrjn)
	net.ConnectLayers(ofcPTPred, m1, full, axon.ForwardPrjn)
	net.ConnectLayers(accPTPred, m1, full, axon.ForwardPrjn)
	net.ConnectLayers(notMaint, m1, full, axon.ForwardPrjn).SetClass("ToM1")

	// net.ConnectLayers(ofcPT, m1, full, axon.ForwardPrjn) // todo: test
	// net.ConnectLayers(accPT, m1, full, axon.ForwardPrjn)
	net.ConnectLayers(ofcPTPred, vl, full, axon.ForwardPrjn)
	net.ConnectLayers(accPTPred, vl, full, axon.ForwardPrjn)
	net.ConnectLayers(notMaint, vl, full, axon.ForwardPrjn).SetClass("ToVL")

	// key point: cs does not project directly to alm -- no simple S -> R mappings!?

	net.BidirConnectLayers(alm, m1, full) // todo: alm weaker?
	ff, fb := net.BidirConnectLayers(m1, vl, full)
	ff.SetClass("ToVL")
	fb.SetClass("ToM1")
	// net.BidirConnectLayers(alm, vl, full) // todo: test skip prjn
	// net.BidirConnectLayers(almCT, vl, full)

	net.ConnectLayers(vl, alm, full, axon.BackPrjn)
	net.ConnectLayers(vl, almCT, full, axon.BackPrjn)

	////////////////////////////////////////////////
	// BG / DA connections

	// same prjns to stn as mtxgo
	net.ConnectToMatrix(usPos, vPmtxGo, pone2one)
	net.ConnectToMatrix(blaPosA, vPmtxGo, mtxRndPrjn).SetClass("BLAToBG")
	net.ConnectToMatrix(blaPosA, vPmtxNo, mtxRndPrjn).SetClass("BLAToBG")
	net.ConnectLayers(blaPosA, vPstnp, full, axon.ForwardPrjn)
	net.ConnectLayers(blaPosA, vPstns, full, axon.ForwardPrjn)

	// net.ConnectToMatrix(blaPosE, vPmtxGo, pone2one) // todo: add!
	// net.ConnectToMatrix(blaPosE, vPmtxNo, pone2one)
	net.ConnectToMatrix(drives, vPmtxGo, pone2one).SetClass("DrivesToMtx") // modulatory in params
	net.ConnectToMatrix(drives, vPmtxNo, pone2one).SetClass("DrivesToMtx")
	net.ConnectToMatrix(ofc, vPmtxGo, pone2one)
	net.ConnectToMatrix(ofc, vPmtxNo, pone2one)
	// net.ConnectLayers(ofc, vPstnp, full, axon.ForwardPrjn) // todo: test
	// net.ConnectLayers(ofc, vPstns, full, axon.ForwardPrjn)
	net.ConnectToMatrix(acc, vPmtxGo, full)
	net.ConnectToMatrix(acc, vPmtxNo, full)
	// net.ConnectLayers(acc, vPstnp, full, axon.ForwardPrjn) // todo: test
	// net.ConnectLayers(acc, vPstns, full, axon.ForwardPrjn)

	net.ConnectToVSPatch(drives, vsPatch, pone2one).SetClass("DrivesToVSPatch") // modulatory
	net.ConnectToVSPatch(ofcPTPred, vsPatch, pone2one)
	net.ConnectToVSPatch(accPTPred, vsPatch, full)
	net.ConnectToVSPatch(almCT, vsPatch, full) // give access to motor action -- CT has prev action..

	////////////////////////////////////////////////
	// position

	vPgpi.PlaceRightOf(vta, space)

	vsPatch.PlaceRightOf(vPstns, space)
	vsGated.PlaceRightOf(vPgpeTA, space)

	usPos.PlaceAbove(vta)

	cs.PlaceRightOf(pvPos, space)
	dist.PlaceRightOf(cs, space)
	pos.PlaceRightOf(dist, space)

	m1.PlaceRightOf(pos, space)
	m1P.PlaceBehind(m1, space)
	vl.PlaceBehind(m1P, space)
	act.PlaceBehind(vl, space)

	blaPosA.PlaceAbove(usPos)
	ofc.PlaceRightOf(blaPosA, space)
	ofcMD.PlaceBehind(ofcPTPred, space)
	acc.PlaceRightOf(ofc, space)
	accMD.PlaceBehind(accPTPred, space)
	alm.PlaceRightOf(acc, space)
	notMaint.PlaceBehind(almCT, space)

	err := net.Build()
	if err != nil {
		log.Println(err)
		return
	}
	net.Defaults()
	ss.Params.SetObject("Network")
	ss.InitWts(net)
}

// InitWts configures initial weights according to structure
func (ss *Sim) InitWts(net *axon.Network) {
	net.InitWts()
	ss.ViewUpdt.RecordSyns() // note: critical to update weights here so DWt is visible
}

////////////////////////////////////////////////////////////////////////////////
// 	    Init, utils

// Init restarts the run, and initializes everything, including network weights
// and resets the epoch log table
func (ss *Sim) Init() {
	ss.Loops.ResetCounters()
	ss.InitRndSeed()
	// ss.ConfigEnv() // re-config env just in case a different set of patterns was
	// selected or patterns have been modified etc
	ss.GUI.StopNow = false
	ss.Params.SetAll()
	ss.NewRun()
	ss.ViewUpdt.Update()
	ss.ViewUpdt.RecordSyns()
}

// InitRndSeed initializes the random seed based on current training run number
func (ss *Sim) InitRndSeed() {
	run := ss.Loops.GetLoop(etime.Train, etime.Run).Counter.Cur
	ss.RndSeeds.Set(run)
}

// ConfigLoops configures the control loops: Training, Testing
func (ss *Sim) ConfigLoops() {
	man := looper.NewManager()

	ev := ss.Envs[ss.Context.Mode.String()].(*Approach)
	maxTrials := ev.TimeMax + 5 // allow for extra time for gating trials

	man.AddStack(etime.Train).AddTime(etime.Run, 5).AddTime(etime.Epoch, 40).AddTime(etime.Sequence, 25).AddTime(etime.Trial, maxTrials).AddTime(etime.Cycle, 200)

	// note: not using Test mode at this point, so just commenting all this out
	// in case there is a future need for it.
	// man.AddStack(etime.Test).AddTime(etime.Epoch, 1).AddTime(etime.Sequence, 25).AddTime(etime.Trial, maxTrials).AddTime(etime.Cycle, 200)

	axon.LooperStdPhases(man, &ss.Context, ss.Net, 150, 199)            // plus phase timing
	axon.LooperSimCycleAndLearn(man, ss.Net, &ss.Context, &ss.ViewUpdt) // std algo code

	man.GetLoop(etime.Train, etime.Trial).OnEnd.Replace("UpdateWeights", func() {
		ss.Net.DWt(&ss.Context)
		ss.ViewUpdt.RecordSyns() // note: critical to update weights here so DWt is visible
		ss.Net.WtFmDWt(&ss.Context)
	})

	for m, _ := range man.Stacks {
		mode := m // For closures
		stack := man.Stacks[mode]
		stack.Loops[etime.Trial].OnStart.Add("Env:Step", func() {
			// note: OnStart for env.Env, others may happen OnEnd
			ss.Envs[mode.String()].Step()
		})
		stack.Loops[etime.Trial].OnStart.Add("ApplyInputs", func() {
			ss.ApplyInputs()
		})
		stack.Loops[etime.Trial].OnEnd.Add("StatCounters", ss.StatCounters)
		stack.Loops[etime.Trial].OnEnd.Add("TrialStats", ss.TrialStats)
	}

	// note: phase is shared between all stacks!
	plusPhase, _ := man.Stacks[etime.Train].Loops[etime.Cycle].EventByName("PlusPhase")
	plusPhase.OnEvent.InsertBefore("PlusPhase:Start", "TakeAction", func() {
		// note: critical to have this happen *after* MinusPhase:End and *before* PlusPhase:Start
		// because minus phase end has gated info, and plus phase start applies action input
		ss.TakeAction(ss.Net)
	})

	man.GetLoop(etime.Train, etime.Run).OnStart.Add("NewRun", ss.NewRun)

	// Add Testing
	// trainEpoch := man.GetLoop(etime.Train, etime.Epoch)
	// trainEpoch.OnStart.Add("TestAtInterval", func() {
	// 	if (ss.TestInterval > 0) && ((trainEpoch.Counter.Cur+1)%ss.TestInterval == 0) {
	// 		// Note the +1 so that it doesn't occur at the 0th timestep.
	// 		ss.TestAll()
	// 	}
	// })

	/////////////////////////////////////////////
	// Logging

	man.GetLoop(etime.Train, etime.Epoch).OnEnd.Add("PCAStats", func() {
		trnEpc := man.Stacks[etime.Train].Loops[etime.Epoch].Counter.Cur
		if (ss.Sim.PCAInterval > 0) && (trnEpc%ss.Sim.PCAInterval == 0) {
			// if ss.Args.Bool("mpi") {
			// 	ss.Logs.MPIGatherTableRows(etime.Analyze, etime.Trial, ss.Comm)
			// }
			axon.PCAStats(ss.Net, &ss.Logs, &ss.Stats)
			ss.Logs.ResetLog(etime.Analyze, etime.Trial)
		}
	})

	man.AddOnEndToAll("Log", ss.Log)
	axon.LooperResetLogBelow(man, &ss.Logs, etime.Sequence) // exclude sequence, doesn't reset Trial
	man.GetLoop(etime.Train, etime.Epoch).OnStart.Add("ResetLogTrial", func() {
		ss.Logs.ResetLog(etime.Train, etime.Trial)
	})
	man.GetLoop(etime.Train, etime.Sequence).OnStart.Add("ResetLogTrial", func() {
		ss.Logs.ResetLog(etime.Debug, etime.Trial)
	})
	// man.GetLoop(etime.Test, etime.Epoch).OnStart.Add("ResetLogTrial", func() {
	// 	ss.Logs.ResetLog(etime.Train, etime.Trial)
	// })

	man.GetLoop(etime.Train, etime.Trial).OnEnd.Add("LogAnalyze", func() {
		trnEpc := man.Stacks[etime.Train].Loops[etime.Epoch].Counter.Cur
		if (ss.Sim.PCAInterval > 0) && (trnEpc%ss.Sim.PCAInterval == 0) {
			ss.Log(etime.Analyze, etime.Trial)
		}
	})

	// Save weights to file, to look at later
	man.GetLoop(etime.Train, etime.Run).OnEnd.Add("SaveWeights", func() {
		ctrString := ss.Stats.PrintVals([]string{"Run", "Epoch"}, []string{"%03d", "%05d"}, "_")
		axon.SaveWeightsIfArgSet(ss.Net, &ss.Args, ctrString, ss.Stats.String("RunName"))
	})

	man.GetLoop(etime.Train, etime.Epoch).OnEnd.Add("PctCortex", func() {
		trnEpc := ss.Loops.Stacks[etime.Train].Loops[etime.Epoch].Counter.Cur
		if trnEpc >= ss.Sim.PctCortexStEpc && trnEpc%ss.Sim.PctCortexInterval == 0 {
			ss.Sim.PctCortex = ss.Sim.PctCortexMax * float32(trnEpc-ss.Sim.PctCortexStEpc) / float32(ss.Sim.PctCortexNEpc)
			if ss.Sim.PctCortex > ss.Sim.PctCortexMax {
				ss.Sim.PctCortex = ss.Sim.PctCortexMax
			} else {
				mpi.Printf("PctCortex updated to: %g at epoch: %d\n", ss.Sim.PctCortex, trnEpc)
			}
		}
	})

	////////////////////////////////////////////
	// GUI

	if ss.Args.Bool("nogui") {
		// man.GetLoop(etime.Test, etime.Trial).Main.Add("NetDataRecord", func() {
		// 	ss.GUI.NetDataRecord(ss.ViewUpdt.Text)
		// })
	} else {
		axon.LooperUpdtNetView(man, &ss.ViewUpdt, ss.Net)
		axon.LooperUpdtPlots(man, &ss.GUI)
	}

	if Debug {
		mpi.Println(man.DocString())
	}
	ss.Loops = man
}

// TakeAction takes action for this step, using either decoded cortical
// or reflexive subcortical action from env.
func (ss *Sim) TakeAction(net *axon.Network) {
	ev := ss.Envs[ss.Context.Mode.String()].(*Approach)
	ev.ActGen() // always update comparison
	mtxLy := net.AxonLayerByName("VpMtxGo")
	didGate := mtxLy.AnyGated()
	if didGate && !ss.Context.PVLV.HasPosUS() {
		ev.DidGate = true
		ss.Stats.SetString("Debug", "skip gate")
		if Debug {
			fmt.Printf("skipped action because gated\n")
		}
		ev.Action("None", nil)
		ss.ApplyAction()
		ly := ss.Net.AxonLayerByName("VL")
		ly.Pools[0].Inhib.Clamped.SetBool(false) // not clamped this trial
		ss.Net.GPU.SyncPoolsToGPU()
		ss.Stats.SetFloat("ActMatch", 1) // whatever it is, it is ok
		return                           // no time to do action while also gating
	}

	netAct, anm := ss.DecodeAct(ev)
	genAct := ev.ActGen()
	genActNm := ev.Acts[genAct]
	ss.Stats.SetString("NetAction", anm)
	ss.Stats.SetString("InstinctAction", genActNm)
	if netAct == genAct {
		ss.Stats.SetFloat("ActMatch", 1)
	} else {
		ss.Stats.SetFloat("ActMatch", 0)
	}

	actAct := genAct
	if ss.Sim.CortexDriving {
		actAct = netAct
	}
	actActNm := ev.Acts[actAct]
	ss.Stats.SetString("ActAction", actActNm)

	ev.Action(actActNm, nil)
	ss.ApplyAction()
	// fmt.Printf("action: %s\n", ev.Acts[act])
}

// DecodeAct decodes the VL ActM state to find closest action pattern
func (ss *Sim) DecodeAct(ev *Approach) (int, string) {
	vt := ss.Stats.SetLayerTensor(ss.Net, "VL", "CaSpkP") // was "Act"
	return ev.DecodeAct(vt)
}

func (ss *Sim) ApplyAction() {
	net := ss.Net
	ev := ss.Envs[ss.Context.Mode.String()]
	ap := ev.State("Action")
	ly := net.AxonLayerByName("Act")
	ly.ApplyExt(ap)
	ss.Net.ApplyExts(&ss.Context)
}

// ApplyInputs applies input patterns from given environment.
// It is good practice to have this be a separate method with appropriate
// args so that it can be used for various different contexts
// (training, testing, etc).
func (ss *Sim) ApplyInputs() {
	ss.Stats.SetString("Debug", "") // start clear
	net := ss.Net
	ev := ss.Envs[ss.Context.Mode.String()].(*Approach)

	trl := ss.Loops.GetLoop(ss.Context.Mode, etime.Trial)
	if trl.Counter.Cur == 0 {
		ss.Sim.CortexDriving = erand.BoolP32(ss.Sim.PctCortex, -1)
	}

	ss.Net.InitExt() // clear any existing inputs -- not strictly necessary if always
	// going to the same layers, but good practice and cheap anyway

	lays := []string{"Pos", "CS", "Dist"}
	for _, lnm := range lays {
		ly := net.AxonLayerByName(lnm)
		itsr := ev.State(lnm)
		ly.ApplyExt(itsr)
	}

	// this is key step to drive DA and US-ACh
	if ev.US != -1 {
		ss.Context.NeuroMod.SetRew(ev.Rew, true)
	} else {
		ss.Context.NeuroMod.SetRew(0, false)
	}

	ss.ApplyPVLV(&ss.Context, ev)
	ss.Net.ApplyExts(&ss.Context)
}

// ApplyPVLV applies current PVLV values to Context.mDrivePVLV,
// from given trial data.
func (ss *Sim) ApplyPVLV(ctx *axon.Context, ev *Approach) {
	dr := &ctx.PVLV
	dr.InitUS()
	ctx.NeuroMod.HasRew.SetBool(false)
	if ev.US != -1 {
		dr.SetPosUS(int32(ev.US), 1) // magnitude always 1
		ctx.NeuroMod.HasRew.SetBool(true)
	}
	if ss.Context.PVLV.VSMatrix.JustGated.IsTrue() {
		ss.Context.PVLV.Effort.Reset() // always start counting at start of goal
	}
	dr.Effort.AddEffort(1) // should be based on action taken last step
	dr.InitDrives()
	dr.SetDrive(int32(ev.Drive), 1)
}

// NewRun intializes a new run of the model, using the TrainEnv.Run counter
// for the new run value
func (ss *Sim) NewRun() {
	ss.InitRndSeed()
	// ss.Envs.ByMode(etime.Train).Init(0)
	// ss.Envs.ByMode(etime.Test).Init(0)
	ss.Context.Reset()
	ss.Context.Mode = etime.Train
	ss.Sim.PctCortex = 0
	ss.InitWts(ss.Net)
	ss.InitStats()
	ss.StatCounters()
	ss.Logs.ResetLog(etime.Train, etime.Epoch)
	// ss.Logs.ResetLog(etime.Test, etime.Epoch)
}

// TestAll runs through the full set of testing items
// func (ss *Sim) TestAll() {
// 	// ss.Envs.ByMode(etime.Test).Init(0)
// 	ss.Loops.ResetAndRun(etime.Test)
// 	ss.Loops.Mode = etime.Train // Important to reset Mode back to Train because this is called from within the Train Run.
// }

////////////////////////////////////////////////////////////////////////////////////////////
// 		Stats

// InitStats initializes all the statistics.
// called at start of new run
func (ss *Sim) InitStats() {
	ss.Stats.SetFloat("PctCortex", 0)
	ss.Stats.SetFloat("Dist", 0)
	ss.Stats.SetFloat("Drive", 0)
	ss.Stats.SetFloat("CS", 0)
	ss.Stats.SetFloat("US", 0)
	ss.Stats.SetFloat("HasRew", 0)
	ss.Stats.SetString("NetAction", "")
	ss.Stats.SetString("InstinctAction", "")
	ss.Stats.SetString("ActAction", "")
	ss.Stats.SetFloat("JustGated", 0)
	ss.Stats.SetFloat("Should", 0)
	ss.Stats.SetFloat("GateUS", 0)
	ss.Stats.SetFloat("GateCS", 0)
	ss.Stats.SetFloat("GatedEarly", 0)
	ss.Stats.SetFloat("MaintEarly", 0)
	ss.Stats.SetFloat("GatedAgain", 0)
	ss.Stats.SetFloat("WrongCSGate", 0)
	ss.Stats.SetFloat("AChShould", 0)
	ss.Stats.SetFloat("AChShouldnt", 0)
	ss.Stats.SetFloat("Rew", 0)
	ss.Stats.SetFloat("DA", 0)
	ss.Stats.SetFloat("RewPred", 0)
	ss.Stats.SetFloat("DA_NR", 0)
	ss.Stats.SetFloat("RewPred_NR", 0)
	ss.Stats.SetFloat("ActMatch", 0)
	ss.Stats.SetFloat("AllGood", 0)
	lays := ss.Net.LayersByType(axon.PTMaintLayer)
	for _, lnm := range lays {
		ss.Stats.SetFloat("Maint"+lnm, 0)
		ss.Stats.SetFloat("MaintFail"+lnm, 0)
		ss.Stats.SetFloat("PreAct"+lnm, 0)
	}
	ss.Stats.SetString("Debug", "") // special debug notes per trial
}

// StatCounters saves current counters to Stats, so they are available for logging etc
// Also saves a string rep of them for ViewUpdt.Text
func (ss *Sim) StatCounters() {
	mode := ss.Context.Mode
	ss.Loops.Stacks[mode].CtrsToStats(&ss.Stats)
	// always use training epoch..
	trnEpc := ss.Loops.Stacks[etime.Train].Loops[etime.Epoch].Counter.Cur
	ev := ss.Envs[ss.Context.Mode.String()].(*Approach)
	ss.Stats.SetInt("Epoch", trnEpc)
	ss.Stats.SetInt("Cycle", int(ss.Context.Cycle))
	ss.Stats.SetFloat32("PctCortex", ss.Sim.PctCortex)
	ss.Stats.SetFloat32("Dist", float32(ev.Dist))
	ss.Stats.SetFloat32("Drive", float32(ev.Drive))
	ss.Stats.SetFloat32("CS", float32(ev.CS))
	ss.Stats.SetFloat32("US", float32(ev.US))
	ss.Stats.SetFloat32("HasRew", float32(ss.Context.NeuroMod.HasRew))
	ss.Stats.SetString("TrialName", "trl")
	ss.ViewUpdt.Text = ss.Stats.Print([]string{"Run", "Epoch", "Sequence", "Trial", "Cycle", "NetAction", "InstinctAction", "ActAction", "ActMatch", "JustGated", "Should", "Rew"})
}

// TrialStats computes the trial-level statistics.
// Aggregation is done directly from log data.
func (ss *Sim) TrialStats() {
	ctx := &ss.Context
	dr := &ctx.PVLV
	dr.DriveEffortUpdt(1, ctx.NeuroMod.HasRew.IsTrue(), false)

	ss.GatedStats()
	ss.MaintStats()

	if ss.Context.PVLV.HasPosUS() {
		ss.Stats.SetFloat32("DA", ss.Context.NeuroMod.DA)
		ss.Stats.SetFloat32("RewPred", ss.Context.NeuroMod.RewPred) // gets from VSPatch or RWPred etc
		ss.Stats.SetFloat("DA_NR", math.NaN())
		ss.Stats.SetFloat("RewPred_NR", math.NaN())
	} else {
		ss.Stats.SetFloat32("DA_NR", ss.Context.NeuroMod.DA)
		ss.Stats.SetFloat32("RewPred_NR", ss.Context.NeuroMod.RewPred)
		ss.Stats.SetFloat("DA", math.NaN())
		ss.Stats.SetFloat("RewPred", math.NaN())
	}

	ss.Stats.SetFloat32("ACh", ss.Context.NeuroMod.ACh)
	ss.Stats.SetFloat32("AChRaw", ss.Context.NeuroMod.AChRaw)

	var allGood float64
	agN := 0
	if fv := ss.Stats.Float("GateUS"); !math.IsNaN(fv) {
		allGood += fv
		agN++
	}
	if fv := ss.Stats.Float("GateCS"); !math.IsNaN(fv) {
		allGood += fv
		agN++
	}
	if fv := ss.Stats.Float("ActMatch"); !math.IsNaN(fv) {
		allGood += fv
		agN++
	}
	if fv := ss.Stats.Float("GatedEarly"); !math.IsNaN(fv) {
		allGood += 1 - fv
		agN++
	}
	if fv := ss.Stats.Float("GatedAgain"); !math.IsNaN(fv) {
		allGood += 1 - fv
		agN++
	}
	if fv := ss.Stats.Float("WrongCSGate"); !math.IsNaN(fv) {
		allGood += 1 - fv
		agN++
	}
	if agN > 0 {
		allGood /= float64(agN)
	}
	ss.Stats.SetFloat("AllGood", allGood)

	if ss.Context.PVLV.HasPosUS() { // got an outcome -- skip to next Sequence
		trl := ss.Loops.GetLoop(ss.Context.Mode, etime.Trial)
		trl.SkipToMax()
	}
}

// GatedStats updates the gated states
func (ss *Sim) GatedStats() {
	ev := ss.Envs[ss.Context.Mode.String()].(*Approach)
	justGated := ss.Context.PVLV.VSMatrix.JustGated.IsTrue()
	hasGated := ss.Context.PVLV.VSMatrix.HasGated.IsTrue()
	ss.Stats.SetFloat32("JustGated", bools.ToFloat32(justGated))
	ss.Stats.SetFloat32("Should", bools.ToFloat32(ev.ShouldGate))
	ss.Stats.SetFloat32("HasGated", bools.ToFloat32(hasGated))
	ss.Stats.SetFloat32("GateUS", mat32.NaN())
	ss.Stats.SetFloat32("GateCS", mat32.NaN())
	ss.Stats.SetFloat32("GatedEarly", mat32.NaN())
	ss.Stats.SetFloat32("MaintEarly", mat32.NaN())
	ss.Stats.SetFloat32("GatedAgain", mat32.NaN())
	ss.Stats.SetFloat32("WrongCSGate", mat32.NaN())
	ss.Stats.SetFloat32("AChShould", mat32.NaN())
	ss.Stats.SetFloat32("AChShouldnt", mat32.NaN())
	if justGated {
		ss.Stats.SetFloat32("WrongCSGate", bools.ToFloat32(ev.Drive != ev.USForPos()))
	}
	if ev.ShouldGate {
		if ss.Context.PVLV.HasPosUS() {
			ss.Stats.SetFloat32("GateUS", bools.ToFloat32(justGated))
		} else {
			ss.Stats.SetFloat32("GateCS", bools.ToFloat32(justGated))
		}
	} else {
		if hasGated {
			ss.Stats.SetFloat32("GatedAgain", bools.ToFloat32(justGated))
		} else { // !should gate means early..
			ss.Stats.SetFloat32("GatedEarly", bools.ToFloat32(justGated))
		}
	}
	// We get get ACh when new CS or Rew
	if ss.Context.PVLV.HasPosUS() || ev.LastCS != ev.CS {
		ss.Stats.SetFloat32("AChShould", ss.Context.NeuroMod.ACh)
	} else {
		ss.Stats.SetFloat32("AChShouldnt", ss.Context.NeuroMod.ACh)
	}

	ss.Stats.SetFloat32("Rew", ev.Rew)
}

// MaintStats updates the PFC maint stats
func (ss *Sim) MaintStats() {
	ev := ss.Envs[ss.Context.Mode.String()].(*Approach)
	// should be maintaining while going forward
	isFwd := ev.LastAct == ev.ActMap["Forward"]
	isCons := ev.LastAct == ev.ActMap["Consume"]
	actThr := float32(0.05) // 0.1 too high
	net := ss.Net
	lays := net.LayersByType(axon.PTMaintLayer)
	hasMaint := false
	for _, lnm := range lays {
		mnm := "Maint" + lnm
		fnm := "MaintFail" + lnm
		pnm := "PreAct" + lnm
		ptly := net.AxonLayerByName(lnm)
		var mact float32
		if ptly.Is4D() {
			for pi := 1; pi < len(ptly.Pools); pi++ {
				avg := ptly.Pools[pi].AvgMax.Act.Plus.Avg
				if avg > mact {
					mact = avg
				}
			}
		} else {
			mact = ptly.Pools[0].AvgMax.Act.Plus.Avg
		}
		overThr := mact > actThr
		if overThr {
			hasMaint = true
		}
		ss.Stats.SetFloat32(pnm, mat32.NaN())
		ss.Stats.SetFloat32(mnm, mat32.NaN())
		ss.Stats.SetFloat32(fnm, mat32.NaN())
		if isFwd {
			ss.Stats.SetFloat32(mnm, mact)
			ss.Stats.SetFloat32(fnm, bools.ToFloat32(!overThr))
		} else if !isCons {
			ss.Stats.SetFloat32(pnm, bools.ToFloat32(overThr))
		}
	}
	if hasMaint {
		ss.Stats.SetFloat32("MaintEarly", bools.ToFloat32(ev.Drive != ev.USForPos()))
	}
}

//////////////////////////////////////////////////////////////////////////////
// 		Logging

func (ss *Sim) ConfigLogs() {
	ss.Stats.SetString("RunName", ss.Params.RunName(0)) // used for naming logs, stats, etc

	ss.Logs.AddCounterItems(etime.Run, etime.Epoch, etime.Sequence, etime.Trial, etime.Cycle)
	ss.Logs.AddStatStringItem(etime.AllModes, etime.AllTimes, "RunName")
	// ss.Logs.AddStatStringItem(etime.AllModes, etime.Trial, "TrialName")
	ss.Logs.AddStatFloatNoAggItem(etime.AllModes, etime.AllTimes, "PctCortex")
	ss.Logs.AddStatFloatNoAggItem(etime.AllModes, etime.Trial, "Drive", "CS", "Dist", "US", "HasRew")
	ss.Logs.AddStatStringItem(etime.AllModes, etime.Trial, "NetAction", "InstinctAction", "ActAction")

	ss.Logs.AddPerTrlMSec("PerTrlMSec", etime.Run, etime.Epoch, etime.Trial)

	ss.ConfigLogItems()

	axon.LogAddPulvCorSimItems(&ss.Logs, ss.Net, etime.Train, etime.Run, etime.Epoch, etime.Trial)

	// ss.ConfigActRFs()

	layers := ss.Net.LayersByType(axon.SuperLayer, axon.CTLayer, axon.TargetLayer)
	axon.LogAddDiagnosticItems(&ss.Logs, layers, etime.Train, etime.Epoch, etime.Trial)
	axon.LogInputLayer(&ss.Logs, ss.Net, etime.Train)

	// todo: PCA items should apply to CT layers too -- pass a type here.
	axon.LogAddPCAItems(&ss.Logs, ss.Net, etime.Train, etime.Run, etime.Epoch, etime.Trial)

	// axon.LogAddLayerGeActAvgItems(&ss.Logs, ss.Net, etime.Test, etime.Cycle)

	ss.Logs.PlotItems("AllGood", "ActMatch", "GateCS", "GateUS", "WrongCSGate", "DA", "RewPred", "RewPred_NR", "LeftCor")
	// "MaintofcPT", "MaintaccPT", "MaintFailofcPT", "MaintFailaccPT"
	// "GateUS", "GatedEarly", "GatedAgain", "JustGated", "PctCortex",
	// "Rew", "DA", "MtxGo_ActAvg"

	ss.Logs.CreateTables()
	ss.Logs.SetContext(&ss.Stats, ss.Net)
	// don't plot certain combinations we don't use
	// ss.Logs.NoPlot(etime.Train, etime.Cycle)
	ss.Logs.NoPlot(etime.Train, etime.Phase, etime.Sequence)
	ss.Logs.NoPlot(etime.Test, etime.Run, etime.Epoch, etime.Sequence, etime.Trial)
	// note: Analyze not plotted by default
	ss.Logs.SetMeta(etime.Train, etime.Run, "LegendCol", "RunName")
	// ss.Logs.SetMeta(etime.Test, etime.Cycle, "LegendCol", "RunName")

	axon.LayerActsLogConfig(ss.Net, &ss.Logs)
}

func (ss *Sim) ConfigLogItems() {
	ss.Logs.AddStatAggItem("AllGood", "AllGood", etime.Run, etime.Epoch, etime.Trial)
	ss.Logs.AddStatAggItem("ActMatch", "ActMatch", etime.Run, etime.Epoch, etime.Trial)
	ss.Logs.AddStatAggItem("JustGated", "JustGated", etime.Run, etime.Epoch, etime.Trial)
	ss.Logs.AddStatAggItem("Should", "Should", etime.Run, etime.Epoch, etime.Trial)
	ss.Logs.AddStatAggItem("GateUS", "GateUS", etime.Run, etime.Epoch, etime.Trial)
	ss.Logs.AddStatAggItem("GateCS", "GateCS", etime.Run, etime.Epoch, etime.Trial)
	ss.Logs.AddStatAggItem("GatedEarly", "GatedEarly", etime.Run, etime.Epoch, etime.Trial)
	ss.Logs.AddStatAggItem("MaintEarly", "MaintEarly", etime.Run, etime.Epoch, etime.Trial)
	ss.Logs.AddStatAggItem("GatedAgain", "GatedAgain", etime.Run, etime.Epoch, etime.Trial)
	ss.Logs.AddStatAggItem("WrongCSGate", "WrongCSGate", etime.Run, etime.Epoch, etime.Trial)
	ss.Logs.AddStatAggItem("AChShould", "AChShould", etime.Run, etime.Epoch, etime.Trial)
	ss.Logs.AddStatAggItem("AChShouldnt", "AChShouldnt", etime.Run, etime.Epoch, etime.Trial)

	// Add a special debug message -- use of etime.Debug triggers
	// inclusion
	ss.Logs.AddStatStringItem(etime.Debug, etime.Trial, "Debug")

	lays := ss.Net.LayersByType(axon.PTMaintLayer)
	for _, lnm := range lays {
		nm := "Maint" + lnm
		ss.Logs.AddStatAggItem(nm, nm, etime.Run, etime.Epoch, etime.Trial)
		nm = "MaintFail" + lnm
		ss.Logs.AddStatAggItem(nm, nm, etime.Run, etime.Epoch, etime.Trial)
		nm = "PreAct" + lnm
		ss.Logs.AddStatAggItem(nm, nm, etime.Run, etime.Epoch, etime.Trial)
	}
	li := ss.Logs.AddStatAggItem("Rew", "Rew", etime.Run, etime.Epoch, etime.Trial)
	li.FixMin = false
	li = ss.Logs.AddStatAggItem("DA", "DA", etime.Run, etime.Epoch, etime.Trial)
	li.FixMin = false
	li = ss.Logs.AddStatAggItem("ACh", "ACh", etime.Run, etime.Epoch, etime.Trial)
	li.FixMin = false
	li = ss.Logs.AddStatAggItem("AChRaw", "AChRaw", etime.Run, etime.Epoch, etime.Trial)
	li.FixMin = false
	li = ss.Logs.AddStatAggItem("RewPred", "RewPred", etime.Run, etime.Epoch, etime.Trial)
	li.FixMin = false
	li = ss.Logs.AddStatAggItem("DA_NR", "DA_NR", etime.Run, etime.Epoch, etime.Trial)
	li.FixMin = false
	li = ss.Logs.AddStatAggItem("RewPred_NR", "RewPred_NR", etime.Run, etime.Epoch, etime.Trial)
	li.FixMin = false

	ev := TheSim.Envs[etime.Train.String()].(*Approach)
	ss.Logs.AddItem(&elog.Item{
		Name:      "ActCor",
		Type:      etensor.FLOAT64,
		CellShape: []int{ev.NActs},
		DimNames:  []string{"Acts"},
		// Plot:      true,
		Range:     minmax.F64{Min: 0},
		TensorIdx: -1, // plot all values
		Write: elog.WriteMap{
			etime.Scope(etime.Train, etime.Epoch): func(ctx *elog.Context) {
				ix := ctx.Logs.IdxView(ctx.Mode, etime.Trial)
				spl := split.GroupBy(ix, []string{"InstinctAction"})
				split.AggTry(spl, "ActMatch", agg.AggMean)
				ags := spl.AggsToTable(etable.ColNameOnly)
				ss.Logs.MiscTables["ActCor"] = ags
				ctx.SetTensor(ags.Cols[0]) // cors
			}}})
	for _, nm := range ev.Acts { // per-action % correct
		anm := nm // closure
		ss.Logs.AddItem(&elog.Item{
			Name: anm + "Cor",
			Type: etensor.FLOAT64,
			// Plot:  true,
			Range: minmax.F64{Min: 0},
			Write: elog.WriteMap{
				etime.Scope(etime.Train, etime.Epoch): func(ctx *elog.Context) {
					ags := ss.Logs.MiscTables["ActCor"]
					rw := ags.RowsByString("InstinctAction", anm, etable.Equals, etable.UseCase)
					if len(rw) > 0 {
						ctx.SetFloat64(ags.CellFloat("ActMatch", rw[0]))
					}
				}}})
	}
}

// Log is the main logging function, handles special things for different scopes
func (ss *Sim) Log(mode etime.Modes, time etime.Times) {
	if mode != etime.Analyze && mode != etime.Debug {
		ss.Context.Mode = mode // Also set specifically in a Loop callback.
	}
	ss.StatCounters()

	dt := ss.Logs.Table(mode, time)
	if dt == nil {
		return
	}
	row := dt.Rows

	switch {
	case time == etime.Cycle:
		row = ss.Stats.Int("Cycle")
	case mode == etime.Train && time == etime.Trial:
		// if !ss.Context.PVLV.HasPosUS() && ss.Stats.Float("JustGated") == 1 {
		// 	axon.LayerActsLog(ss.Net, &ss.Logs, &ss.GUI)
		// }
		if !ss.Context.PVLV.HasPosUS() && ss.Context.PVLV.VSMatrix.HasGated.IsTrue() { // maint
			axon.LayerActsLog(ss.Net, &ss.Logs, &ss.GUI)
		}
		ss.Logs.Log(etime.Debug, etime.Trial)
		if !ss.Args.Bool("nogui") {
			ss.GUI.UpdateTableView(etime.Debug, etime.Trial)
		}

		// if ss.Stats.Float("GatedEarly") > 0 {
		// 	fmt.Printf("STOPPED due to gated early: %d  %g\n", ev.US, ev.Rew)
		// 	ss.Loops.Stop(etime.Trial)
		// }

		// trnEpc := ss.Loops.Stacks[etime.Train].Loops[etime.Epoch].Counter.Cur
		// if ss.StopOnErr && trnEpc > 5 && ev.Rew < 0 {
		// 	fmt.Printf("STOPPED due to negative reward for US: %d  %g\n", ev.US, ev.Rew)
		// 	ss.Loops.Stop(etime.Trial)
		// }
	case mode == etime.Train && time == etime.Epoch:
		axon.LayerActsLogAvg(ss.Net, &ss.Logs, &ss.GUI, true) // reset recs
	}

	ss.Logs.LogRow(mode, time, row) // also logs to file, etc
}

////////////////////////////////////////////////////////////////////////////////////////////
// 		Gui

// ConfigGui configures the GoGi gui interface for this simulation,
func (ss *Sim) ConfigGui() *gi.Window {
	title := "BOA = BG, OFC ACC Test"
	ss.GUI.MakeWindow(ss, "boa", title, `This project tests learning in the BG, OFC & ACC for basic approach learning to a CS associated with a US. See <a href="https://github.com/emer/axon">axon on GitHub</a>.</p>`)
	ss.GUI.CycleUpdateInterval = 20

	nv := ss.GUI.AddNetView("NetView")
	nv.Params.MaxRecs = 300
	nv.Params.LayNmSize = 0.02
	nv.SetNet(ss.Net)
	ss.ViewUpdt.Config(nv, etime.AlphaCycle, etime.AlphaCycle)

	nv.Scene().Camera.Pose.Pos.Set(0, 1.4, 2.6)
	nv.Scene().Camera.LookAt(mat32.Vec3{0, 0, 0}, mat32.Vec3{0, 1, 0})

	ss.GUI.ViewUpdt = &ss.ViewUpdt

	ss.GUI.AddPlots(title, &ss.Logs)

	ss.GUI.AddTableView(&ss.Logs, etime.Debug, etime.Trial)

	axon.LayerActsLogConfigGUI(&ss.Logs, &ss.GUI)

	ss.GUI.AddToolbarItem(egui.ToolbarItem{Label: "Init", Icon: "update",
		Tooltip: "Initialize everything including network weights, and start over.  Also applies current params.",
		Active:  egui.ActiveStopped,
		Func: func() {
			ss.Init()
			ss.GUI.UpdateWindow()
		},
	})

	ss.GUI.AddLooperCtrl(ss.Loops, []etime.Modes{etime.Train})

	////////////////////////////////////////////////
	ss.GUI.ToolBar.AddSeparator("log")
	ss.GUI.AddToolbarItem(egui.ToolbarItem{Label: "Reset RunLog",
		Icon:    "reset",
		Tooltip: "Reset the accumulated log of all Runs, which are tagged with the ParamSet used",
		Active:  egui.ActiveAlways,
		Func: func() {
			ss.Logs.ResetLog(etime.Train, etime.Run)
			ss.GUI.UpdatePlot(etime.Train, etime.Run)
		},
	})
	////////////////////////////////////////////////
	ss.GUI.ToolBar.AddSeparator("misc")
	ss.GUI.AddToolbarItem(egui.ToolbarItem{Label: "New Seed",
		Icon:    "new",
		Tooltip: "Generate a new initial random seed to get different results.  By default, Init re-establishes the same initial seed every time.",
		Active:  egui.ActiveAlways,
		Func: func() {
			ss.RndSeeds.NewSeeds()
		},
	})
	ss.GUI.AddToolbarItem(egui.ToolbarItem{Label: "README",
		Icon:    "file-markdown",
		Tooltip: "Opens your browser on the README file that contains instructions for how to run this model.",
		Active:  egui.ActiveAlways,
		Func: func() {
			gi.OpenURL("https://github.com/emer/axon/blob/master/examples/boa/README.md")
		},
	})
	ss.GUI.FinalizeGUI(false)
	if GPU {
		ss.Net.ConfigGPUwithGUI(&TheSim.Context)
		gi.SetQuitCleanFunc(func() {
			ss.Net.GPU.Destroy()
		})
	}
	return ss.GUI.Win
}

func (ss *Sim) ConfigArgs() {
	ss.Args.Init()
	ss.Args.AddStd()
	ss.Args.SetInt("epochs", 200)
	ss.Args.SetInt("runs", 10)
	ss.Args.AddInt("seqs", 25, "sequences per epoch")
	ss.Args.Parse() // always parse
}

func (ss *Sim) CmdArgs() {
	ss.Args.ProcStd(&ss.Params)
	ss.Args.ProcStdLogs(&ss.Logs, &ss.Params, ss.Net.Name())
	ss.Args.SetBool("nogui", true)                                       // by definition if here
	ss.Stats.SetString("RunName", ss.Params.RunName(ss.Args.Int("run"))) // used for naming logs, stats, etc

	netdata := ss.Args.Bool("netdata")
	if netdata {
		mpi.Printf("Saving NetView data from testing\n")
		ss.GUI.InitNetData(ss.Net, 200)
	}

	runs := ss.Args.Int("runs")
	run := ss.Args.Int("run")
	mpi.Printf("Running %d Runs starting at %d\n", runs, run)
	rc := &ss.Loops.GetLoop(etime.Train, etime.Run).Counter
	rc.Set(run)
	rc.Max = run + runs

	ss.Loops.GetLoop(etime.Train, etime.Epoch).Counter.Max = ss.Args.Int("epochs")

	ss.Loops.GetLoop(etime.Train, etime.Sequence).Counter.Max = ss.Args.Int("seqs")

	ss.NewRun()
	ss.Loops.Run(etime.Train)

	ss.Logs.CloseLogFiles()

	if netdata {
		ss.GUI.SaveNetData(ss.Stats.String("RunName"))
	}
}
