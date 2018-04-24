// Copyright 2017 Google Inc.
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

package view

import (
	"log"

	"github.com/google/shenzhen-go/dev/dom"
	"golang.org/x/net/context"
)

// Channel is the view's model of a channel.
type Channel struct {
	Group // Container for all the channel elements.

	cc     ChannelController
	view   *View
	errors errorViewer
	graph  *Graph

	// Cache of raw Pin objects which are connected.
	Pins map[*Pin]*Route

	created bool // create operation sent to server?

	steiner            dom.Element // symbol representing the channel itself, not used if channel is simple
	logical            Point       // centre of steiner point, for snapping
	visual             Point       // temporary centre of steiner point, for display
	dragLine, dragCirc dom.Element // temporarily visible, for dragging to more pins
	potentialPin       *Pin        // considering attaching to this pin
}

func (c *Channel) reallyCreate() {
	if err := c.cc.Commit(context.TODO()); err != nil {
		c.errors.setError("Couldn't create a channel: " + err.Error())
		return
	}
	c.created = true
}

// MakeElements creates elements for this channel and adds them to the parent.
func (c *Channel) MakeElements(doc dom.Document, parent dom.Element) {
	if c.Group == (Group{}) {
		c.Group = NewGroup(doc, parent)
	}

	if c.steiner == nil {
		c.steiner = doc.MakeSVGElement("circle").
			SetAttribute("r", pinRadius).
			AddEventListener("mousedown", c.view.dragStarter(c))
	}

	if c.dragLine == nil {
		c.dragLine = doc.MakeSVGElement("line").
			SetAttribute("stroke-width", lineWidth).
			Hide()
	}

	if c.dragCirc == nil {
		c.dragCirc = doc.MakeSVGElement("circle").
			SetAttribute("r", pinRadius).
			Hide()
	}

	c.Group.AddChildren(c.steiner, c.dragLine, c.dragCirc)
}

// Pt implements Pointer.
func (c *Channel) Pt() (x, y float64) { return c.logical.Pt() }

func (c *Channel) commit() {
	if c == nil {
		return
	}
	c.logical = c.visual
	if !c.created {
		go c.reallyCreate()
	}
}

// Show the temporary drag elements.
func (c *Channel) dragTo(x, y float64) {
	c.dragLine.
		SetAttribute("x1", x).
		SetAttribute("y1", y).
		Show()

	c.dragCirc.
		SetAttribute("cx", x).
		SetAttribute("cy", y).
		Show()
}

func (c *Channel) dragStart(x, y float64) {
	// TODO: make it so that if the current configuration is invalid
	// (e.g. all input pins / output pins) then use errorColour, and
	// delete the whole channel if dropped.

	c.steiner.Show()
	c.SetColour(activeColour)

	c.reposition(Point{x, y})
	c.dragTo(x, y)
}

func (c *Channel) drag(x, y float64) {
	c.steiner.Show()
	c.dragTo(x, y)
	d, q := c.graph.nearestPoint(x, y)
	p, _ := q.(*Pin)

	// Already connected to this pin?
	if p != nil && p == c.potentialPin && d < snapQuad {
		return
	}

	// Was considering connecting to a pin, but now connecting to a
	// different pin?
	if c.potentialPin != nil && (c.potentialPin != p || d >= snapQuad) {
		c.potentialPin.disconnect()
		c.potentialPin.SetColour(normalColour)
		c.potentialPin = nil
	}

	noSnap := func() {
		c.dragCirc.Show()
		c.dragLine.Show()
		c.reposition(Point{x, y})
	}

	// Too far from something to snap to?
	// Trying to snap to itself?
	// Don't snap, but not an error.
	if d >= snapQuad || q == c || (p != nil && p.channel == c) {
		c.errors.clearError()
		noSnap()
		c.SetColour(activeColour)
		return
	}

	// Trying to snap to a different channel.
	if p == nil || p.channel != nil {
		c.errors.setError("Can't connect different channels together (use another goroutine)")
		noSnap()
		c.SetColour(errorColour)
		return
	}

	// Snap to pin p!
	c.errors.clearError()
	c.potentialPin = p
	c.SetColour(activeColour)
	c.dragLine.Hide()
	c.dragCirc.Hide()
}

func (c *Channel) drop() {
	c.errors.clearError()
	c.reposition(nil)
	c.commit()
	c.SetColour(normalColour)
	if c.potentialPin != nil {
		c.potentialPin = nil
		return
	}
	c.dragCirc.Hide()
	c.dragLine.Hide()
	if len(c.Pins) <= 2 {
		c.steiner.Hide()
	}
}

func (c *Channel) gainFocus() {
	log.Print("TODO(josh): implement Channel.gainFocus")
}

func (c *Channel) loseFocus() {
	log.Print("TODO(josh): implement Channel.loseFocus")
}

func (c *Channel) save() {
	log.Print("TODO(josh): implement Channel.save")
}

func (c *Channel) delete() {
	if c.created {
		if err := c.cc.Delete(context.TODO()); err != nil {
			c.errors.setError("Couldn't delete channel: " + err.Error())
			return
		}
	}

	// Reset all attached pins, remove all the elements, delete from graph.
	for q := range c.Pins {
		q.channel = nil
	}
	c.Group.Remove()
	delete(c.graph.Channels, c.cc.Name())
}

func (c *Channel) reposition(additional Pointer) {
	if c == nil {
		return
	}

	np := len(c.Pins)
	if additional != nil {
		np++
	}
	if np < 2 {
		// Not actually a channel anymore - hide.
		c.Hide()
		return
	}
	c.visual = Point{0, 0}
	if additional != nil {
		c.visual.Set(additional.Pt())
	}
	for p := range c.Pins {
		c.visual.Add(p.Pt())
	}
	c.visual.Scale(1.0 / float64(np))
	c.steiner.
		SetAttribute("cx", c.visual.x).
		SetAttribute("cy", c.visual.y)
	c.dragLine.
		SetAttribute("x2", c.visual.x).
		SetAttribute("y2", c.visual.y)
	for _, r := range c.Pins {
		r.Reroute()
	}
	if np <= 2 {
		c.steiner.Hide()
	} else {
		c.steiner.Show()
	}
}

// SetColour changes the colour of the whole channel.
func (c *Channel) SetColour(col string) {
	c.steiner.SetAttribute("fill", col)
	c.dragCirc.SetAttribute("fill", col)
	c.dragLine.SetAttribute("stroke", col)
	for p, r := range c.Pins {
		p.SetColour(col)
		r.SetStroke(col)
	}
}
