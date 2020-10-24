// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package leabra

import (
	"fmt"
	"unsafe"

	"github.com/goki/ki/bitflag"
	"github.com/goki/ki/kit"
)

// NeuronVarStart is the byte offset of fields in the Neuron structure
// where the float32 named variables start.
// Note: all non-float32 infrastructure variables must be at the start!
const NeuronVarStart = 8

// leabra.Neuron holds all of the neuron (unit) level variables -- this is the most basic version with
// rate-code only and no optional features at all.
// All variables accessible via Unit interface must be float32 and start at the top, in contiguous order
type Neuron struct {
	Flags   NeurFlags `desc:"bit flags for binary state variables"`
	SubPool int32     `desc:"index of the sub-level inhibitory pool that this neuron is in (only for 4D shapes, the pool (unit-group / hypercolumn) structure level) -- indicies start at 1 -- 0 is layer-level pool (is 0 if no sub-pools)."`
	Act     float32   `desc:"rate-coded activation value reflecting final output of neuron communicated to other neurons, typically in range 0-1.  This value includes adaptation and synaptic depression / facilitation effects which produce temporal contrast (see ActLrn for version without this).  For rate-code activation, this is noisy-x-over-x-plus-one (NXX1) function; for discrete spiking it is computed from the inverse of the inter-spike interval (ISI), and Spike reflects the discrete spikes."`
	ActLrn  float32   `desc:"learning activation value, reflecting *dendritic* activity that is not affected by synaptic depression or adapdation channels which are located near the axon hillock.  This is the what drives the Avg* values that drive learning. Computationally, neurons strongly discount the signals sent to other neurons to provide temporal contrast, but need to learn based on a more stable reflection of their overall inputs in the dendrites."`
	Ge      float32   `desc:"total excitatory synaptic conductance -- the net excitatory input to the neuron -- does *not* include Gbar.E"`
	Gi      float32   `desc:"total inhibitory synaptic conductance -- the net inhibitory input to the neuron -- does *not* include Gbar.I"`
	Gk      float32   `desc:"total potassium conductance, typically reflecting sodium-gated potassium currents involved in adaptation effects -- does *not* include Gbar.K"`
	Inet    float32   `desc:"net current produced by all channels -- drives update of Vm"`
	Vm      float32   `desc:"membrane potential -- integrates Inet current over time"`

	Targ float32 `desc:"target value: drives learning to produce this activation value"`
	Ext  float32 `desc:"external input: drives activation of unit from outside influences (e.g., sensory input)"`

	AvgSS   float32 `desc:"super-short time-scale average of ActLrn activation -- provides the lowest-level time integration -- for spiking this integrates over spikes before subsequent averaging, and it is also useful for rate-code to provide a longer time integral overall"`
	AvgS    float32 `desc:"short time-scale average of ActLrn activation -- tracks the most recent activation states (integrates over AvgSS values), and represents the plus phase for learning in XCAL algorithms"`
	AvgM    float32 `desc:"medium time-scale average of ActLrn activation -- integrates over AvgS values, and represents the minus phase for learning in XCAL algorithms"`
	AvgL    float32 `desc:"long time-scale average of medium-time scale (trial level) activation, used for the BCM-style floating threshold in XCAL"`
	AvgLLrn float32 `desc:"how much to learn based on the long-term floating threshold (AvgL) for BCM-style Hebbian learning -- is modulated by level of AvgL itself (stronger Hebbian as average activation goes higher) and optionally the average amount of error experienced in the layer (to retain a common proportionality with the level of error-driven learning across layers)"`
	AvgSLrn float32 `desc:"short time-scale activation average that is actually used for learning -- typically includes a small contribution from AvgM in addition to mostly AvgS, as determined by LrnActAvgParams.LrnM -- important to ensure that when unit turns off in plus phase (short time scale), enough medium-phase trace remains so that learning signal doesn't just go all the way to 0, at which point no learning would take place"`

	ActQ0  float32 `desc:"the activation state at start of current alpha cycle (same as the state at end of previous cycle)"`
	ActQ1  float32 `desc:"the activation state at end of first quarter of current alpha cycle"`
	ActQ2  float32 `desc:"the activation state at end of second quarter of current alpha cycle"`
	ActM   float32 `desc:"the activation state at end of third quarter, which is the traditional posterior-cortical minus phase activation"`
	ActP   float32 `desc:"the activation state at end of fourth quarter, which is the traditional posterior-cortical plus_phase activation"`
	ActDif float32 `desc:"ActP - ActM -- difference between plus and minus phase acts -- reflects the individual error gradient for this neuron in standard error-driven learning terms"`
	ActDel float32 `desc:"delta activation: change in Act from one cycle to next -- can be useful to track where changes are taking place"`
	ActAvg float32 `desc:"average activation (of final plus phase activation state) over long time intervals (time constant = DtPars.AvgTau -- typically 200) -- useful for finding hog units and seeing overall distribution of activation"`

	Noise float32 `desc:"noise value added to unit (ActNoiseParams determines distribution, and when / where it is added)"`

	GiSyn    float32 `desc:"aggregated synaptic inhibition (from Inhib projections) -- time integral of GiRaw -- this is added with computed FFFB inhibition to get the full inhibition in Gi"`
	GiSelf   float32 `desc:"total amount of self-inhibition -- time-integrated to avoid oscillations"`
	ActSent  float32 `desc:"last activation value sent (only send when diff is over threshold)"`
	GeRaw    float32 `desc:"raw excitatory conductance (net input) received from sending units (send delta's are added to this value)"`
	GeInc    float32 `desc:"delta increment in GeRaw sent using SendGeDelta"`
	GiRaw    float32 `desc:"raw inhibitory conductance (net input) received from sending units (send delta's are added to this value)"`
	GiInc    float32 `desc:"delta increment in GiRaw sent using SendGeDelta"`
	GknaFast float32 `desc:"conductance of sodium-gated potassium channel (KNa) fast dynamics (M-type) -- produces accommodation / adaptation of firing"`
	GknaMed  float32 `desc:"conductance of sodium-gated potassium channel (KNa) medium dynamics (Slick) -- produces accommodation / adaptation of firing"`
	GknaSlow float32 `desc:"conductance of sodium-gated potassium channel (KNa) slow dynamics (Slack) -- produces accommodation / adaptation of firing"`

	Spike  float32 `desc:"whether neuron has spiked or not (0 or 1), for discrete spiking neurons."`
	ISI    float32 `desc:"current inter-spike-interval -- counts up since last spike.  Starts at -1 when initialized."`
	ISIAvg float32 `desc:"average inter-spike-interval -- average time interval between spikes.  Starts at -1 when initialized, and goes to -2 after first spike, and is only valid after the second spike post-initialization."`

	EffAct float32 `desc:"SLEEP rate-coded activation value reflecting final output of neuron communicated to other neurons, typically in range 0-1."`
}

var NeuronVars = []string{"Act", "ActLrn", "Ge", "Gi", "Gk", "Inet", "Vm", "Targ", "Ext", "AvgSS", "AvgS", "AvgM", "AvgL",
	"AvgLLrn", "AvgSLrn", "ActQ0", "ActQ1", "ActQ2", "ActM", "ActP", "ActDif", "ActDel", "ActAvg", "Noise", "GiSyn", "GiSelf",
	"ActSent", "GeRaw", "GeInc", "GiRaw", "GiInc", "GknaFast", "GknaMed", "GknaSlow", "Spike", "ISI", "ISIAvg", "EffAct"}

var NeuronVarsMap map[string]int

var NeuronVarProps = map[string]string{
	"Vm":     `min:"0" max:"1"`,
	"ActDel": `auto-scale:"+"`,
	"ActDif": `auto-scale:"+"`,
}

func init() {
	NeuronVarsMap = make(map[string]int, len(NeuronVars))
	for i, v := range NeuronVars {
		NeuronVarsMap[v] = i
	}
}

func (nrn *Neuron) VarNames() []string {
	return NeuronVars
}

// NeuronVarByName returns the index of the variable in the Neuron, or error
func NeuronVarByName(varNm string) (int, error) {
	i, ok := NeuronVarsMap[varNm]
	if !ok {
		return 0, fmt.Errorf("Neuron VarByName: variable name: %v not valid", varNm)
	}
	return i, nil
}

// VarByIndex returns variable using index (0 = first variable in NeuronVars list)
func (nrn *Neuron) VarByIndex(idx int) float32 {
	fv := (*float32)(unsafe.Pointer(uintptr(unsafe.Pointer(nrn)) + uintptr(NeuronVarStart+4*idx)))
	return *fv
}

// VarByName returns variable by name, or error
func (nrn *Neuron) VarByName(varNm string) (float32, error) {
	i, err := NeuronVarByName(varNm)
	if err != nil {
		return 0, err
	}
	return nrn.VarByIndex(i), nil
}

func (nrn *Neuron) HasFlag(flag NeurFlags) bool {
	return bitflag.Has32(int32(nrn.Flags), int(flag))
}

func (nrn *Neuron) SetFlag(flag NeurFlags) {
	bitflag.Set32((*int32)(&nrn.Flags), int(flag))
}

func (nrn *Neuron) ClearFlag(flag NeurFlags) {
	bitflag.Clear32((*int32)(&nrn.Flags), int(flag))
}

func (nrn *Neuron) SetMask(mask int32) {
	bitflag.SetMask32((*int32)(&nrn.Flags), mask)
}

func (nrn *Neuron) ClearMask(mask int32) {
	bitflag.ClearMask32((*int32)(&nrn.Flags), mask)
}

// IsOff returns true if the neuron has been turned off (lesioned)
func (nrn *Neuron) IsOff() bool {
	return nrn.HasFlag(NeurOff)
}

// NeurFlags are bit-flags encoding relevant binary state for neurons
type NeurFlags int32

//go:generate stringer -type=NeurFlags

var KiT_NeurFlags = kit.Enums.AddEnum(NeurFlagsN, kit.BitFlag, nil)

func (ev NeurFlags) MarshalJSON() ([]byte, error)  { return kit.EnumMarshalJSON(ev) }
func (ev *NeurFlags) UnmarshalJSON(b []byte) error { return kit.EnumUnmarshalJSON(ev, b) }

// The neuron flags
const (
	// NeurOff flag indicates that this neuron has been turned off (i.e., lesioned)
	NeurOff NeurFlags = iota

	// NeurHasExt means the neuron has external input in its Ext field
	NeurHasExt

	// NeurHasTarg means the neuron has external target input in its Targ field
	NeurHasTarg

	// NeurHasCmpr means the neuron has external comparison input in its Targ field -- used for computing
	// comparison statistics but does not drive neural activity ever
	NeurHasCmpr

	NeurFlagsN
)

/*
more specialized flags in C++ emergent -- only add in specialized cases where needed, although
there could be conflicts potentially, so may want to just go ahead and add here..
  enum LeabraUnitFlags {        // #BITS extra flags on top of ext flags for leabra
    SUPER       = 0x00000100,   // superficial layer neocortical cell -- has deep.on role = SUPER
    DEEP        = 0x00000200,   // deep layer neocortical cell -- has deep.on role = DEEP
    TRC         = 0x00000400,   // thalamic relay cell (Pulvinar) cell -- has deep.on role = TRC

    D1R         = 0x00001000,   // has predominantly D1 receptors
    D2R         = 0x00002000,   // has predominantly D2 receptors
    ACQUISITION = 0x00004000,   // involved in Acquisition
    EXTINCTION  = 0x00008000,   // involved in Extinction
    APPETITIVE  = 0x00010000,   // appetitive (positive valence) coding
    AVERSIVE    = 0x00020000,   // aversive (negative valence) coding
    PATCH       = 0x00040000,   // patch-like structure (striosomes)
    MATRIX      = 0x00080000,   // matrix-like structure
    DORSAL      = 0x00100000,   // dorsal
    VENTRAL     = 0x00200000,   // ventral
  };

*/
