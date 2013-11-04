// Copyright 2013 Artur Grabowski. All rights reserved.
// Use of this source code is governed by a ISC-style
// license that can be found in the LICENSE file.
package timers

import (
	"github.com/art4711/stopwatch"
	"time"
	"sync"
	"sync/atomic"
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
//
// Timers are only partiall concurrency safe. Timers will be allocated safely, but
// a timer being stopped while foreach is running can give inconsistent results.
// This is worth the tradeoff of not doing constant locking.
type Timer struct {
	children map[string]*Timer
	children_mtx sync.Mutex
	parent *Timer

	count int64			// not updated until stopped.

	tot time.Duration
	max time.Duration
	min time.Duration
}

type event struct {
	timer *Timer
	sw stopwatch.Stopwatch
}

// Allocate a new timer.
func New() *Timer {
	return &Timer{ min: 1 << 63 - 1 }
}

func (t *Timer)getChild(name string) *Timer {
	if t.children == nil {
		t.children_mtx.Lock()
		// recheck with lock
		if t.children == nil {
			t.children = make(map[string]*Timer)
		}
		t.children_mtx.Unlock()
	}
	n, exists := t.children[name]
	if !exists {
		n = New()
		n.parent = t
		t.children_mtx.Lock()
		// recheck with lock
		nn, ex2 := t.children[name]
		if !ex2 {
			t.children[name] = n
		} else {
			n = nn
		}
		t.children_mtx.Unlock()
	}
	return n
}

// Start measuring an event. 
func (t *Timer)Start(name string) *event {
	e := event{ timer: t.getChild(name) }
	e.sw.Start()
	return &e
}

func (e *event)Start(name string) *event {
	return e.timer.Start(name)
}

// Create a new event as a child to the parent of this events timer and start it.
func (e *event)Handover(name string) *event {
	ne := event{ timer: e.timer.parent.getChild(name) }
	e.sw.Handover(&ne.sw)
	e.accumulate()
	return &ne
}

// Stop the event.
func (e *event)Stop() {
	e.sw.Stop()
	e.accumulate()
}

func (e *event)accumulate() {
	t := e.timer
	d := int64(e.sw.Duration())
	atomic.AddInt64((*int64)(&t.tot), d)
	atomic.AddInt64((*int64)(&t.count), 1)

	max := int64(t.max)
	for d > max && atomic.CompareAndSwapInt64((*int64)(&t.max), max, d) {
		max = int64(t.max)
	}

	min := int64(t.min)
	for d < min && atomic.CompareAndSwapInt64((*int64)(&t.min), min, d) {
		min = int64(t.min)
	}
}

// Callback for the Foreach function.
type ForeachFunc func(name []string, total, avg, max, min time.Duration, count int64)

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
	f(name, t.tot, t.tot / time.Duration(t.count), t.max, t.min, t.count)
}
