// Copyright 2013 Artur Grabowski. All rights reserved.
// Use of this source code is governed by a ISC-style
// license that can be found in the LICENSE file.

package timers

import (
	"net/http"
	"encoding/json"
)

// StdJSON is the default layout for outputing the timer data as JSON.
type StdJSON struct {
	Name string
	Cnt Counts
	Children []*StdJSON
}

func (sj *StdJSON) Insert(name []string, counts *Counts) {
	if len(name) == 0 {
		sj.Cnt = *counts
		return
	}

	// This would be so much easier with a map, but then the json is a pain in the ass to handle.
	var c *StdJSON
	for _, v := range sj.Children {
		if v.Name == name[0] {
			c = v
			break
		}
	}
	if c == nil {
		c = &StdJSON{}
		c.Name = name[0]
		sj.Children = append(sj.Children, c)
	}
	c.Insert(name[1:], counts)
}

// Standard http handler for returning a Timer with JSON.
func (t Timer) JSONHandler(w http.ResponseWriter, req *http.Request) {
	out := StdJSON{ Name: "top" }

	t.Foreach(func (name []string, counts *Counts) {
		out.Insert(name, counts)
	})
	json, _ := json.MarshalIndent(out, "", "    ")
	w.Write(json)	
}

func JSONHandlerGraph(w http.ResponseWriter, req *http.Request, timers_url string) {
d :=`<html>
  <head>
    <script src="http://d3js.org/d3.v3.min.js" charset="utf-8"></script>
  </head>
  <body>
    <a href="#" id="reload">reload</a>
    <div id="timerdata"></div>
    <div id="timercanvas"></div>

    <script type="text/javascript">
	var cols = ['purple', 'red', 'orange', 'yellow', 'lime', 'green', 'blue', 'navy', 'black'];

	var w = 800;
	var h = 300;

	var svg = d3.select("#timercanvas").append("svg").attr("width", w).attr("height", h);

	function reload() {
		function add_timer(timer, addto, idname) {
			idname += "_" + timer.Name;

			var ul = d3.select("#" + idname)
			var li = d3.select("#" + idname + "_data")
			if (ul.empty()) {
				addto.append("ul").attr("id", idname)
				ul = d3.select("#" + idname)
				ul.append("li").attr("id", idname + "_data")
				li = d3.select("#" + idname + "_data")
			}
			li.text(timer.Name + ", counts: " + timer.Cnt.Count + ", avg: " + timer.Cnt.Avg)
			if (timer.Children) {
				for (var i = 0; i < timer.Children.length; i++) {
					add_timer(timer.Children[i], ul, idname);
				}
			}
		}
		function redraw(timer) {
			var gdata = [ 1 ];	// Force update.
			var g = svg.selectAll("g").data(gdata);

			g.enter().append("g");

			var rects = g.selectAll("rect").data(timer.Children);

			rects.enter().append("rect");

			var xoff = 0;
			rects.attr("x", function(d, i) {
				r = xoff;
				xoff += d.Cnt.Avg;
				return r;
			})
			.attr("y", 0).attr("height", 50)
			.attr("width", function(d, i) {
				return d.Cnt.Avg;
			})
			.attr("fill", function(d, i) {
				return cols[i % cols.length];
			})
			.on("click", function(d, i) {
				console.log("clicked: " + d.Name);
			});

			g.attr("transform", function(d, i) {
				return "scale(" + w / xoff + ", 1)";
			});

		}
	        d3.json("` + timers_url + `", function(error, data) {
			add_timer(data, d3.select("#timerdata"), "timerdata")
			redraw(data)
                });
	}
	d3.select("#reload").on("click", reload);
	reload();
    </script>
  </body>
</html>`
	w.Write([]byte(d))
}
