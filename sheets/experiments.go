/*******************************************************************************
 * Copyright (c) 2025 Genome Research Ltd.
 *
 * Authors:
 *	- Sendu Bala <sb10@sanger.ac.uk>
 *
 * Permission is hereby granted, free of charge, to any person obtaining
 * a copy of this software and associated documentation files (the
 * "Software"), to deal in the Software without restriction, including
 * without limitation the rights to use, copy, modify, merge, publish,
 * distribute, sublicense, and/or sell copies of the Software, and to
 * permit persons to whom the Software is furnished to do so, subject to
 * the following conditions:
 *
 * The above copyright notice and this permission notice shall be included
 * in all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
 * EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
 * MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
 * IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY
 * CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT,
 * TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
 * SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 ******************************************************************************/

package sheets

type SequenceType string

const (
	SequenceTypeNC   SequenceType = "noncoding"
	SequenceTypeC    SequenceType = "coding"
	SequenceTypeAuto SequenceType = "auto"
)

type MutagenesisType string

const (
	MutagenesisTypeRandom MutagenesisType = "random"
	MutagenesisTypeCodon  MutagenesisType = "codon"
)

type Experiment struct {
	ExperimentID                   string
	Assay                          string
	ProjectName                    string
	StartStage                     int
	StopStage                      int
	BarcodeDesignPath              string
	BarcodeErrorRate               string
	ExperimentDesignPairDuplicates bool
	CountPath                      string
	BarcodeIdentityPath            string
	Cutadapt5First                 string
	Cutadapt5Second                string
	CutadaptMinLength              int
	CutadaptErrorRate              string
	CutadaptOverlap                int
	CutadaptCut5First              string
	CutadaptCut5Second             string
	CutadaptCut3First              string
	CutadaptCut3Second             string
	VsearchMinQual                 int
	VsearchMaxQual                 int
	VsearchMaxee                   int
	VsearchMinovlen                int
	ReverseComplement              bool
	WildtypeSequence               string
	PermittedSequences             string
	SequenceType                   SequenceType
	MutagenesisType                MutagenesisType
	Indels                         string
	MaxSubstitutions               int
	MixedSubstitutions             bool
	FitnessMinInputCountAll        int
	FitnessMinInputCountAny        int
	FitnessMinOutputCountAll       int
	FitnessMinOutputCountAny       int
	FitnessNormalise               bool
	FitnessErrorModel              bool
	FitnessDropoutPseudocount      int
	RetainedReplicates             string
	Stranded                       bool
	Paired                         bool
	SynonymSequencePath            string
	TransLibrary                   bool
	TransLibraryReverseComplement  bool
	Samples                        []*Sample
}

// Clone returns a new Experiment with the same values as the original, but
// with the given samples inside it.
func (e *Experiment) Clone(samples []*Sample) *Experiment {
	newE := *e
	newE.Samples = samples

	return &newE
}
