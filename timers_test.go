// Copyright 2013 Artur Grabowski. All rights reserved.
// Use of this source code is governed by a ISC-style
// license that can be found in the LICENSE file.
package timers_test

import (
	"github.com/art4711/timers"
	"testing"
	"time"
)

const q = time.Millisecond * 200
const qd = 5.0

func af(t *testing.T, h time.Duration, e float64, name string) {
	e /= qd
	aff(t, h, e, name)
}

func aff(t *testing.T, h time.Duration, e float64, name string) {
	if h.Seconds() < e || h.Seconds() > e + 0.02 {
		t.Errorf("bad %v: %v !~= %v", name, h, e)
	}
}

func ai(t *testing.T, h, e int64, name string) {
	if h != e {
		t.Errorf("bad %v: %v !~= %v", name, h, e)
	}
}

func nam(n []string) string {
	var r string
	for k, v := range n {
		if k > 0 {
			r += "." + v
		} else {
			r = v
		}
	}
	return r
}

func TestBasic(t *testing.T) {
	tm := timers.New()
	t1 := tm.Start("first")
	time.Sleep(q)
	t1.Stop()
	count := 0
	tm.Foreach(func (na []string, tot, a, mx, mi time.Duration, c int64) {
		n := nam(na)
		if n != "first" {
			t.Errorf("bad name: %v", n)
		}
		af(t, tot, 1.0, "tot")
		af(t, a, 1.0, "a")
		af(t, mx, 1.0, "mx")
		af(t, mi, 1.0, "mi")
		ai(t, c, 1, "c")
		count++
	})
	if count != 1 {
		t.Errorf("too many timers %v", count)
	}
}

func TestNested(t *testing.T) {
	tm := timers.New()
	t1 := tm.Start("first")
	time.Sleep(q)
	t2 := t1.Start("first")
	time.Sleep(q)
	t2.Stop()
	t1.Stop()
	count := 0
	tm.Foreach(func (na []string, tot, a, mx, mi time.Duration, c int64) {
		n := nam(na)
		switch n {
		case "first":
			af(t, tot, 2.0, n + ".tot")
		case "first.first":
			af(t, tot, 1.0, n + ".tot")
		default:
			t.Errorf("bad name: %v", n)
		}
		aff(t, a, tot.Seconds(), n + ".a")
		aff(t, mx, tot.Seconds(), n + ".mx")
		aff(t, mi, tot.Seconds(), n + ".mi")
		ai(t, c, 1, n + ".c")
		count++
	})
	if count != 2 {
		t.Errorf("too many timers %v", count)
	}
}

func TestRepeat(t *testing.T) {
	tm := timers.New()
	t1 := tm.Start("1")
	time.Sleep(q)
	t1.Stop()

	t1 = tm.Start("1")
	time.Sleep(q * 2)
	t1.Stop()

	count := 0
	tm.Foreach(func (na []string, tot, a, mx, mi time.Duration, c int64) {
		n := nam(na)
		switch n {
		case "1":
			if c != 2 {
				t.Errorf("bad count: %v", c)
			}
		default:
			t.Errorf("bad name: %v", n)
		}
		af(t, mi, 1.0, "mi")
		af(t, mx, 2.0, "mx")
		af(t, tot, 3.0, "tot")
		af(t, a, 1.5, "avg")
		count++
	})
	if count != 1 {
		t.Errorf("too many timers %v", count)
	}
}

func TestNestedHandover(t *testing.T) {
	tm := timers.New()
	t1 := tm.Start("1")
	time.Sleep(q)
	t2 := t1.Start("1")
	time.Sleep(q)
	t2 = t2.Handover("2")
	time.Sleep(q)
	t2 = t2.Handover("1")
	time.Sleep(q * 2)
	t2.Stop()
	t1.Stop()
	count := 0
	tm.Foreach(func (na []string, tot, a, mx, mi time.Duration, c int64) {
		n := nam(na)
		switch n {
		case "1":
			af(t, tot, 5.0, n + ".tot")
			af(t, mx, 5.0, n + ".mx")
			af(t, mi, 5.0, n + ".mi")
			ai(t, c, 1, n + ".c")
		case "1.1":
			af(t, tot, 3.0, n + ".tot")
			af(t, mx, 2.0, n + ".mx")
			af(t, mi, 1.0, n + ".mi")
			ai(t, c, 2, n + ".c")
		case "1.2":
			af(t, tot, 1.0, n + ".tot")
			af(t, mx, 1.0, n + ".mx")
			af(t, mi, 1.0, n + ".mi")
			ai(t, c, 1, n + ".c")
		default:
			t.Errorf("bad name: %v", n)
		}
		count++
	})
	if count != 3 {
		t.Errorf("too many timers %v", count)
	}
}

func tfun(t2 *timers.Event) {
	t2 = t2.Handover("2")
	time.Sleep(q)
	t2 = t2.Handover("1")
	time.Sleep(q * 2)
	t2.Stop()	
}

func TestNestedHandoverFunction(t *testing.T) {
	tm := timers.New()
	t1 := tm.Start("1")
	time.Sleep(q)
	t2 := t1.Start("1")
	time.Sleep(q)
	tfun(t2)
	t1.Stop()
	count := 0
	tm.Foreach(func (na []string, tot, a, mx, mi time.Duration, c int64) {
		n := nam(na)
		switch n {
		case "1":
			af(t, tot, 5.0, n + ".tot")
			af(t, mx, 5.0, n + ".mx")
			af(t, mi, 5.0, n + ".mi")
			ai(t, c, 1, n + ".c")
		case "1.1":
			af(t, tot, 3.0, n + ".tot")
			af(t, mx, 2.0, n + ".mx")
			af(t, mi, 1.0, n + ".mi")
			ai(t, c, 2, n + ".c")
		case "1.2":
			af(t, tot, 1.0, n + ".tot")
			af(t, mx, 1.0, n + ".mx")
			af(t, mi, 1.0, n + ".mi")
			ai(t, c, 1, n + ".c")
		default:
			t.Errorf("bad name: %v", n)
		}
		count++
	})
	if count != 3 {
		t.Errorf("too many timers %v", count)
	}
}

func BenchmarkBasic(b *testing.B) {
	for i:= 0; i < b.N; i++ {
		tm := timers.New()
		t1 := tm.Start("1")
		t2 := t1.Start("1")
		t2 = t2.Handover("2")
		t2 = t2.Handover("1")
		t2.Stop()
		t1.Handover("2")
		t1.Stop()
	}
}
