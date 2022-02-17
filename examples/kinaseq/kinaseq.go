// Copyright (c) 2020, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// kinaseq plots kinase learning simulation over time
package main

import (
	"math/rand"
	"strconv"
	"strings"

	"github.com/emer/etable/eplot"
	"github.com/emer/etable/etable"
	"github.com/emer/etable/etensor"
	_ "github.com/emer/etable/etview" // include to get gui views
	"github.com/goki/gi/gi"
	"github.com/goki/gi/gimain"
	"github.com/goki/gi/giv"
	"github.com/goki/ki/ki"
	"github.com/goki/ki/kit"
	"github.com/goki/mat32"
)

func main() {
	TheSim.Config()
	gimain.Main(func() { // this starts gui -- requires valid OpenGL display connection (e.g., X11)
		guirun()
	})
}

func guirun() {
	win := TheSim.ConfigGui()
	win.StartEventLoop()
}

// LogPrec is precision for saving float values in logs
const LogPrec = 4

// KinaseRules are different options for Kinase-based learning rules
type KinaseRules int32

//go:generate stringer -type=KinaseRules

var KiT_KinaseRules = kit.Enums.AddEnum(KinaseRulesN, kit.NotBitFlag, nil)

func (ev KinaseRules) MarshalJSON() ([]byte, error)  { return kit.EnumMarshalJSON(ev) }
func (ev *KinaseRules) UnmarshalJSON(b []byte) error { return kit.EnumUnmarshalJSON(ev, b) }

// The time scales
const (
	// NeurSpkCa uses neuron-level spike-driven calcium signals
	// integrated at P vs. D time scales -- this is the original
	// Leabra and Axon XCAL / CHL learning rule.
	NeurSpkCa KinaseRules = iota

	// SynSpkCaOR uses synapse-level spike-driven calcium signals
	// with an OR rule for pre OR post spiking driving the CaM up,
	// which is then integrated at P vs. D time scales.
	// Basically a synapse version of original learning rule.
	SynSpkCaOR

	// SynSpkNMDAOR uses synapse-level spike-driven calcium signals
	// with an OR rule for pre OR post spiking driving the CaM up,
	// with NMDAo multiplying the spike drive to fit Bio Ca better
	// including the Bonus factor.
	// which is then integrated at P vs. D time scales.
	SynSpkNMDAOR

	// SynNMDACa uses synapse-level NMDA-driven calcium signals
	// (which can be either Urakubo allosteric or Kinase abstract)
	// integrated at P vs. D time scales -- abstract version
	// of the KinaseB biophysical learniung rule
	SynNMDACa

	KinaseRulesN
)

// KinaseSynParams has rate constants for averaging over activations
// at different time scales, to produce the running average activation
// values that then drive learning.
type KinaseSynParams struct {
	Rule   KinaseRules     `desc:"which learning rule to use"`
	OptP   float32         `desc:"optimized p inc gain"`
	SpikeG float32         `def:"20,8,200" desc:"spiking gain for Spk rules"`
	MTau   float32         `def:"10,40" min:"1" desc:"CaM mean running-average time constant in cycles, which should be milliseconds typically (tau is roughly how long it takes for value to change significantly -- 1.4x the half-life). This provides a pre-integration step before integrating into the CaP short time scale"`
	PTau   float32         `def:"40,10" min:"1" desc:"LTP Ca-driven factor time constant in cycles, which should be milliseconds typically (tau is roughly how long it takes for value to change significantly -- 1.4x the half-life). Continuously updates based on current CaI value, resulting in faster tracking of plus-phase signals."`
	DTau   float32         `def:"40" min:"1" desc:"LTD Ca-driven factor time constant in cycles, which should be milliseconds typically (tau is roughly how long it takes for value to change significantly -- 1.4x the half-life).  Continuously updates based on current CaP value, resulting in slower integration that still reflects earlier minus-phase signals."`
	DScale float32         `def:"0.93,1.05" desc:"scaling factor on CaD as it enters into the learning rule, to compensate for systematic decrease in activity over the course of a theta cycle"`
	PFun   etensor.Float32 `view:"no-inline" desc:"P function table"`
	DFun   etensor.Float32 `view:"no-inline" desc:"D function table"`
	Tmax   int             `desc:"maximum time in lookup tables"`
	Ymax   float32         `desc:"maximum y value in lookup tables"`
	Yres   float32         `desc:"resolution of Y value in lookup tables"`

	MDt float32 `view:"-" json:"-" xml:"-" inactive:"+" desc:"rate = 1 / tau"`
	PDt float32 `view:"-" json:"-" xml:"-" inactive:"+" desc:"rate = 1 / tau"`
	DDt float32 `view:"-" json:"-" xml:"-" inactive:"+" desc:"rate = 1 / tau"`
}

func (kp *KinaseSynParams) Defaults() {
	kp.Rule = SynSpkCaOR
	kp.OptP = 100
	kp.SpikeG = 8 // 200 // 8
	kp.MTau = 10  // 40
	kp.PTau = 40  // 10
	kp.DTau = 40
	kp.DScale = 1.05
	kp.Tmax = 100
	kp.Ymax = 2
	kp.Yres = 0.01
	kp.Update()
}

func (kp *KinaseSynParams) Update() {
	kp.MDt = 1 / kp.MTau
	kp.PDt = 1 / kp.PTau
	kp.DDt = 1 / kp.DTau
	kp.FillFun()
}

// FmSpike computes updates from current spike value
func (kp *KinaseSynParams) FmSpike(spike float32, caM, caP, caD *float32) {
	*caM += kp.MDt * (kp.SpikeG*spike - *caM)
	*caP += kp.PDt * (*caM - *caP)
	*caD += kp.DDt * (*caP - *caD)
}

// DWt computes the weight change from CaP, CaD values
func (kp *KinaseSynParams) DWt(caP, caD float32) float32 {
	return caP - kp.DScale*caD
}

// FillFun fill in the functions
func (kp *KinaseSynParams) FillFun() {
	yn := int(kp.Ymax/kp.Yres) + 1
	kp.PFun.SetShape([]int{yn, yn, kp.Tmax + 1}, nil, []string{"ps", "ms", "time"})
	for pi := 0; pi < yn; pi++ {
		for mi := 0; mi < yn; mi++ {
			pv := float32(pi) * kp.Yres
			mv := float32(mi) * kp.Yres
			m := mv
			ps := pv
			for t := 0; t <= kp.Tmax; t++ {
				m -= kp.MDt * m
				ps += kp.PDt * (m - ps)
				kp.PFun.Set([]int{pi, mi, t}, ps)
			}
		}
	}
}

// PVal returns P value for given initial P value, m value, and time point
func (kp *KinaseSynParams) PVal(pv, mv float32, t int) float32 {
	if pv > kp.Ymax {
		pv = kp.Ymax
	}
	if pv < 0 {
		pv = 0
	}
	if mv > kp.Ymax {
		mv = kp.Ymax
	}
	if mv < 0 {
		mv = 0
	}
	if t < 0 {
		t = 0
	}
	if t > kp.Tmax {
		t = kp.Tmax
	}
	pi := int(pv / kp.Yres)
	mi := int(mv / kp.Yres)
	yv := kp.PFun.Value([]int{pi, mi, t})
	// if pi < kp.PFun.Dim(0)-1 {
	// 	pr := (pv - (float32(pi) * kp.Yres)) / kp.Yres
	// 	yvh := kp.PFun.Value([]int{pi + 1, mi, t})
	// 	yv += pr * (yvh - yv)
	// }
	// if mi < kp.PFun.Dim(0)-1 {
	// 	mr := (mv - (float32(mi) * kp.Yres)) / kp.Yres
	// 	yvh := kp.PFun.Value([]int{pi, mi + 1, t})
	// 	yv += mr * (yvh - yv)
	// }
	return yv
}

///////////////////////////////////////////////

// Sim holds the params, table, etc
type Sim struct {
	Kinase    KinaseSynParams `desc:Kinase rate constants"`
	PGain     float32         `desc:"multiplier on product factor to equate to SynC"`
	SpikeDisp float32         `desc:"spike multiplier for display purposes"`
	NReps     int             `desc:"number of repetitions -- if > 1 then only final @ end of Dur shown"`
	DurMsec   int             `desc:"duration for activity window"`
	SendHz    float32         `desc:"sending firing frequency (used as minus phase for ThetaErr)"`
	RecvHz    float32         `desc:"receiving firing frequency (used as plus phase for ThetaErr)"`
	Table     *etable.Table   `view:"no-inline" desc:"table for plot"`
	Plot      *eplot.Plot2D   `view:"-" desc:"the plot"`
	Win       *gi.Window      `view:"-" desc:"main GUI window"`
	ToolBar   *gi.ToolBar     `view:"-" desc:"the master toolbar"`
}

// TheSim is the overall state for this simulation
var TheSim Sim

// Config configures all the elements using the standard functions
func (ss *Sim) Config() {
	ss.Kinase.Defaults()
	ss.PGain = 10
	ss.SpikeDisp = 0.1
	ss.NReps = 1
	ss.DurMsec = 200
	ss.SendHz = 20
	ss.RecvHz = 20
	ss.Update()
	ss.Table = &etable.Table{}
	ss.ConfigTable(ss.Table)
}

// Update updates computed values
func (ss *Sim) Update() {
	ss.Kinase.Update()
}

// Run runs the equation.
func (ss *Sim) Run() {
	ss.Update()
	dt := ss.Table

	if ss.NReps == 1 {
		dt.SetNumRows(ss.DurMsec)
	} else {
		dt.SetNumRows(ss.NReps)
	}

	Sint := mat32.Exp(-1000.0 / float32(ss.SendHz))
	Rint := mat32.Exp(-1000.0 / float32(ss.RecvHz))
	Sp := float32(1)
	Rp := float32(1)

	kp := &ss.Kinase

	var rSpk, rSpkCaM, rSpkCaP, rSpkCaD float32                   // recv
	var sSpk, sSpkCaM, sSpkCaP, sSpkCaD float32                   // send
	var pSpkCaM, pSpkCaP, pSpkCaD, pDWt float32                   // product
	var cSpk, cSpkCaM, cSpkCaP, cSpkCaD, cDWt, cISI float32       // syn continuous
	var oSpk, oSpkCaM, oSpkCaP, oSpkCaD, oCaM, oCaP, oDWt float32 // syn optimized compute

	for nr := 0; nr < ss.NReps; nr++ {
		for t := 0; t < ss.DurMsec; t++ {
			row := t
			if ss.NReps == 1 {
				dt.SetCellFloat("Time", row, float64(row))
			} else {
				row = nr
				dt.SetCellFloat("Time", row, float64(row))
			}

			Sp *= rand.Float32()
			if Sp <= Sint {
				sSpk = 1
				Sp = 1
			} else {
				sSpk = 0
			}

			Rp *= rand.Float32()
			if Rp <= Rint {
				rSpk = 1
				Rp = 1
			} else {
				rSpk = 0
			}

			kp.FmSpike(rSpk, &rSpkCaM, &rSpkCaP, &rSpkCaD)
			kp.FmSpike(sSpk, &sSpkCaM, &sSpkCaP, &sSpkCaD)

			// this is standard CHL form
			pSpkCaM = ss.PGain * rSpkCaM * sSpkCaM
			pSpkCaP = ss.PGain * rSpkCaP * sSpkCaP
			pSpkCaD = ss.PGain * rSpkCaD * sSpkCaD

			pDWt = kp.DWt(pSpkCaP, pSpkCaD)

			// either side drives up..
			cSpk = 0
			switch kp.Rule {
			case SynSpkCaOR:
				if sSpk > 0 || rSpk > 0 {
					cSpk = 1
				}
			}
			kp.FmSpike(cSpk, &cSpkCaM, &cSpkCaP, &cSpkCaD)
			cDWt = kp.DWt(cSpkCaP, cSpkCaD)

			// optimized
			if cSpk > 0 {
				isi := int(cISI)
				mprv := float32(0)
				if isi > 0 {
					mprv = oSpkCaM * mat32.FastExp(-kp.MDt*cISI)
				}
				minc := kp.MDt * (kp.SpikeG*cSpk - mprv)
				oSpkCaM = mprv + minc

				if isi > 0 {
					oSpkCaP = kp.PVal(oSpkCaP, oSpkCaM, isi) // update
				}
				// oSpkCaP += kp.PVal(oSpkCaP, oSpkCaM, 0) // update
				oCaM = oSpkCaM
				oCaP = oSpkCaP
				cISI = 0
			} else {
				cISI += 1
				isi := int(cISI)

				oCaM = oSpkCaM * mat32.Exp(-kp.MDt*cISI)
				oCaP = kp.PVal(oSpkCaP, oSpkCaM, (isi - 1))
			}

			if ss.NReps == 1 || t == ss.DurMsec-1 {
				dt.SetCellFloat("RSpike", row, float64(ss.SpikeDisp*rSpk))
				dt.SetCellFloat("RSpkCaM", row, float64(rSpkCaM))
				dt.SetCellFloat("RSpkCaP", row, float64(rSpkCaP))
				dt.SetCellFloat("RSpkCaD", row, float64(rSpkCaD))
				dt.SetCellFloat("SSpike", row, float64(ss.SpikeDisp*sSpk))
				dt.SetCellFloat("SSpkCaM", row, float64(sSpkCaM))
				dt.SetCellFloat("SSpkCaP", row, float64(sSpkCaP))
				dt.SetCellFloat("SSpkCaD", row, float64(sSpkCaD))
				dt.SetCellFloat("SynPSpkCaM", row, float64(pSpkCaM))
				dt.SetCellFloat("SynPSpkCaP", row, float64(pSpkCaP))
				dt.SetCellFloat("SynPSpkCaD", row, float64(pSpkCaD))
				dt.SetCellFloat("SynPDWt", row, float64(pDWt))
				dt.SetCellFloat("SynCSpike", row, float64(ss.SpikeDisp*cSpk))
				dt.SetCellFloat("SynCSpkCaM", row, float64(cSpkCaM))
				dt.SetCellFloat("SynCSpkCaP", row, float64(cSpkCaP))
				dt.SetCellFloat("SynCSpkCaD", row, float64(cSpkCaD))
				dt.SetCellFloat("SynCDWt", row, float64(cDWt))
				dt.SetCellFloat("SynOSpike", row, float64(ss.SpikeDisp*oSpk))
				dt.SetCellFloat("SynOSpkCaM", row, float64(oCaM))
				dt.SetCellFloat("SynOSpkCaP", row, float64(oCaP))
				dt.SetCellFloat("SynOSpkCaD", row, float64(oSpkCaD))
				dt.SetCellFloat("SynODWt", row, float64(oDWt))
			}
		}
	}
	ss.Plot.Update()
}

func (ss *Sim) ConfigTable(dt *etable.Table) {
	dt.SetMetaData("name", "Kinase Opt Table")
	dt.SetMetaData("read-only", "true")
	dt.SetMetaData("precision", strconv.Itoa(LogPrec))

	sch := etable.Schema{
		{"Time", etensor.FLOAT64, nil, nil},
		{"RSpike", etensor.FLOAT64, nil, nil},
		{"RSpkCaM", etensor.FLOAT64, nil, nil},
		{"RSpkCaP", etensor.FLOAT64, nil, nil},
		{"RSpkCaD", etensor.FLOAT64, nil, nil},
		{"SSpike", etensor.FLOAT64, nil, nil},
		{"SSpkCaM", etensor.FLOAT64, nil, nil},
		{"SSpkCaP", etensor.FLOAT64, nil, nil},
		{"SSpkCaD", etensor.FLOAT64, nil, nil},
		{"SynPSpkCaM", etensor.FLOAT64, nil, nil},
		{"SynPSpkCaP", etensor.FLOAT64, nil, nil},
		{"SynPSpkCaD", etensor.FLOAT64, nil, nil},
		{"SynPDWt", etensor.FLOAT64, nil, nil},
		{"SynCSpike", etensor.FLOAT64, nil, nil},
		{"SynCSpkCaM", etensor.FLOAT64, nil, nil},
		{"SynCSpkCaP", etensor.FLOAT64, nil, nil},
		{"SynCSpkCaD", etensor.FLOAT64, nil, nil},
		{"SynCDWt", etensor.FLOAT64, nil, nil},
		{"SynOSpkCaM", etensor.FLOAT64, nil, nil},
		{"SynOSpkCaP", etensor.FLOAT64, nil, nil},
		{"SynOSpkCaD", etensor.FLOAT64, nil, nil},
		{"SynODWt", etensor.FLOAT64, nil, nil},
	}
	dt.SetFromSchema(sch, 0)
}

func (ss *Sim) ConfigPlot(plt *eplot.Plot2D, dt *etable.Table) *eplot.Plot2D {
	plt.Params.Title = "Kinase Learning Plot"
	plt.Params.XAxisCol = "Time"
	plt.SetTable(dt)

	for _, cn := range dt.ColNames {
		if cn == "Time" {
			continue
		}
		if strings.Contains(cn, "DWt") {
			plt.SetColParams(cn, eplot.Off, eplot.FloatMin, 0, eplot.FloatMax, 0)
		} else {
			plt.SetColParams(cn, eplot.Off, eplot.FixMin, 0, eplot.FloatMax, 0)
		}
	}
	// plt.SetColParams("SynCSpkCaM", eplot.On, eplot.FloatMin, 0, eplot.FloatMax, 0)
	// plt.SetColParams("SynOSpkCaM", eplot.On, eplot.FloatMin, 0, eplot.FloatMax, 0)
	plt.SetColParams("SynCSpkCaP", eplot.On, eplot.FloatMin, 0, eplot.FloatMax, 0)
	plt.SetColParams("SynOSpkCaP", eplot.On, eplot.FloatMin, 0, eplot.FloatMax, 0)
	return plt
}

// ConfigGui configures the GoGi gui interface for this simulation,
func (ss *Sim) ConfigGui() *gi.Window {
	width := 1600
	height := 1200

	// gi.WinEventTrace = true

	gi.SetAppName("kinaseq")
	gi.SetAppAbout(`Exploration of kinase equations. See <a href="https://github.com/emer/axon/blob/master/examples/kinaseq"> GitHub</a>.</p>`)

	win := gi.NewMainWindow("kinaseq", "Kinase Equation Exploration", width, height)
	ss.Win = win

	vp := win.WinViewport2D()
	updt := vp.UpdateStart()

	mfr := win.SetMainFrame()

	tbar := gi.AddNewToolBar(mfr, "tbar")
	tbar.SetStretchMaxWidth()
	ss.ToolBar = tbar

	split := gi.AddNewSplitView(mfr, "split")
	split.Dim = mat32.X
	split.SetStretchMax()

	sv := giv.AddNewStructView(split, "sv")
	sv.SetStruct(ss)

	tv := gi.AddNewTabView(split, "tv")

	plt := tv.AddNewTab(eplot.KiT_Plot2D, "Plot").(*eplot.Plot2D)
	ss.Plot = ss.ConfigPlot(plt, ss.Table)

	split.SetSplits(.3, .7)

	tbar.AddAction(gi.ActOpts{Label: "Run", Icon: "update", Tooltip: "Run the equations and plot results."}, win.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
		ss.Run()
		vp.SetNeedsFullRender()
	})

	tbar.AddAction(gi.ActOpts{Label: "README", Icon: "file-markdown", Tooltip: "Opens your browser on the README file that contains instructions for how to run this model."}, win.This(),
		func(recv, send ki.Ki, sig int64, data interface{}) {
			gi.OpenURL("https://github.com/emer/axon/blob/master/examples/kinaseq/README.md")
		})

	vp.UpdateEndNoSig(updt)

	// main menu
	appnm := gi.AppName()
	mmen := win.MainMenu
	mmen.ConfigMenus([]string{appnm, "File", "Edit", "Window"})

	amen := win.MainMenu.ChildByName(appnm, 0).(*gi.Action)
	amen.Menu.AddAppMenu(win)

	emen := win.MainMenu.ChildByName("Edit", 1).(*gi.Action)
	emen.Menu.AddCopyCutPaste(win)

	win.MainMenuUpdated()
	return win
}