// Copyright 2013 Artur Grabowski. All rights reserved.
// Use of this source code is governed by a ISC-style
// license that can be found in the LICENSE file.
package timers

import (
	"github.com/art4711/stopwatch"
	"time"
)

// Timer collects structured timing data.
// Timers are organized in a tree where each sub-Timer is added to the totals
// of the parent. We currently only measure wall-clock time, partly because
// that's the only thing that Go provides us, partly because we're interested
// in I/O time. So collecting timing data for things where you block for a long
// timers might be useless for what you're doing.
//
// How to use this in practice:
//  allTimers := New()
//  t = allTimers
//  t = t.Start("foo")
//  foo()
//  t = t.Handover("bar")
//  bar(t)
//  t = t.Stop()
//  func bar(t timers.Timer) {
//       t = t.Start("a")
//       a()
//       t = t.Handover("b")
//       b()
//  }
//
//  This will create a structue with timers as:
//  foo
//  bar
//  bar.a
//  bar.b
type Timer struct {
	children map[string]*Timer
	parent *Timer
	running bool

	count int			// not updated until stopped.
	sw stopwatch.Stopwatch

	max time.Duration
	min time.Duration
}

// Allocate a new timer.
func New() *Timer {
	return &Timer{ min: 1 << 63 - 1 }
}

func (t *Timer)getChild(name string) *Timer {
	if t.children == nil {
		t.children = make(map[string]*Timer)
	}
	n, exists := t.children[name]
	if !exists {
		n = New()
		t.children[name] = n
		n.parent = t
	}
	return n
}

// Create a new timer as a child to this timer and start it.
func (t *Timer)Start(name string) *Timer {
	n := t.getChild(name)
	n.sw.Start()
	n.running = true
	return n
}

// Create a new timer as a child to the parent of this timer and start it.
func (t *Timer)Handover(name string) *Timer {
	n := t.parent.getChild(name)
	t.accumulate(t.sw.Handover(&n.sw))
	n.running = true
	return n
}

// Stop the timer.
func (t *Timer)Stop() {
	t.accumulate(t.sw.Stop())
}

func (t *Timer)accumulate(d time.Duration) {
	t.running = false
	t.count++
	if d > t.max {
		t.max = d
	}
	if d < t.min {
		t.min = d
	}
}

// Callback for the Foreach function.
type ForeachFunc func(name []string, total, avg, max, min time.Duration, count int)

// Iterate over all timers that are the children on this timer and
// call the callback function for each non-zero timer.
func (t Timer)Foreach(f ForeachFunc) {
	t.foreach([]string{}, f)
}

func (t Timer)foreach(name []string, f ForeachFunc) {
	if t.children != nil {
		for k, v := range t.children {
			v.foreach(append(name, k), f)
		}
	}
	if t.count == 0 {
		return
	}
	sw := t.sw
	if t.running {
		sw = t.sw.Snapshot()
	}
	f(name, sw.Duration(), sw.Duration() / time.Duration(t.count), t.max, t.min, t.count)
}
