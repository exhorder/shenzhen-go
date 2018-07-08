// Copyright 2018 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package parts

import (
	"fmt"

	"github.com/google/shenzhen-go/dev/model"
	"github.com/google/shenzhen-go/dev/model/pin"
)

const keyCounterTypeParam = "$Key"

var keyCounterPins = pin.NewMap(
	&pin.Definition{
		Name:      "input",
		Direction: pin.Input,
		Type:      keyCounterTypeParam,
	},
	&pin.Definition{
		Name:      "output",
		Direction: pin.Output,
		Type:      keyCounterTypeParam,
	},
	&pin.Definition{
		Name:      "result",
		Direction: pin.Output,
		Type:      fmt.Sprintf("map[%s]uint", keyCounterTypeParam),
	})

func init() {
	model.RegisterPartType("KeyCounter", &model.PartType{
		New: func() model.Part { return &KeyCounter{} },
		Panels: []model.PartPanel{{
			Name: "Help",
			Editor: `<div>
			<p>
				A KeyCounter passes through values from input to output, counting
				how many of each value it sees in a map. When the input is closed,
				the map is sent on the result channel, and both outputs are closed.
			</p>
			</div>`,
		}},
	})
}

// KeyCounter produces a frequency table.
type KeyCounter struct{}

// Clone returns a clone of this Closer.
func (KeyCounter) Clone() model.Part { return &KeyCounter{} }

// Impl returns the Closer implementation.
func (KeyCounter) Impl(types map[string]string) (head, body, tail string) {
	return fmt.Sprintf("m := make(map[%s]uint)", types[keyCounterTypeParam]),
		"for in := range input { m[in]++; output <- in }",
		"result <- m; close(output); close(result)"
}

// Imports returns nil.
func (KeyCounter) Imports() []string { return nil }

// Pins returns a map declaring a single output of any type.
func (KeyCounter) Pins() pin.Map { return keyCounterPins }

// TypeKey returns "Closer".
func (KeyCounter) TypeKey() string { return "KeyCounter" }
