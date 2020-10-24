// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package leabra

import (
	"fmt"
	"reflect"
)

// leabra.Synapse holds state for the synaptic connection between neurons
type Synapse struct {
	Wt     float32 `desc:"synaptic weight value -- sigmoid contrast-enhanced"`
	LWt    float32 `desc:"linear (underlying) weight value -- learns according to the lrate specified in the connection spec -- this is converted into the effective weight value, Wt, via sigmoidal contrast enhancement (see WtSigParams)"`
	DWt    float32 `desc:"change in synaptic weight, from learning"`
	Norm   float32 `desc:"DWt normalization factor -- reset to max of abs value of DWt, decays slowly down over time -- serves as an estimate of variance in weight changes over time"`
	Moment float32 `desc:"momentum -- time-integrated DWt changes, to accumulate a consistent direction of weight change and cancel out dithering contradictory changes"`
	Scale  float32 `desc:"scaling parameter for this connection: effective weight value is scaled by this factor -- useful for topographic connectivity patterns e.g., to enforce more distant connections to always be lower in magnitude than closer connections.  Value defaults to 1 (cannot be exactly 0 -- otherwise is automatically reset to 1 -- use a very small number to approximate 0).  Typically set by using the prjn.Pattern Weights() values where appropriate"`

	// DS: Sleep related var additions
	SRAvgDp   float32 `desc:"Synaptic Depression scaling variable based on sender-receiver neuron average activation (represented by the inverse of sum of co-activation)"`
	Cai       float32 `desc:"cai intacellular calcium. Default to be 0."`
	Effwt     float32 `desc:"Maybe it is needed. I don't know yet. Default to be the same as Wt."`
	SenRecAct float32 `desc:"#READ_ONLY rescaling factor taking into account sd_ca_gain and sd_ca_thr (= sd_ca_gain/(1 - sd_ca_thr))"`
	SynDepFac float32 `desc:"#READ_ONLY Final Calcualted Synaptic Depression at the Synapse"`
	ActPAvg	  float32 `desc:"#READ_ONLY Final ActP at the Synapse"`
	ActMAvg   float32 `desc:"#READ_ONLY Final ActM at the Synapse"`
	RunSum 	  float32 `desc:"#READ_ONLY Running Sum of coactivation"`
	//su_act    float32 `desc:"#READ_ONLY rescaling factor taking into account sd_ca_gain and sd_ca_thr (= sd_ca_gain/(1 - sd_ca_thr))"`
	//ru_act    float32 `desc:"#READ_ONLY rescaling factor taking into account sd_ca_gain and sd_ca_thr (= sd_ca_gain/(1 - sd_ca_thr))"`
	//Rec            float32 `desc:"// #DEF_0.002 rate of recovery from depression"`
	CaInc          float32 `desc:"#DEF_0.2 time constant for increases in Ca_i (from NMDA etc currents) -- default base value is .01 per cycle -- multiply by network->ct_learn.syndep_int to get this value (default = 20)"`
	CaDec          float32 `desc:"#DEF_0.2 time constant for decreases in Ca_i (from Ca pumps pushing Ca back out into the synapse) -- default base value is .01 per cycle -- multiply by network->ct_learn.syndep_int to get this value (default = 20)"`
	SdCaThr        float32 `desc:"#DEF_0.2 synaptic depression ca threshold: only when ca_i has increased by this amount (thus synaptic ca depleted) does it affect firing rates and thus synaptic depression"`
	SdCaGain       float32 `desc:"#DEF_0.3 multiplier on cai value for computing synaptic depression -- modulates overall level of depression independent of rate parameters"`
	SdCaThrRescale float32 `desc:"#READ_ONLY rescaling factor taking into account sd_ca_gain and sd_ca_thr (= sd_ca_gain/(1 - sd_ca_thr))"`
	Sleep          float32 `desc:"#READ_ONLY rescaling factor taking into account sd_ca_gain and sd_ca_thr (= sd_ca_gain/(1 - sd_ca_thr))"`
}

var SynapseVars = []string{"Wt", "LWt", "DWt", "Norm", "Moment", "Scale", "SRAvgDp", "Cai", "Effwt", "SenRecAct", "SynDepFac", "ActPAvg", "ActMAvg"} //, "su_act", "ru_act"} //, "CaInc", "CaDec", "SdCaThr", "SdCaGain", "SdCaThrRescale"}

var SynapseVarProps = map[string]string{
	"DWt":    `auto-scale:"+"`,
	"Moment": `auto-scale:"+"`,
}

var SynapseVarsMap map[string]int

func init() {
	SynapseVarsMap = make(map[string]int, len(SynapseVars))
	for i, v := range SynapseVars {
		SynapseVarsMap[v] = i
	}
}

func (sy *Synapse) VarNames() []string {
	return SynapseVars
}

// SynapseVarByName returns the index of the variable in the Synapse, or error
func SynapseVarByName(varNm string) (int, error) {
	i, ok := SynapseVarsMap[varNm]
	if !ok {
		return 0, fmt.Errorf("Synapse VarByName: variable name: %v not valid", varNm)
	}
	return i, nil
}

// VarByIndex returns variable using index (0 = first variable in SynapseVars list)
func (sy *Synapse) VarByIndex(idx int) float32 {
	// todo: would be ideal to avoid having to use reflect here..
	v := reflect.ValueOf(*sy)
	return v.Field(idx).Interface().(float32)
}

// VarByName returns variable by name, or error
func (sy *Synapse) VarByName(varNm string) (float32, error) {
	i, err := SynapseVarByName(varNm)
	if err != nil {
		return 0, err
	}
	return sy.VarByIndex(i), nil
}

func (sy *Synapse) SetVarByIndex(idx int, val float32) {
	// todo: would be ideal to avoid having to use reflect here..
	v := reflect.ValueOf(sy)
	v.Elem().Field(idx).SetFloat(float64(val))
}

// SetVarByName sets synapse variable to given value
func (sy *Synapse) SetVarByName(varNm string, val float32) error {
	i, err := SynapseVarByName(varNm)
	if err != nil {
		return err
	}
	sy.SetVarByIndex(i, val)
	return nil
}

// Added by DZ: SynDep calculated at each synapse
func (sy *Synapse) SynDep() float32 {
	cao_thr := float32(1.0)
	if sy.Cai > sy.SdCaThr {
		cao_thr = 1.0 - (sy.SdCaThrRescale * (sy.Cai - sy.SdCaThr))
	}
	sy.SynDepFac = cao_thr * cao_thr
	return (cao_thr * cao_thr)
}

// Added by DZ: CaUpdt calculated the Cai for each synapses
func (sy *Synapse) CaUpdt(ru_act float32, su_act float32) {
	sy.SenRecAct = ru_act * su_act
	//sy.su_act = su_act
	//sy.ru_act = ru_act
	//fmt.Println(ru_act, su_act, sy.SenRecAct)
	//drive := float32(0.0)
	//if ru_act*su_act > 0.0001 {
	//	drive = ((ru_act + su_act) / 2) * sy.Effwt
	//} else {
	//	drive = (ru_act * su_act) * sy.Effwt
	//}
	drive := (ru_act * su_act) * sy.Effwt
	sy.Cai += (sy.CaInc * (1.0 - sy.Cai) * drive) - (sy.CaDec * sy.Cai)
}

func (sy *Synapse) RunSumUpdt(init bool, ru_act float32, su_act float32, ) {
	if sy.Sleep == 1 { // adding extra check just in case
		if init {
			sy.RunSum = 0
			sy.RunSum = (ru_act * su_act)
		} else {
			sy.RunSum = sy.RunSum + (ru_act * su_act)
		}
	}
}

// CalcActP calculates final ActP values for each synapse
func (sy *Synapse) CalcActP(pluscount int) {
	sy.ActPAvg = sy.RunSum / float32(pluscount)
	sy.RunSum = 0
}

// CalcActQ calculates final ActQ values for each synapse
func (sy *Synapse) CalcActM(minuscount int) {
	sy.ActMAvg = sy.RunSum / float32(minuscount)
	sy.RunSum = 0
}

func (sy *Synapse) EffwtUpdt() {
	if sy.Sleep == 0 {
		sy.Effwt = sy.Wt * sy.SynDep()
	} else if sy.Sleep == 1 {
		sy.Effwt = sy.Wt * (sy.SynDep())
	}
	//fmt.Println(sy.Wt - sy.Effwt)
	// Final checking if the Effwt is out of bounds
	if sy.Effwt > sy.Wt {
		sy.Effwt = sy.Wt
	}
	if sy.Effwt < 0.0 {
		sy.Effwt = 0.0
	}
}
