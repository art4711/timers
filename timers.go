// Copyright 2013 Artur Grabowski. All rights reserved.
// Use of this source code is governed by a ISC-style
// license that can be found in the LICENSE file.
package timers

import (
	"github.com/art4711/stopwatch"
	"time"
	"sync"
	"sync/atomic"
	"runtime"
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
//  e = allTimers.Start("foo")
//  foo()
//  e = e.Handover("bar")
//  bar(t)
//  e.Stop()
//  func bar(e *timers.Event) {
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
// Timers are only partially concurrency safe. Timers will be allocated safely, but
// a timer being stopped while foreach is running can give inconsistent results.
// This is worth the tradeoff of not doing constant locking.
type Timer struct {
	children map[string]*Timer
	children_mtx sync.Mutex
	parent *Timer

	memstats bool
	cnt Counts
}

// Contains the total counts of all the events
type Counts struct {
	Count int64			// not updated until stopped.

	Tot time.Duration
	Max time.Duration
	Min time.Duration

	Avg time.Duration		// only updated in ForEach

	BytesAlloc uint64
	NumGC uint32	
}


type Event struct {
	timer *Timer
	sw stopwatch.Stopwatch

	bytesAlloc uint64
	numGC uint32
}

// Allocate a new timer.
func New() *Timer {
	return &Timer{ cnt: Counts{ Min: 1 << 63 - 1 } }
}

// Allocate a new timer that also records memory allocation stats.
func NewMemStats() *Timer {
	t := New()
	t.memstats = true
	return t
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
		n.memstats = t.memstats
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

func (e *Event)startMemStats() {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	e.bytesAlloc = ms.TotalAlloc
	e.numGC = ms.NumGC
}

// Start measuring an event. 
func (t *Timer)Start(name string) *Event {
	e := Event{ timer: t.getChild(name) }
	if t.memstats {
		e.startMemStats()
	}
	e.sw.Start()
	return &e
}

func (e *Event)Start(name string) *Event {
	return e.timer.Start(name)
}

// Create a new event as a child to the parent of this events timer and start it.
func (e *Event)Handover(name string) *Event {
	ne := Event{ timer: e.timer.parent.getChild(name) }
	e.sw.Handover(&ne.sw)
	if ne.timer.memstats {
		ne.startMemStats()
	}
	e.accumulate()
	return &ne
}

// Stop the event.
func (e *Event)Stop() {
	e.sw.Stop()
	e.accumulate()
}

func (e *Event)accumulate() {
	t := e.timer
	d := int64(e.sw.Duration())
	atomic.AddInt64((*int64)(&t.cnt.Tot), d)
	atomic.AddInt64((*int64)(&t.cnt.Count), 1)

	max := int64(t.cnt.Max)
	for d > max && atomic.CompareAndSwapInt64((*int64)(&t.cnt.Max), max, d) {
		max = int64(t.cnt.Max)
	}

	min := int64(t.cnt.Min)
	for d < min && atomic.CompareAndSwapInt64((*int64)(&t.cnt.Min), min, d) {
		min = int64(t.cnt.Min)
	}

	if t.memstats {
		var ms runtime.MemStats
		runtime.ReadMemStats(&ms)
		atomic.AddUint64(&t.cnt.BytesAlloc, ms.TotalAlloc - e.bytesAlloc)
		atomic.AddUint32(&t.cnt.NumGC, ms.NumGC - e.numGC)
	}
}

// Callback for the Foreach function.
type ForeachFunc func(name []string, cnt *Counts)

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

	if t.cnt.Count == 0 {
		return
	}
	t.cnt.Avg = t.cnt.Tot / time.Duration(t.cnt.Count)
	f(name, &t.cnt)
}
