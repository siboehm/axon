// Copyright (c) 2020, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"github.com/emer/emergent/params"
)

var ParamSets = params.Sets{
	{Name: "Base", Desc: "base params", Sheets: params.Sheets{
		"Network": &params.Sheet{
			{Sel: "Layer", Desc: "default Axon layer params",
				Params: params.Params{
					"Layer.Act.Init.Decay":     "0",
					"Layer.Act.Erev.K":         "0.1",
					"Layer.Act.KNa.On":         "false",
					"Layer.Inhib.Layer.On":     "true",
					"Layer.Inhib.Layer.Gi":     "1",
					"Layer.Inhib.Layer.FB":     "0",
					"Layer.Inhib.Pool.On":      "false",
					"Layer.Inhib.Pool.Gi":      "2",
					"Layer.Inhib.Pool.FB":      "0",
					"Layer.Inhib.Self.On":      "true",
					"Layer.Inhib.Self.Gi":      "0.3",
					"Layer.Inhib.ActAvg.Fixed": "true",
					"Layer.Learn.AvgL.Init":    "0.1",
				}},
			{Sel: "Prjn", Desc: "no extra learning factors",
				Params: params.Params{
					"Prjn.Learn.Norm.On":     "false",
					"Prjn.Learn.Momentum.On": "false",
					"Prjn.Learn.WtBal.On":    "false",
					"Prjn.WtInit.Sym":        "false",
				}},
			{Sel: ".PVLVLrnCons", Desc: "Base for most projections",
				Params: params.Params{
					"Prjn.Learn.Learn": "true",
					"Prjn.Learn.Lrate": "0.02",
					"Prjn.WtInit.Mean": "0.01",
					"Prjn.WtInit.Var":  "0",
					"Prjn.WtInit.Sym":  "false",
				}},
			{Sel: ".BLAmygConsStim", Desc: "StimIn to BLA",
				Params: params.Params{
					"Prjn.Learn.Learn":      "true",
					"Prjn.Learn.Lrate":      "0.2",
					"Prjn.Learn.WtSig.Gain": "1",
					"Prjn.WtInit.Dist":      "Gaussian",
					"Prjn.WtInit.Mean":      "0",
					"Prjn.WtInit.Var":       "1",
					"Prjn.WtInit.Sym":       "false",
					"Prjn.WtScale.Abs":      "0.8",
					"Prjn.SetScale":         "true",
					"Prjn.SetScaleMin":      "0.07",
					"Prjn.SetScaleMax":      "1",
					"Prjn.InitWtVal":        "0.07",
				}},
			{Sel: ".BLAmygConsUS", Desc: "US to BLAmyg",
				Params: params.Params{
					"Prjn.Learn.Learn": "false",
					"Prjn.WtScale.Rel": "0.08",
					"Prjn.WtInit.Mean": "0.5",
					"Prjn.WtInit.Var":  "0",
					"Prjn.WtInit.Sym":  "false",
				}},
			{Sel: ".BLAmygConsCntxtExt", Desc: "ContextIn to BLA projections",
				Params: params.Params{
					"Prjn.WtScale.Abs":      "0.75",
					"Prjn.Learn.Lrate":      "0.5",
					"Prjn.Learn.WtSig.Gain": "1",
					"Prjn.DALRBase":         "0",
					"Prjn.DALrnThr":         "0.01",
					"Prjn.ActDeltaThr":      "0",
					"Prjn.WtInit.Dist":      "Gaussian",
					"Prjn.WtInit.Mean":      "0",
					"Prjn.WtInit.Var":       "1",
					"Prjn.Learn.Learn":      "true",
					"Prjn.SetScale":         "true",
					"Prjn.SetScaleMin":      "0.1",
					"Prjn.SetScaleMax":      "1",
				}},
			{Sel: ".BLAmygConsInhib", Desc: "BLAmyg lateral inhibition",
				Params: params.Params{
					"Prjn.Learn.Learn": "false",
					"Prjn.WtScale.Abs": "1.2",
					"Prjn.WtInit.Mean": "0.9",
					"Prjn.WtInit.Var":  "0",
				}},
			{Sel: ".VSPatchConsToPosD1", Desc: "To VSPatchPosD1",
				Params: params.Params{
					"Prjn.WtInit.Mean": "0.02",
					"Prjn.WtInit.Var":  "0",
					"Prjn.Learn.Lrate": "0.015",
					//"Prjn.Learn.Lrate": "0.005",
				}},
			{Sel: ".VSPatchConsToPosD2", Desc: "To VSPatchPosD2",
				Params: params.Params{
					"Prjn.WtInit.Mean": "0.02",
					"Prjn.WtInit.Var":  "0",
					"Prjn.Learn.Lrate": "0.012",
				}},
			{Sel: ".VSPatchConsToNegD1", Desc: "To VSPatchNegD1",
				Params: params.Params{
					"Prjn.WtInit.Mean": "0.02",
					"Prjn.WtInit.Var":  "0",
					"Prjn.Learn.Lrate": "0.01",
				}},
			{Sel: ".VSPatchConsToNegD2", Desc: "To VSPatchNegD2",
				Params: params.Params{
					"Prjn.WtInit.Mean": "0.02",
					"Prjn.WtInit.Var":  "0",
					"Prjn.Learn.Lrate": "0.01",
				}},
			{Sel: ".VSMatrixConsToPosD1", Desc: "To VSMatrixPosD1",
				Params: params.Params{
					"Prjn.WtScale.Abs": "0.25",
					"Prjn.WtInit.Mean": "0.01",
					"Prjn.WtInit.Var":  "0",
					"Prjn.Learn.Lrate": "0.03",
					//"Prjn.Learn.Lrate": "0.02",
				}},
			{Sel: ".VSMatrixConsToPosD2", Desc: "To VSMatrixPosD2",
				Params: params.Params{
					"Prjn.WtScale.Abs": "0.25",
					"Prjn.WtInit.Mean": "0.01",
					"Prjn.WtInit.Var":  "0",
					"Prjn.Learn.Lrate": "0.06",
				}},
			{Sel: ".VSMatrixConsToNegD1", Desc: "To VSMatrixPosD1",
				Params: params.Params{
					"Prjn.WtScale.Abs": "0.25",
					"Prjn.WtInit.Mean": "0.01",
					"Prjn.WtInit.Var":  "0",
					"Prjn.Learn.Lrate": "0.06",
				}},
			{Sel: ".VSMatrixConsToNegD2", Desc: "To VSMatrixNegD2",
				Params: params.Params{
					"Prjn.WtScale.Abs": "0.25",
					"Prjn.WtInit.Mean": "0.01",
					"Prjn.WtInit.Var":  "0",
					"Prjn.Learn.Lrate": "0.03",
				}},
			{Sel: ".CElExtToAcqInhib", Desc: "CElAmyg lateral",
				Params: params.Params{
					"Prjn.Learn.Learn": "false",
					"Prjn.WtInit.Mean": "0.9",
					"Prjn.WtInit.Var":  "0",
					"Prjn.WtScale.Abs": "2.1",
				}},
			{Sel: ".CElAcqToExtInhib", Desc: "CElAmyg lateral",
				Params: params.Params{
					"Prjn.Learn.Learn": "false",
					"Prjn.WtInit.Mean": "0.9",
					"Prjn.WtInit.Var":  "0",
					"Prjn.WtInit.Sym":  "false",
					"Prjn.WtScale.Abs": "0.6",
				}},
			{Sel: ".CElAcqToExtDeepMod", Desc: "CElAmyg ModLevel",
				Params: params.Params{
					"Prjn.Learn.Learn": "false",
					"Prjn.WtInit.Mean": "0.9",
					"Prjn.WtInit.Var":  "0",
					"Prjn.WtInit.Sym":  "false",
				}},
			{Sel: ".StimToBLAmyg", Desc: "StimIn to BLAmyg",
				Params: params.Params{
					"Prjn.Learn.Lrate": "0.2",
				}},
			{Sel: ".PVToCElPVAct", Desc: "Clamp from PV to CEl",
				Params: params.Params{
					"Prjn.Learn.Learn": "false",
					"Prjn.WtInit.Mean": "1",
					"Prjn.WtInit.Var":  "0",
				}},
			{Sel: ".CEltoCeMFixed", Desc: "CEl to CEm",
				Params: params.Params{
					"Prjn.Learn.Learn": "false",
					"Prjn.WtScale.Abs": "1.3",
					"Prjn.WtInit.Mean": "0.9",
					"Prjn.WtInit.Var":  "0",
				}},
			{Sel: ".PVLVFixedCons", Desc: "Fixed projections",
				Params: params.Params{
					"Prjn.Learn.Learn": "false",
					"Prjn.WtInit.Mean": "0.9",
					"Prjn.WtInit.Var":  "0",
				}},
			{Sel: ".CtxtToBLAmyg", Desc: "ContextIn to BLAmyg",
				Params: params.Params{
					"Prjn.Learn.Lrate": "0.5",
				}},
			{Sel: ".StimExtToBLAmyg", Desc: "StimInExt to BLAmyg",
				Params: params.Params{
					"Prjn.Learn.Lrate": "0.5",
				}},
			{Sel: ".CElAmygCons", Desc: "CElAmyg",
				Params: params.Params{
					"Prjn.Learn.Lrate":      "0.2",
					"Prjn.WtScale.Abs":      "2.5",
					"Prjn.Learn.WtSig.Gain": "1",
					"Prjn.WtInit.Mean":      "0.02",
					"Prjn.WtInit.Var":       "0",
					"Prjn.WtInit.Sym":       "false",
					"Prjn.DALRBase":         "0",
					"Prjn.DALrnThr":         "0.01",
					"Prjn.ActDeltaThr":      "0",
				}},
			{Sel: ".CElAmygConsFmBLA", Desc: "CElAmyg",
				Params: params.Params{
					"Prjn.Learn.Lrate": "0.05",
					"Prjn.WtInit.Mean": "0.1",
					"Prjn.WtInit.Var":  "0",
					"Prjn.DALRBase":    "0",
					"Prjn.DALrnThr":    "0.01",
					"Prjn.ActDeltaThr": "0",
				}},
			{Sel: ".CElAmygConsExtFmBLA", Desc: "CElAmyg",
				Params: params.Params{
					"Prjn.Learn.Lrate": "0",
					"Prjn.WtInit.Mean": "0.9",
					"Prjn.WtInit.Var":  "0",
					"Prjn.DALrnThr":    "0.01",
					"Prjn.ActDeltaThr": "0",
				}},
			{Sel: ".BLAToCElA", Desc: "BL to CEl",
				Params: params.Params{
					"Prjn.Learn.Lrate": "0.05",
				}},
			{Sel: ".BLAToCElAExt", Desc: "BL to CEl ext",
				Params: params.Params{
					"Prjn.Learn.Lrate": "0.0",
				}},
			{Sel: ".BLAToVSDeepMod", Desc: "BLA to VS, also lateral within BLA",
				Params: params.Params{
					"Prjn.Learn.Learn": "false",
					"Prjn.WtInit.Mean": "0.9",
					"Prjn.WtInit.Var":  "0",
				}},
			{Sel: ".PVLVLayer", Desc: "Global layer params",
				Params: params.Params{
					"Layer.Inhib.ActAvg.Fixed":  "true",
					"Layer.Inhib.ActAvg.Adjust": "0",
					"Layer.Inhib.ActAvg.Init":   "0.25",
				}},
			{Sel: ".DALayer", Desc: "DA source layer",
				Params: params.Params{
					"Layer.Inhib.Self.On": "true",
				}},
			{Sel: ".AmygdalaLayer", Desc: "Amygdala layers",
				Params: params.Params{
					"Layer.Inhib.Layer.FB":      "1",
					"Layer.Inhib.Layer.FF0":     "0",
					"Layer.Inhib.Pool.On":       "false",
					"Layer.Inhib.Self.DaOn":     "true",
					"Layer.Inhib.Self.Gi":       "0.3",
					"Layer.Inhib.ActAvg.Fixed":  "true",
					"Layer.Inhib.ActAvg.Adjust": "0",
					"Layer.Inhib.ActAvg.Init":   "0.25",
					"Layer.Inhib.Self.FB":       "2.5",
				}},
			{Sel: ".CEmLayer", Desc: "both CEm layers",
				Params: params.Params{
					"Layer.Act.Dt.VmTau":       "5",
					"Layer.Act.Dt.GTau":        "3",
					"Layer.Act.XX1.Gain":       "200",
					"Layer.Inhib.Layer.FB":     "0",
					"Layer.Inhib.Layer.FBTau":  "3",
					"Layer.Inhib.ActAvg.Fixed": "true",
					"Layer.Inhib.Pool.On":      "false",
					"Layer.Inhib.Self.On":      "true",
					"Layer.Inhib.Self.Gi":      "2.05",
					"Layer.Inhib.Self.Tau":     "3.8",
				}},
			{Sel: "#VSPatchPosD1", Desc: "VSPatchPosD1 layer",
				Params: params.Params{
					"Layer.Act.XX1.Gain":    "40",
					"Layer.Act.Dt.VmTau":    "5",
					"Layer.Act.Dt.GTau":     "3",
					"Layer.ActModZero":      "false",
					"Layer.ModNetThreshold": "0.05",
					"Layer.DaMod.On":        "false",
					"Layer.DaMod.BurstGain": "1",
					"Layer.DaMod.DipGain":   "1",
					"Layer.Inhib.Layer.Gi":  "1.5",
					"Layer.Inhib.Layer.FB":  "0.1",
					"Layer.Inhib.Self.Gi":   "0.5",
					"Layer.Learn.AvgL.Init": "0.4",
				}},
			{Sel: "#VSPatchPosD2", Desc: "VSPatchPosD2 layer",
				Params: params.Params{
					"Layer.Act.XX1.Gain":    "40",
					"Layer.Act.Dt.VmTau":    "5",
					"Layer.Act.Dt.GTau":     "3",
					"Layer.ActModZero":      "true",
					"Layer.ModNetThreshold": "0.05",
					"Layer.DaMod.On":        "false",
					"Layer.DaMod.BurstGain": "3",
					"Layer.DaMod.DipGain":   "1",
					"Layer.Inhib.Layer.Gi":  "1.5",
					"Layer.Inhib.Layer.FB":  "0.1",
					"Layer.Inhib.Self.Gi":   "0.5",
					"Layer.Learn.AvgL.Init": "0.4",
				}},
			{Sel: "#VSPatchNegD1", Desc: "VSPatchNegD1 layer",
				Params: params.Params{
					"Layer.Act.XX1.Gain":    "40",
					"Layer.Act.Dt.VmTau":    "5",
					"Layer.Act.Dt.GTau":     "3",
					"Layer.ActModZero":      "true",
					"Layer.ModNetThreshold": "0.05",
					"Layer.DaMod.On":        "false",
					"Layer.DaMod.BurstGain": "0.5",
					"Layer.DaMod.DipGain":   "1",
					"Layer.Inhib.Layer.Gi":  "1.5",
					"Layer.Inhib.Layer.FB":  "0.1",
					"Layer.Inhib.Self.Gi":   "0.5",
					"Layer.Learn.AvgL.Init": "0.4",
				}},
			{Sel: "#VSPatchNegD2", Desc: "VSPatchNegD2 layer",
				Params: params.Params{
					"Layer.Act.XX1.Gain":    "40",
					"Layer.Act.Dt.VmTau":    "5",
					"Layer.Act.Dt.GTau":     "3",
					"Layer.ActModZero":      "false",
					"Layer.ModNetThreshold": "0.05",
					"Layer.DaMod.On":        "false",
					"Layer.DaMod.BurstGain": "1",
					"Layer.DaMod.DipGain":   "1",
					"Layer.Inhib.Layer.Gi":  "1.5",
					"Layer.Inhib.Layer.FB":  "0.1",
					"Layer.Inhib.Self.Gi":   "0.5",
					"Layer.Learn.AvgL.Init": "0.4",
				}},
			{Sel: ".VSMatrixLayer", Desc: "VSMatrix layers",
				Params: params.Params{
					"Layer.Act.XX1.Gain":    "200",
					"Layer.ActModZero":      "false",
					"Layer.DIParams.Active": "true",
					"Layer.DIParams.PrvQ":   "0",
					"Layer.DIParams.PrvTrl": "6",
					"Layer.DaMod.On":        "false",
					"Layer.Inhib.Self.Gi":   "0.3",
					"Layer.ModNetThreshold": "0.05",
					"Layer.Learn.AvgL.Init": "0.4",
				}},
			{Sel: "#VSMatrixPosD1", Desc: "VSMatrixPosD1 layer",
				Params: params.Params{
					"Layer.Act.XX1.Gain":    "200",
					"Layer.ActModZero":      "false",
					"Layer.DIParams.Active": "true",
					"Layer.DIParams.PrvQ":   "0",
					"Layer.DIParams.PrvTrl": "6",
					"Layer.DaMod.On":        "false",
					"Layer.ModNetThreshold": "0.05",
					"Layer.DaMod.BurstGain": "1",
					"Layer.DaMod.DipGain":   "1",
					"Layer.Inhib.Self.Gi":   "0.3",
					"Layer.Learn.AvgL.Init": "0.4",
				}},
			{Sel: "#VSMatrixPosD2", Desc: "VSMatrixPosD2 layer",
				Params: params.Params{
					"Layer.Act.XX1.Gain":    "200",
					"Layer.ActModZero":      "false",
					"Layer.DIParams.Active": "true",
					"Layer.DIParams.PrvQ":   "0",
					"Layer.DIParams.PrvTrl": "6",
					"Layer.DaMod.On":        "false",
					"Layer.ModNetThreshold": "0.05",
					"Layer.DaMod.BurstGain": "1",
					"Layer.DaMod.DipGain":   "1",
					"Layer.Inhib.Self.Gi":   "0.3",
					"Layer.Learn.AvgL.Init": "0.4",
				}},
			{Sel: "#VSMatrixNegD1", Desc: "VSMatrixNegD1 layer",
				Params: params.Params{
					"Layer.Act.XX1.Gain":    "200",
					"Layer.ActModZero":      "false",
					"Layer.DIParams.Active": "true",
					"Layer.DIParams.PrvQ":   "0",
					"Layer.DIParams.PrvTrl": "6",
					"Layer.DaMod.On":        "false",
					"Layer.ModNetThreshold": "0.05",
					"Layer.DaMod.BurstGain": "1",
					"Layer.DaMod.DipGain":   "1",
					"Layer.Inhib.Self.Gi":   "0.3",
					"Layer.Learn.AvgL.Init": "0.4",
				}},
			{Sel: "#VSMatrixNegD2", Desc: "VSMatrixNegD2 layer",
				Params: params.Params{
					"Layer.Act.XX1.Gain":    "200",
					"Layer.ActModZero":      "false",
					"Layer.DIParams.Active": "true",
					"Layer.DIParams.PrvQ":   "0",
					"Layer.DIParams.PrvTrl": "6",
					"Layer.DaMod.On":        "false",
					"Layer.ModNetThreshold": "0.05",
					"Layer.DaMod.BurstGain": "0.2",
					"Layer.DaMod.DipGain":   "1",
					"Layer.Inhib.Self.Gi":   "0.3",
					"Layer.Learn.AvgL.Init": "0.4",
				}},
			{Sel: "BLAmygLayer", Desc: "BLAmyg layer defaults",
				Params: params.Params{
					"Layer.Inhib.ActAvg.Init":  "0.1",
					"Layer.Inhib.ActAvg.Fixed": "true",
				}},
			{Sel: ".CemPosToPPTg", Desc: "CemPos-PPTg projection, only input to PPTg",
				Params: params.Params{
					"Prjn.Learn.Learn": "false",
				}},
			{Sel: "#BLAmygPosD1", Desc: "BLAmygPosD1 layer",
				Params: params.Params{
					"Layer.DaMod.On":            "true",
					"Layer.Minus":               "1",
					"Layer.Plus":                "1",
					"Layer.DaMod.BurstGain":     "0",
					"Layer.DaMod.DipGain":       "0.05",
					"Layer.ActModZero":          "false",
					"Layer.Act.Dt.VmTau":        "5",
					"Layer.Act.Dt.GTau":         "3",
					"Layer.Act.XX1.Gain":        "40",
					"Layer.Inhib.Layer.On":      "false",
					"Layer.Inhib.Pool.On":       "false",
					"Layer.Inhib.Pool.Gi":       "1",
					"Layer.Inhib.Pool.FBTau":    "3",
					"Layer.Inhib.Pool.MaxVsAvg": "0.5",
					"Layer.Inhib.Self.On":       "true",
					"Layer.Inhib.Self.Gi":       "0.5",
					"Layer.Inhib.ActAvg.Init":   "0.05",
					"Layer.Inhib.ActAvg.Fixed":  "true",
					"Layer.Learn.AvgL.Init":     "0.4",
					"Layer.ILI.Gi":              "1.5",
					"Layer.ILI.Add":             "true",
				}},
			{Sel: "#BLAmygNegD2", Desc: "BLAmygNegD2 layer",
				Params: params.Params{
					"Layer.DaMod.On":            "true",
					"Layer.Minus":               "1",
					"Layer.Plus":                "1",
					"Layer.DaMod.BurstGain":     "0.1",
					"Layer.DaMod.DipGain":       "0",
					"Layer.ActModZero":          "false",
					"Layer.Act.Dt.VmTau":        "5",
					"Layer.Act.Dt.GTau":         "3",
					"Layer.Act.XX1.Gain":        "40",
					"Layer.Inhib.Layer.On":      "false",
					"Layer.Inhib.Pool.On":       "false",
					"Layer.Inhib.Pool.Gi":       "1",
					"Layer.Inhib.Pool.FBTau":    "3",
					"Layer.Inhib.Pool.MaxVsAvg": "0.5",
					"Layer.Inhib.Self.On":       "true",
					"Layer.Inhib.Self.Gi":       "0.5",
					"Layer.Inhib.ActAvg.Init":   "0.05",
					"Layer.Inhib.ActAvg.Fixed":  "true",
					"Layer.ILI.Gi":              "1.5",
					"Layer.ILI.Add":             "true",
				}},
			{Sel: "#BLAmygPosD2", Desc: "BLAmygPosD2 layer",
				Params: params.Params{
					"Layer.Act.XX1.Gain":       "40",
					"Layer.Act.Dt.VmTau":       "5",
					"Layer.Act.Dt.GTau":        "3",
					"Layer.DaMod.On":           "true",
					"Layer.Minus":              "1",
					"Layer.Plus":               "1",
					"Layer.DaMod.BurstGain":    "1",
					"Layer.DaMod.DipGain":      "1",
					"Layer.ActModZero":         "true",
					"Layer.Inhib.Layer.On":     "false",
					"Layer.Inhib.Pool.On":      "false",
					"Layer.Inhib.ActAvg.Init":  "0.2",
					"Layer.Inhib.ActAvg.Fixed": "true",
					"Layer.Inhib.Self.On":      "true",
					"Layer.Inhib.Self.Gi":      "1.2",
					"Layer.ModNetThreshold":    "0.1",
				}},
			{Sel: "#BLAmygNegD1", Desc: "BLAmygNegD1 layer",
				Params: params.Params{
					"Layer.Act.XX1.Gain":       "80",
					"Layer.Act.Dt.VmTau":       "5",
					"Layer.Act.Dt.GTau":        "3",
					"Layer.DaMod.On":           "true",
					"Layer.Minus":              "1",
					"Layer.Plus":               "1",
					"Layer.DaMod.BurstGain":    "1",
					"Layer.DaMod.DipGain":      "1",
					"Layer.ActModZero":         "true",
					"Layer.Inhib.Layer.On":     "false",
					"Layer.Inhib.Pool.On":      "false",
					"Layer.Inhib.ActAvg.Init":  "0.2",
					"Layer.Inhib.ActAvg.Fixed": "true",
					"Layer.Inhib.Self.On":      "true",
					"Layer.Inhib.Self.Gi":      "1.2",
					"Layer.ModNetThreshold":    "0.1",
				}},
			{Sel: "#CElAcqPosD1", Desc: "CElAcqPosD1 layer",
				Params: params.Params{
					"Layer.DaMod.On":          "true",
					"Layer.Minus":             "1",
					"Layer.Plus":              "1",
					"Layer.Act.Init.Vm":       "0.55",
					"Layer.DaMod.BurstGain":   "0",
					"Layer.DaMod.DipGain":     "0.2",
					"Layer.AcqDeepMod":        "true",
					"Layer.ActModZero":        "false",
					"Layer.Inhib.Layer.FB":    "1",
					"Layer.Inhib.Layer.FBTau": "3",
					"Layer.Inhib.Pool.On":     "false",
					"Layer.Inhib.Self.On":     "true",
					"Layer.Inhib.Self.Gi":     "2.5",
					"Layer.Act.Dt.VmTau":      "5",
					"Layer.Act.Dt.GTau":       "3",
					"Layer.Act.XX1.Gain":      "40",
					"Layer.Act.Erev.L":        "0.55",
					"Layer.Act.Erev.I":        "0.4",
				}},
			{Sel: "#CElAcqNegD2", Desc: "CElAcqNegD2 layer",
				Params: params.Params{
					"Layer.DaMod.On":          "true",
					"Layer.Minus":             "1",
					"Layer.Plus":              "1",
					"Layer.Act.Init.Vm":       "0.55",
					"Layer.DaMod.BurstGain":   "0.002",
					"Layer.DaMod.DipGain":     "0",
					"Layer.AcqDeepMod":        "true",
					"Layer.ActModZero":        "false",
					"Layer.Inhib.Layer.FB":    "1",
					"Layer.Inhib.Layer.FBTau": "3",
					"Layer.Inhib.Pool.On":     "false",
					"Layer.Inhib.Self.On":     "true",
					"Layer.Inhib.Self.Gi":     "2.5",
					"Layer.Act.Dt.VmTau":      "5",
					"Layer.Act.Dt.GTau":       "3",
					"Layer.Act.XX1.Gain":      "40",
					"Layer.Act.Erev.L":        "0.55",
					"Layer.Act.Erev.I":        "0.4",
				}},
			{Sel: "#CElExtPosD2", Desc: "CElExtPosD2 layer",
				Params: params.Params{
					"Layer.DaMod.On":          "true",
					"Layer.Minus":             "1",
					"Layer.Plus":              "1",
					"Layer.Act.Init.Vm":       "0.55",
					"Layer.DaMod.BurstGain":   "0.01",
					"Layer.DaMod.DipGain":     "0.05",
					"Layer.AcqDeepMod":        "true",
					"Layer.ActModZero":        "false",
					"Layer.Inhib.Layer.FB":    "1",
					"Layer.Inhib.Layer.FBTau": "3",
					"Layer.Inhib.Pool.On":     "false",
					"Layer.Inhib.Self.On":     "true",
					"Layer.Inhib.Self.Gi":     "3",
					"Layer.Act.Dt.VmTau":      "5",
					"Layer.Act.Dt.GTau":       "3",
					"Layer.Act.XX1.Gain":      "40",
					"Layer.Act.Gbar.L":        "0.8",
					"Layer.Act.Erev.L":        "0.55",
					"Layer.Act.Erev.I":        "0.4",
				}},
			{Sel: "#CElExtNegD1", Desc: "CElExtNegD1 layer",
				Params: params.Params{
					"Layer.DaMod.On":          "true",
					"Layer.Minus":             "1",
					"Layer.Plus":              "1",
					"Layer.Act.Init.Vm":       "0.55",
					"Layer.DaMod.BurstGain":   "0.05",
					"Layer.DaMod.DipGain":     "0.01",
					"Layer.AcqDeepMod":        "true",
					"Layer.ActModZero":        "false",
					"Layer.Inhib.Layer.FB":    "1",
					"Layer.Inhib.Layer.FBTau": "3",
					"Layer.Inhib.Pool.On":     "false",
					"Layer.Inhib.Self.On":     "true",
					"Layer.Inhib.Self.Gi":     "3",
					"Layer.Act.Dt.VmTau":      "5",
					"Layer.Act.Dt.GTau":       "3",
					"Layer.Act.XX1.Gain":      "40",
					"Layer.Act.Erev.L":        "0.55",
					"Layer.Act.Erev.I":        "0.4",
				}},
			{Sel: "#LHbRMTg", Desc: "Lateral Habenula / RMTg",
				Params: params.Params{
					"Layer.Gains.VSPatchPosD1": "1.7",
					"Layer.Gains.VSPatchPosD2": "1.7",
					"Layer.Inhib.Self.On":      "true",
					"Layer.Inhib.Self.Gi":      "0.8",
					"Layer.Inhib.ActAvg.Fixed": "true",
				}},
			{Sel: "#StimIn", Desc: "StimIn layer",
				Params: params.Params{
					"Layer.Inhib.Layer.Gi":     "2",
					"Layer.Inhib.Layer.FB":     "0.5",
					"Layer.Inhib.ActAvg.Init":  "0.1",
					"Layer.Inhib.ActAvg.Fixed": "true",
				}},
			{Sel: "#USTimeIn", Desc: "USTimeIn layer",
				Params: params.Params{
					"Layer.Inhib.Layer.Gi":     "2",
					"Layer.Inhib.Layer.FB":     "0.5",
					"Layer.Inhib.ActAvg.Init":  "0.002",
					"Layer.Inhib.ActAvg.Fixed": "true",
				}},
			{Sel: "#ContextIn", Desc: "ContextIn layer",
				Params: params.Params{
					"Layer.Inhib.Layer.Gi":     "2",
					"Layer.Inhib.Layer.FB":     "0.5",
					"Layer.Inhib.ActAvg.Init":  "0.02",
					"Layer.Inhib.ActAvg.Fixed": "true",
				}},
			{Sel: "#PPTg", Desc: "PPTg layer",
				Params: params.Params{
					"Layer.ActThreshold":    "0.02",
					"Layer.ClampActivation": "true",
					"Layer.DNetGain":        "1",
				}},
			{Sel: "#VTAp", Desc: "VTAp layer",
				Params: params.Params{
					"Layer.Act.Clamp.Range.Min": "-2",
					"Layer.Act.Clamp.Range.Max": "2",
					"Layer.DAGains.PPTg":        "1.1",
				}},
			{Sel: "#VTAn", Desc: "VTAn layer",
				Params: params.Params{
					"Layer.Act.Clamp.Range.Min":   "-2",
					"Layer.Act.Clamp.Range.Max":   "2",
					"Layer.DAGains.PVIBurstShunt": "0.9",
				}},
		},
	}},
}
