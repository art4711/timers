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
    <style type="text/css">

.chart {
  display: block;
  font-size: 11px;
}

rect {
  stroke: #eee;
  cursor: pointer;
  fill: steelblue;
  fill-opacity: .8;
}

text {
  pointer-events: none;
}
    </style>
  </head>
  <body>
    <a href="#" id="reload">reload</a>
    <div id="timercanvas" class="chart"></div>
    <div id="timerdata"></div>


    <script type="text/javascript">
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
			var g = svg.selectAll("g").data(part.nodes(data));

			g.enter().append("g");

			g.attr("transform", function(d) { return "translate(" + x(d.y) + "," + y(d.x) + ")"; })
			 .on("click", click);
			var kx = w / data.dx,
			    ky = h / 1;

			g.append("rect")
			 .attr("width", data.dy * kx)
			 .attr("height", function(d) { return d.dx * ky; })
			 .attr("class", function(d) { return d.children ? "parent" : "child"; });

			g.append("text")
			 .attr("transform", transform)
			 .attr("dy", ".35em")
			 .style("opacity", function(d) { return d.dx * ky > 12 ? 1 : 0; })
			 .text(function(d) { return d.Name; })

			d3.select(window).on("click", function() { click(data); })

			function click(d) {
				kx = (d.y ? w - 40 : w) / (1 - d.y);
				ky = h / d.dx;
				x.domain([d.y, 1]).range([d.y ? 40 : 0, w]);
				y.domain([d.x, d.x + d.dx]);

				var t = g.transition()
				 .duration(d3.event.altKey ? 7500 : 750)
				 .attr("transform", function(d) { return "translate(" + x(d.y) + "," + y(d.x) + ")"; });

				t.select("rect")
				 .attr("width", d.dy * kx)
				 .attr("height", function(d) { return d.dx * ky; });

				t.select("text")
				 .attr("transform", transform)
				 .style("opacity", function(d) { return d.dx * ky > 12 ? 1 : 0; });

				d3.event.stopPropagation();
			}
			function transform(d) {
				return "translate(8, " + d.dx * ky / 2 + ")";
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
