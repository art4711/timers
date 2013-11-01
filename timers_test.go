// Copyright 2013 Artur Grabowski. All rights reserved.
// Use of this source code is governed by a ISC-style
// license that can be found in the LICENSE file.
package timers_test

import (
	"github.com/art4711/timers"
	"testing"
	"time"
)

func af(t *testing.T, h time.Duration, e float64, name string) {
	if h.Seconds() < e || h.Seconds() > e + 0.1 {
		t.Errorf("bad %v: %v !~= %v", name, h, e)
	}
}

func ai(t *testing.T, h, e int, name string) {
	if h != e {
		t.Errorf("bad %v: %v !~= %v", name, h, e)
	}
}

func w1() {
	d, _ := time.ParseDuration("1s")
	select {
	case <- time.After(d):
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
	w1()
	t1.Stop()
	count := 0
	tm.Foreach(func (na []string, tot, a, mx, mi time.Duration, c int) {
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
	w1()
	t2 := t1.Start("first")
	w1()
	t2.Stop()
	t1.Stop()
	count := 0
	tm.Foreach(func (na []string, tot, a, mx, mi time.Duration, c int) {
		n := nam(na)
		switch n {
		case "first":
			af(t, tot, 2.0, n + ".tot")
		case "first.first":
			af(t, tot, 1.0, n + ".tot")
		default:
			t.Errorf("bad name: %v", n)
		}
		af(t, a, tot.Seconds(), n + ".a")
		af(t, mx, tot.Seconds(), n + ".mx")
		af(t, mi, tot.Seconds(), n + ".mi")
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
	w1()
	t1.Stop()

	t1 = tm.Start("1")
	w1()
	t1.Stop()

	count := 0
	tm.Foreach(func (na []string, tot, a, mx, mi time.Duration, c int) {
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
		af(t, mx, 1.0, "mx")
		af(t, tot, 2.0, "tot")
		af(t, a, 1.0, "avg")
		count++
	})
	if count != 1 {
		t.Errorf("too many timers %v", count)
	}
}

func TestNestedHandover(t *testing.T) {
	tm := timers.New()
	t1 := tm.Start("1")
	w1()
	t2 := t1.Start("1")
	w1()
	t2 = t2.Handover("2")
	w1()
	t2 = t2.Handover("1")
	w1()
	t2.Stop()
	t1.Stop()
	count := 0
	tm.Foreach(func (na []string, tot, a, mx, mi time.Duration, c int) {
		n := nam(na)
		switch n {
		case "1":
			af(t, tot, 4.0, n + ".tot")
			af(t, mx, 4.0, n + ".mx")
			af(t, mi, 4.0, n + ".mi")
			ai(t, c, 1, n + ".c")
		case "1.1":
			af(t, tot, 2.0, n + ".tot")
			af(t, mx, 1.0, n + ".mx")
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
