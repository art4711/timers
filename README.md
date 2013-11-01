## Timers ##

Timers is a simple framework for accumulating timing data in
applications in a hierarchical way to keep track of how much time gets
spent in various parts of the code.

How to use this in practice:

    allTimers := New()
    t = allTimers
    t = t.Start("foo")
    foo()
    t = t.Handover("bar")
    bar(t)
    t = t.Stop()
    func bar(t timers.Timer) {
         t = t.Start("a")
         a()
         t = t.Handover("b")
         b()
    }
    
This will create a structue with timers as:
foo
bar
bar.a
bar.b
