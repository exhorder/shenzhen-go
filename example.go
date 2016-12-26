// Copyright 2016 Google Inc.
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

package main

import (
	"strings"

	"shenzhen-go/graph"
)

const exampleSrc = `{
	"name": "Example",
	"package_name": "example",
	"package_path": "example",
	"imports": [
		"fmt"
	],
	"nodes": {
		"Filter divisible by 2": {
			"name": "Filter divisible by 2",
			"wait": true,
			"part": {
				"code": "for n := range raw {\n\tif n \u003e 2 \u0026\u0026 n % 2 == 0 {\n\t\tcontinue\n\t}\n\tdiv2 \u003c- n\n}\nclose(div2)"
			},
			"part_type": "Code"
		},
		"Filter divisible by 3": {
			"name": "Filter divisible by 3",
			"wait": true,
			"part": {
				"code": "for n := range div2 {\n\tif n \u003e 3 \u0026\u0026 n % 3 == 0 {\n\t\tcontinue\n\t}\n\tout \u003c- n\n}\nclose(out)"
			},
			"part_type": "Code"
		},
		"Generate integers ≥ 2": {
			"name": "Generate integers ≥ 2",
			"wait": true,
			"part": {
				"code": "for i:= 2; i\u003c100; i++ {\n\traw \u003c- i\n}\nclose(raw)"
			},
			"part_type": "Code"
		},
		"Print output": {
			"name": "Print output",
			"wait": true,
			"part": {
				"code": "for n := range out {\n\tfmt.Println(n)\n}"
			},
			"part_type": "Code"
		}
	},
	"channels": {
		"div2": {
			"name": "div2",
			"type": "int",
			"cap": 0
		},
		"out": {
			"name": "out",
			"type": "int",
			"cap": 0
		},
		"raw": {
			"name": "raw",
			"type": "int",
			"cap": 0
		}
	}
}`

var exampleGraph *graph.Graph

func init() {
	g, err := graph.LoadJSON(strings.NewReader(exampleSrc))
	if err != nil {
		panic(err)
	}
	exampleGraph = g
}
