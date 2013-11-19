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
    <div id="timercanvas"></div>
    <div id="timerdata"></div>

    <script type="text/javascript">
	var col_level = [
		d3.interpolateRgb(d3.rgb(0, 127, 0), d3.rgb(0, 255, 0)),
		d3.interpolateRgb(d3.rgb(127, 0, 0), d3.rgb(255, 0, 0)),
		d3.interpolateRgb(d3.rgb(0, 0, 127), d3.rgb(0, 0, 255)),
		d3.interpolateRgb(d3.rgb(127, 127, 0), d3.rgb(255, 255, 0)),
		d3.interpolateRgb(d3.rgb(0, 127, 127), d3.rgb(0, 255, 255))
	]

	var w = 960;
	var h = 500;

	var svg = d3.select("#timercanvas").append("svg").attr("width", w).attr("height", h);

	var visible = [ "top" ];

	var part = d3.layout.partition()
		.value(function (d) { return d.Cnt.Avg; })
		.children(function (d) { return d.Children });

	var x = d3.scale.linear().range([0, w]);
	var y = d3.scale.linear().range([0, h]);

	function reload() {
		d3.json("` + timers_url + `", function(error, data) {
			var rect = svg.selectAll("rect").data(part.nodes(data));

			rect.enter().append("rect");
			
			rect.attr("x", function(d) { return x(d.x); })
			 .attr("y", function(d) { return y(d.y); })
			 .attr("width", function(d) { return x(d.dx); })
			 .attr("height", function(d) { return y(d.dy); })
			 .attr("fill", function(d) { return col_level[d.depth](d.x); })
			 .on("click", click);

			function click(d) {
				x.domain([d.x, d.x + d.dx]);
				y.domain([d.y, 1]).range([d.y ? 20 : 0, h]);

				rect.transition()
				 .duration(750)
				 .attr("x", function(d) { return x(d.x); })
				 .attr("y", function(d) { return y(d.y); })
			 	 .attr("width", function(d) { return x(d.x + d.dx) - x(d.x); })
			 	 .attr("height", function(d) { return y(d.y + d.dy) - y(d.y); });
			}
		});
	}
	d3.select("#reload").on("click", reload);
	reload();
    </script>
  </body>
</html>`
	w.Write([]byte(d))
}
