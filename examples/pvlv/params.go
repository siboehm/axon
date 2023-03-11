// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "github.com/emer/emergent/params"

// ParamSets is the default set of parameters -- Base is always applied,
// and others can be optionally selected to apply on top of that
var ParamSets = params.Sets{
	{Name: "Base", Desc: "minimal base params needed for this model", Sheets: params.Sheets{
		"Network": &params.Sheet{
			{Sel: "Layer", Desc: "generic params for all layers: lower gain, slower, soft clamp",
				Params: params.Params{
					// "Layer.Act.Decay.Act":   "0.0", // do this only where needed
					// "Layer.Act.Decay.Glong": "0.0",
					"Layer.Act.Clamp.Ge": "1.5",
				}},
			{Sel: ".CTLayer", Desc: "corticothalamic context -- using FSA-based params -- intermediate",
				Params: params.Params{
					"Layer.Inhib.ActAvg.Nominal": "0.12",
					"Layer.CT.GeGain":            "1.0",
					"Layer.CT.DecayTau":          "50",
					"Layer.Inhib.Layer.Gi":       "2.2",
					"Layer.Inhib.Pool.Gi":        "2.2",
					"Layer.Act.GABAB.Gbar":       "0.25",
					"Layer.Act.NMDA.Gbar":        "0.25",
					"Layer.Act.NMDA.Tau":         "200",
					"Layer.Act.Decay.Act":        "0.0",
					"Layer.Act.Decay.Glong":      "0.0",
					"Layer.Act.Sahp.Gbar":        "1.0",
				}},
			{Sel: ".PTPredLayer", Desc: "PTPred prediction layer -- more dynamic acts",
				Params: params.Params{
					"Layer.Inhib.ActAvg.Nominal": "0.12",
					"Layer.CT.GeGain":            "0.01",
					"Layer.CT.DecayTau":          "50",
					"Layer.Inhib.Layer.Gi":       "0.8", // overridden by OFCPTPred below
					"Layer.Inhib.Pool.Gi":        "0.8",
					"Layer.Act.GABAB.Gbar":       "0.2", // regular
					"Layer.Act.NMDA.Gbar":        "0.15",
					"Layer.Act.NMDA.Tau":         "100",
					"Layer.Act.Decay.Act":        "0.2",
					"Layer.Act.Decay.Glong":      "0.6",
					"Layer.Act.Sahp.Gbar":        "0.1",
					"Layer.Act.KNa.Slow.Max":     "0.2", // maybe too random if higher?
				}},
			{Sel: ".PTMaintLayer", Desc: "time integration params",
				Params: params.Params{
					"Layer.Inhib.Layer.Gi":   "1.8", // was 1.0
					"Layer.Inhib.Pool.Gi":    "1.8", // was 1.8
					"Layer.Act.GABAB.Gbar":   "0.3",
					"Layer.Act.NMDA.Gbar":    "0.3", // 0.3 enough..
					"Layer.Act.NMDA.Tau":     "300",
					"Layer.Act.Decay.Act":    "0.0",
					"Layer.Act.Decay.Glong":  "0.0",
					"Layer.Act.Sahp.Gbar":    "0.01", // not much pressure -- long maint
					"Layer.Act.Dend.ModGain": "20",   // 10?
				}},
			{Sel: ".VThalLayer", Desc: "",
				Params: params.Params{
					"Layer.Inhib.Layer.Gi": "0.8",
					"Layer.Inhib.Pool.Gi":  "0.8", // 0.6 > 0.5 -- 0.8 too high
				}},
			{Sel: "#Drives", Desc: "expect act",
				Params: params.Params{
					"Layer.Inhib.ActAvg.Nominal": "0.1", // 1 / ndrives
				}},
			{Sel: ".USLayer", Desc: "expect act",
				Params: params.Params{
					"Layer.Inhib.ActAvg.Nominal": "0.25", // 1 / ndrives
				}},
			{Sel: ".PVLayer", Desc: "expect act",
				Params: params.Params{
					"Layer.Inhib.ActAvg.Nominal": "0.25", // 1 / ndrives
				}},
			{Sel: ".InputLayer", Desc: "t",
				Params: params.Params{
					"Layer.Act.Decay.Act":   "1.0",
					"Layer.Act.Decay.Glong": "1.0",
				}},
			{Sel: "#StimIn", Desc: "expect act",
				Params: params.Params{
					"Layer.Inhib.ActAvg.Nominal": "0.1", // 1 / css
				}},
			{Sel: ".OFC", Desc: "",
				Params: params.Params{
					"Layer.Act.Decay.Act":        "0.0", // do this only where needed
					"Layer.Act.Decay.Glong":      "0.0",
					"Layer.Act.Decay.OnRew":      "true", // everything clears
					"Layer.Inhib.ActAvg.Nominal": "0.025",
					"Layer.Inhib.Layer.Gi":       "1.1",
					"Layer.Inhib.Pool.On":        "true",
					"Layer.Inhib.Pool.Gi":        "0.8",
					"Layer.Act.Dend.SSGi":        "0",
				}},
			{Sel: "#OFCCT", Desc: "",
				Params: params.Params{
					"Layer.Inhib.Layer.Gi": "2.8", // 2.4 not strong enough to prevent diffuse activity
					"Layer.Inhib.Pool.Gi":  "1.2", // was 1.4
				}},
			{Sel: "#OFCPTPred", Desc: "",
				Params: params.Params{
					"Layer.Inhib.Layer.Gi": "0.8",
					"Layer.Inhib.Pool.Gi":  "0.8",
				}},
			{Sel: "#OFC", Desc: "",
				Params: params.Params{
					"Layer.Inhib.Layer.Gi": "1.2", // was 1.1
				}},
			{Sel: "#OFCPT", Desc: "",
				Params: params.Params{
					"Layer.Inhib.Layer.Gi": "1.8", // was 1.3
					"Layer.Inhib.Pool.Gi":  "2.0", // was 0.6
				}},
			{Sel: "#OFCMD", Desc: "",
				Params: params.Params{
					"Layer.Inhib.Pool.Gi": "0.8",
				}},
			{Sel: "#ACh", Desc: "",
				Params: params.Params{
					"Layer.RSalACh.RewThr": "0.2",
				}},
			{Sel: ".BLALayer", Desc: "",
				Params: params.Params{
					"Layer.Act.Decay.Act":            "1.0",
					"Layer.Act.Decay.Glong":          "1.0",
					"Layer.Inhib.ActAvg.Nominal":     "0.025",
					"Layer.Inhib.Layer.Gi":           "1.8", // needs to be strong to prevent random off-US act
					"Layer.Inhib.Pool.On":            "true",
					"Layer.Inhib.Pool.Gi":            "0.9",
					"Layer.Act.Gbar.L":               "0.2",
					"Layer.Learn.NeuroMod.BurstGain": "0.2",
					"Layer.Learn.NeuroMod.DipGain":   "0",   // ignore small negative DA
					"Layer.BLA.NegLRate":             "0.1", // todo: explore
					"Layer.Learn.RLRate.SigmoidMin":  "1.0",
					"Layer.Learn.RLRate.Diff":        "true", // can turn off if NoDALRate is 0
					"Layer.Learn.RLRate.DiffThr":     "0.01", // based on cur - prv
				}},
			{Sel: "#BLAPosExtD2", Desc: "",
				Params: params.Params{
					"Layer.Act.Gbar.L":                 "0.2",
					"Layer.Inhib.Pool.Gi":              "0.4",
					"Layer.Learn.NeuroMod.BurstGain":   "0",
					"Layer.Learn.NeuroMod.DipGain":     "1",
					"Layer.Learn.NeuroMod.AChLRateMod": "0",
					"Layer.Learn.RLRate.Diff":          "false", // can turn off if NoDALRate is 0
				}},
			{Sel: "#CeMPos", Desc: "",
				Params: params.Params{
					"Layer.Inhib.ActAvg.Nominal": "0.15",
					"Layer.Act.Dend.SSGi":        "0",
					"Layer.Inhib.Layer.Gi":       "0.5",
					"Layer.Inhib.Pool.On":        "true",
					"Layer.Inhib.Pool.Gi":        "0.3",
				}},
			{Sel: ".PPTgLayer", Desc: "",
				Params: params.Params{
					"Layer.Inhib.ActAvg.Nominal": "0.1",
					"Layer.Inhib.Layer.Gi":       "1.0", // todo: explore
					"Layer.Inhib.Pool.On":        "true",
					"Layer.Inhib.Pool.Gi":        "0.5", // todo: could be lower = more bursting
					"Layer.Inhib.Pool.FFPrv":     "10",  // key td param
					"Layer.PVLV.Gain":            "4",   // key for impact on CS bursting
					"Layer.PVLV.Thr":             "0.2",
				}},
			{Sel: ".DrivesLayer", Desc: "",
				Params: params.Params{
					"Layer.Inhib.ActAvg.Nominal": "0.1",
					"Layer.Inhib.Layer.On":       "false",
					"Layer.Inhib.Pool.On":        "true",
					"Layer.Inhib.Pool.Gi":        "0.5",
					"Layer.Pulv.DriveScale":      "0.05",
				}},
			{Sel: ".USLayer", Desc: "",
				Params: params.Params{
					"Layer.Inhib.ActAvg.Nominal": "0.1",
					"Layer.Inhib.Layer.On":       "false",
					"Layer.Inhib.Pool.On":        "true",
					"Layer.Inhib.Pool.Gi":        "0.5",
				}},
			{Sel: ".VSPatchLayer", Desc: "",
				Params: params.Params{
					"Layer.Act.Gbar.L":           "0.4",
					"Layer.Inhib.ActAvg.Nominal": "0.2",
					"Layer.Inhib.Layer.On":       "false",
					"Layer.Inhib.Layer.Gi":       "1.0",
					"Layer.Inhib.Pool.On":        "true",
					"Layer.Inhib.Pool.Gi":        "0.3",
					"Layer.PVLV.Gain":            "6",
					"Layer.PVLV.Thr":             "0.2",
				}},
			{Sel: "#VpSTNp", Desc: "Pausing STN",
				Params: params.Params{
					"Layer.Inhib.ActAvg.Nominal": "0.15",
					"Layer.Inhib.Layer.On":       "true", // this is critical, else too active
					"Layer.Inhib.Layer.Gi":       "0.6",
				}},
			{Sel: "#VpSTNs", Desc: "Sustained STN",
				Params: params.Params{
					"Layer.Act.Init.GeBase":      "0.2",
					"Layer.Act.Init.GeVar":       "0.2",
					"Layer.Inhib.ActAvg.Nominal": "0.15",
					"Layer.Inhib.Layer.On":       "true",
					"Layer.Inhib.Layer.Gi":       "0.2",
				}},
			{Sel: ".GPLayer", Desc: "all gp",
				Params: params.Params{
					"Layer.Act.Init.GeBase":      "0.3",
					"Layer.Act.Init.GeVar":       "0.1",
					"Layer.Act.Init.GiVar":       "0.1",
					"Layer.Inhib.ActAvg.Nominal": "1",
				}},
			{Sel: "#VpGPi", Desc: "",
				Params: params.Params{
					"Layer.Act.Init.GeBase": "0.6",
				}},
			{Sel: ".MatrixLayer", Desc: "all mtx",
				Params: params.Params{
					"Layer.Matrix.GateThr":             "0.05", // 0.05 > 0.08 maybe
					"Layer.Matrix.NoGoGeLrn":           "0.1",  // 0.1 >= 0.2 > 0.5 a bit
					"Layer.Learn.NeuroMod.AChDisInhib": "5",    // key to be 5
					"Layer.Act.Dend.ModGain":           "2",
					"Layer.Inhib.ActAvg.Nominal":       ".03",
					"Layer.Inhib.Layer.On":             "true",
					"Layer.Inhib.Layer.Gi":             "0.0", // was .8
					"Layer.Inhib.Pool.On":              "true",
					"Layer.Inhib.Pool.Gi":              "0.5", // 0.7 > 0.6 more sparse
				}},
			// {Sel: "#SNc", Desc: "SNc -- no clamp limits",
			// 	Params: params.Params{
			// 	}},
			// {Sel: ".RWPredLayer", Desc: "",
			// 	Params: params.Params{
			// 		"Layer.RWPred.PredRange.Min": "0.01",
			// 		"Layer.RWPred.PredRange.Max": "0.99",
			// 	}},
			////////////////////////////////////////////////////////////////
			// cortical prjns
			{Sel: "Prjn", Desc: "all prjns",
				Params: params.Params{
					"Prjn.Learn.Trace.Tau":  "4",
					"Prjn.Learn.LRate.Base": "0.04",
				}},
			{Sel: ".BackPrjn", Desc: "back is weaker",
				Params: params.Params{
					"Prjn.PrjnScale.Rel": "0.1",
				}},
			{Sel: ".SuperToPT", Desc: "one-to-one from super -- just use fixed nonlearning prjn so can control behavior easily",
				Params: params.Params{
					"Prjn.PrjnScale.Rel": "1",    // keep this constant -- only self vs. this -- thal is modulatory
					"Prjn.PrjnScale.Abs": "0.01", // monitor maint early and other maint stats with PTMaintLayer ModGain = 0 to set this so super alone is not able to drive it.
					"Prjn.Learn.Learn":   "false",
					"Prjn.SWt.Init.Mean": "0.8",
					"Prjn.SWt.Init.Var":  "0.0",
				}},
			{Sel: ".PTSelfMaint", Desc: "",
				Params: params.Params{
					"Prjn.PrjnScale.Rel":    "1",      // use abs to manipulate
					"Prjn.PrjnScale.Abs":    "2",      // 2 > 1
					"Prjn.Learn.LRate.Base": "0.0001", // slower > faster
					"Prjn.SWt.Init.Mean":    "0.5",
					"Prjn.SWt.Init.Var":     "0.5", // high variance so not just spreading out over time
				}},
			{Sel: ".SuperToThal", Desc: "",
				Params: params.Params{
					"Prjn.PrjnScale.Rel": "1.0",
					"Prjn.PrjnScale.Abs": "2.0", // if this is too strong, it gates to the wrong CS
					"Prjn.Learn.Learn":   "false",
					"Prjn.SWt.Init.Mean": "0.8",
					"Prjn.SWt.Init.Var":  "0.0",
				}},
			{Sel: ".ThalToSuper", Desc: "",
				Params: params.Params{
					"Prjn.PrjnScale.Rel": "0.1",
				}},
			{Sel: ".ThalToPT", Desc: "",
				Params: params.Params{
					"Prjn.PrjnScale.Rel": "1.0",
					"Prjn.Com.GType":     "ModulatoryG", // this marks as modulatory with extra ModGain factor
					"Prjn.Learn.Learn":   "false",
					"Prjn.SWt.Init.Mean": "0.8",
					"Prjn.SWt.Init.Var":  "0.0",
				}},
			{Sel: ".CTtoThal", Desc: "",
				Params: params.Params{
					"Prjn.SWt.Init.Var":  "0.25",
					"Prjn.SWt.Init.Mean": "0.5",
					"Prjn.PrjnScale.Rel": "0.1",
				}},
			{Sel: ".CTCtxtPrjn", Desc: "all CT context prjns",
				Params: params.Params{
					"Prjn.Learn.LRate.Base":    "0.01", // trace: .01 > .005 > .02; .03 > .02 > .01 -- .03 std
					"Prjn.Learn.Trace.Tau":     "4",    // 2 > 1
					"Prjn.Learn.Trace.SubMean": "0",    // 0 > 1 -- 1 is especially bad
				}},
			//////////////////////////////////////////////
			// To BLA
			{Sel: ".BLAPrjn", Desc: "",
				Params: params.Params{
					"Prjn.Learn.Trace.Tau": "1",
				}},
			{Sel: ".VSPatchPrjn", Desc: "",
				Params: params.Params{
					"Prjn.Learn.LRate.Base": "0.0", // 0.1 def
				}},
			{Sel: ".USToBLA", Desc: "starts strong, learns slow",
				Params: params.Params{
					"Prjn.SWt.Init.SPct":    "0",
					"Prjn.SWt.Init.Mean":    "0.5",
					"Prjn.SWt.Init.Var":     "0.25",
					"Prjn.Learn.LRate.Base": "0.001",
					"Prjn.PrjnScale.Rel":    "0.5",
				}},
			{Sel: "#USposToBLAPosAcqD1", Desc: "",
				Params: params.Params{
					"Prjn.PrjnScale.Abs": "5.0",
				}},
			{Sel: "#StimInToBLAPosAcqD1", Desc: "",
				Params: params.Params{
					"Prjn.Learn.LRate.Base": "0.01",
					"Prjn.PrjnScale.Abs":    "1",
				}},
			/*
				{Sel: "#OFCToBLAPosExtD2", Desc: "",
					Params: params.Params{
						"Prjn.SWt.Init.Mean": "0.5",
						"Prjn.SWt.Init.Var":  "0.25",
					}},
			*/
			{Sel: "#BLAPosAcqD1ToOFC", Desc: "strong",
				Params: params.Params{
					"Prjn.PrjnScale.Abs": "4",
				}},
			{Sel: "#BLAPosExtD2ToBLAPosAcqD1", Desc: "inhibition from extinction",
				Params: params.Params{
					"Prjn.PrjnScale.Abs": "1",
					"Prjn.SWt.Init.Mean": "0.5",
					"Prjn.SWt.Init.Var":  "0.0",
				}},
			{Sel: ".OFCToBLAExt", Desc: "weak",
				Params: params.Params{
					"Prjn.Learn.LRate.Base": "0.2",
					"Prjn.PrjnScale.Abs":    "1",
					"Prjn.SWt.Init.SPct":    "0",
					"Prjn.SWt.Init.Mean":    "0.1",
					"Prjn.SWt.Init.Var":     "0.05",
				}},
			{Sel: ".BLAToCeM_Excite", Desc: "",
				Params: params.Params{
					"Prjn.Learn.Learn":   "false",
					"Prjn.PrjnScale.Abs": "1",
					"Prjn.SWt.Init.SPct": "0",
					"Prjn.SWt.Adapt.On":  "false",
					"Prjn.SWt.Init.Mean": "0.8",
					"Prjn.SWt.Init.Var":  "0.0",
				}},
			{Sel: ".BLAToCeM_Inhib", Desc: "",
				Params: params.Params{
					"Prjn.Learn.Learn":   "false",
					"Prjn.PrjnScale.Abs": "1",
					"Prjn.SWt.Init.SPct": "0",
					"Prjn.SWt.Init.Mean": "0.8",
					"Prjn.SWt.Init.Var":  "0.0",
				}},
			{Sel: ".BLAExtToAcq", Desc: "fixed inhibitory",
				Params: params.Params{
					"Prjn.PrjnScale.Abs": "0.01",
					"Prjn.SWt.Init.SPct": "0",
					"Prjn.SWt.Init.Mean": "0.8",
					"Prjn.SWt.Init.Var":  "0.0",
					"Prjn.Learn.Learn":   "false",
				}},
			{Sel: ".CeMToPPTg", Desc: "",
				Params: params.Params{
					"Prjn.Learn.Learn":   "false",
					"Prjn.PrjnScale.Abs": "1",
					"Prjn.SWt.Init.SPct": "0",
					"Prjn.SWt.Init.Mean": "0.8",
					"Prjn.SWt.Init.Var":  "0.0",
				}},
			{Sel: "#TimePToOFCPTPred", Desc: "needs to be strong so reps are differentiated",
				Params: params.Params{
					"Prjn.PrjnScale.Rel": "1",
				}},

			// BG prjns
			{Sel: ".MatrixPrjn", Desc: "",
				Params: params.Params{
					"Prjn.PrjnScale.Abs":      "1.0", // stronger
					"Prjn.SWt.Init.SPct":      "0",
					"Prjn.SWt.Init.Mean":      "0.5",
					"Prjn.SWt.Init.Var":       "0.25",
					"Prjn.Matrix.NoGateLRate": "0.005", // 0.005 std -- seems ok..
					"Prjn.Matrix.CurTrlDA":    "true",
					"Prjn.Matrix.UseHasRew":   "true", // hack to use US-only timing
					"Prjn.Matrix.AChDecay":    "0",    // not used if UseHasRew is on
					"Prjn.Learn.Learn":        "true",
					"Prjn.Learn.LRate.Base":   "0.1",
				}},
			{Sel: ".BgFixed", Desc: "fixed, non-learning params",
				Params: params.Params{
					"Prjn.SWt.Init.SPct": "0",
					"Prjn.SWt.Init.Mean": "0.8",
					"Prjn.SWt.Init.Var":  "0.0",
					"Prjn.Learn.Learn":   "false",
				}},
			{Sel: "#USposToVpMtxGo", Desc: "",
				Params: params.Params{
					"Prjn.PrjnScale.Abs": "5",
					"Prjn.PrjnScale.Rel": ".2",
				}},
			{Sel: ".BLAToBG", Desc: "",
				Params: params.Params{
					"Prjn.PrjnScale.Rel": "8",
				}},
			{Sel: ".DrivesToMtx", Desc: "",
				Params: params.Params{
					"Prjn.Learn.Learn":   "false",
					"Prjn.PrjnScale.Rel": "1", // even .1 does gating without CS
					"Prjn.Com.GType":     "ModulatoryG",
				}},
			{Sel: ".DrivesToOFC", Desc: "",
				Params: params.Params{
					"Prjn.PrjnScale.Abs": "2", // 2 > 1
					"Prjn.PrjnScale.Rel": ".5",
					// "Prjn.Learn.Learn":   "false",
					// "Prjn.SWt.Init.Mean": "0.8",
					// "Prjn.SWt.Init.Var":  "0.0",
				}},
			{Sel: ".FmSTNp", Desc: "increase to prevent repeated gating",
				Params: params.Params{
					"Prjn.PrjnScale.Abs": "1.2", // 1.2 > 1.0 > 1.5 (too high)
				}},
			{Sel: ".GPeTAToMtx", Desc: "nonspecific gating activity surround inhibition -- wta",
				Params: params.Params{
					"Prjn.PrjnScale.Abs": "2", // 2 nominal
				}},
			{Sel: ".GPeInToMtx", Desc: "provides weak counterbalance for GPeTA -> Mtx to reduce oscillations",
				Params: params.Params{
					"Prjn.PrjnScale.Abs": "0.5",
				}},
		}},
	},
}
