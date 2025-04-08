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

import (
	"strconv"

	"github.com/wtsi-hgi/dimsum-automation/types"
)

// converter converts strings to other types. The conversions do not return
// errors, but instead set the error field. Check that field after doing all
// your conversions.
type converter struct {
	Err error
}

// ToInt converts a string to an int. If the conversion fails, the error
// field is set, and 0 is returned.
//
// If the error field is already set, this function does nothing and returns 0.
func (c *converter) ToInt(s string) int {
	if c.Err != nil {
		return 0
	}

	if s == "" {
		return 0
	}

	i, err := strconv.Atoi(s)
	if err != nil {
		c.Err = err

		return 0
	}

	return i
}

// ToFloatString returns the given string, but checks it was a float. If it was
// not, the error field is set.
func (c *converter) ToFloatString(s string) string {
	c.ToFloat(s)

	if c.Err != nil {
		return s
	}

	return s
}

// ToFloat converts a string to a float. If the conversion fails, the error
// field is set, and 0 is returned.
//
// If the error field is already set, this function does nothing and returns 0.
func (c *converter) ToFloat(s string) float32 {
	if c.Err != nil {
		return 0
	}

	if s == "" {
		return 0
	}

	f, err := strconv.ParseFloat(s, 32)
	if err != nil {
		c.Err = err

		return 0
	}

	return float32(f)
}

// ToBool converts a string to a bool. If the conversion fails, the error field
// is set, and false is returned.
//
// If the error field is already set, this function does nothing and returns
// false.
func (c *converter) ToBool(s string) bool {
	if c.Err != nil {
		return false
	}

	if s == "" {
		return false
	}

	b, err := strconv.ParseBool(s)
	if err != nil {
		c.Err = err

		return false
	}

	return b
}

// ToMutagenesisType converts a string to a MutagenesisType. If the conversion
// fails, the error field is set, and MutagenesisTypeRandom is returned.
//
// If the error field is already set, this function does nothing and returns
// MutagenesisTypeRandom.
func (c *converter) ToMutagenesisType(s string) types.MutagenesisType {
	if c.Err != nil {
		return types.MutagenesisTypeRandom
	}

	mt, err := types.StringToMutagenesisType(s)
	c.Err = err

	return mt
}

// ToSequenceType converts a string to a SequenceType. If the conversion fails,
// the error field is set, and SequenceTypeAuto is returned.
//
// If the error field is already set, this function does nothing and returns
// SequenceTypeAuto.
func (c *converter) ToSequenceType(s string) types.SequenceType {
	if c.Err != nil {
		return types.SequenceTypeAuto
	}

	st, err := types.StringToSequenceType(s)
	c.Err = err

	return st
}

// ToSelection converts a string to a Selection. If the conversion fails, the
// error field is set, and SelectionInput is returned.
//
// If the error field is already set, this function does nothing and returns
// SelectionInput.
func (c *converter) ToSelection(s string) types.Selection {
	if c.Err != nil {
		return types.SelectionInput
	}

	sele, err := types.StringToSelection(s)
	c.Err = err

	return sele
}
