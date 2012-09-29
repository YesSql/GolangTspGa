package main

import (
	"fmt"
	"math/rand"
	"sort"
	"runtime"
	"time"
	"net/http"
	"flag"
	"bytes"
	"strconv"
)

func iterate_n(solutions population, name string, stop chan bool, c chan population_result) {
	
	sort.Sort(solutions)
	best_tour_length := tour_length(solutions.tours[0])

	result := population_result{ name, best_tour_length, solutions.tours[0]}
	c <- result	
	i := 1
	for ; ; {
		iterate(solutions)
		current_best := tour_length(solutions.tours[0])
		if(current_best < best_tour_length) {
			best_tour_length = current_best
			result.length = current_best
			result.best_tour = solutions.tours[0]
			c <- result	
		}
		
		select {
			case  <- stop:	
				c <- result
				return;
			default:
		}
		i++
	}	
}

type population_result struct {
	name string
	length float64
	best_tour tour
}

var running bool

func setup(w http.ResponseWriter, req *http.Request) (theCount int, theGoRoutineCount int) {
	var buffer bytes.Buffer
	var err error
	req.ParseForm()
	
	for i, valu := range req.Form {
		switch i {
			case "count":
				theCount, err = strconv.Atoi(valu[0])
			case "routines":
				theGoRoutineCount, err = strconv.Atoi(valu[0])
		}
	}	

	if theCount > 0 && err == nil {
		return
	}
	
	buffer.WriteString("<html><body><h1>Solve the Traveling Salesman Problem with Golang and HTML5</h1><form><table><tr><td>locations: </td><td><input name = 'count' type = 'textfield' value = '100'></td></tr>")
	buffer.WriteString("<tr><td>go routines: </td><td><input type='textfield' value = '4' name = 'routines'></td></tr>")
	buffer.WriteString("<tr><td colspan = '2'><input type='submit' value = 'go'></td></tr></table></form></html>")

	w.Write([]byte(buffer.String()))
	
	return
}


func javascript() string {
	 ret := "<script>" +
		"window.updatesActive = true;" +
		"$().ready(init);\n" +
		"function update() {\n" +
		"	if(!window.updatesActive) {return;}\n"+
		"	$.getJSON('./?cmd=update', function(data) { " +  
		"	    for(var i=0; i<data.tours.length; i++) {	" +
		"			var graph = document.getElementById('graph_' + i);\n"+
		"           var context = graph.getContext('2d');\n"+
		"			clearCanvas(context, graph); \n" +
		" 			context.fillStyle = 'blue';\n" +          
		"			var tour = data.tours[i].locations;\n" +
		"			$('#length_'+i).text(data.tours[i].tour_length);"+
		"           for(var j=0; j<tour.length; j++) {\n" +
		"				var x = tour[j][0];\n" +
		"               var y = tour[j][1];\n" +
		"				context.fillRect(x-2,y-2,4,4);\n" +				
		"			}\n" +	
		"			for (var j = 0; j< tour.length-1; j++) { " +
		"				context.beginPath();\n " +
		"				var x = tour[j][0]; \n" +
		"				var y = tour[j][1]; \n" +
		"               context.moveTo(x,y);\n" +		
		"				x = tour[j+1][0];\n "+
		"           	y = tour[j+1][1];\n" +
		"				context.lineTo(x,y);\n"+
		"				context.stroke();\n" +
		"			}\n" +
		"       }\n"+		
		"	});\n "+
		"   "+
		"}" +
		"function clearCanvas(context, canvas) { " +
  		"	context.clearRect(0, 0, canvas.width, canvas.height); "+
  		"	var w = canvas.width;" +
  		"	canvas.width = 1; " +
  		"	canvas.width = w; " +
		"}"+
		"var refreshIntervalId = 1\n" +
		"function init() { " +
		"	refreshIntervalId = window.setInterval(update, 50);\n"	+
		"	$('#stopButton').bind('click', function(){\n window.clearInterval(refreshIntervalId);$.getJSON('./?x=stop');});\n"+
		"}\n"+
		"</script>"
	return ret
}


func sendJSON (w http.ResponseWriter) {
	var buffer bytes.Buffer	
	buffer.WriteString("{\"tours\" : [")
	delimiter := ""
	for i :=0; i <len(lengths); i++ {
		the_tour := status_map[fmt.Sprintf("Population: %d", i+1)].best_tour
		buffer.WriteString(fmt.Sprintf("%s\n { \"tour_length\" : %.3f, \"locations\" :[ ", delimiter, tour_length(the_tour)))
		delimiter2 := ""
		for j := range the_tour.order {
			buffer.WriteString(fmt.Sprintf("%s\n\t[%d,%d]",delimiter2,(int)(the_tour.locations[the_tour.order[j]].X), (int)(the_tour.locations[the_tour.order[j]].Y)))
			delimiter2 = ","
		}
		buffer.WriteString("\n] } ")
		delimiter = ","
	}
	buffer.WriteString("] }")
	w.Write([] byte (buffer.String()))
}

func serv(w http.ResponseWriter, req *http.Request) {

		var buffer bytes.Buffer	
		req.ParseForm()
		for i, v := range req.Form{
			fmt.Println(i,v)
			if i == "cmd" && v[0] == "update" {
				sendJSON(w)		
				return
			}
		}
	
	count := 0
	threads := 0
	
	if !running {
		count, threads = setup(w,req)
		if count == 0 || threads == 0 {	
			return
		} else {
		 	running = true
			go launch(count, threads)
		}
	}
		
    for i:=0 ; i<10; i++ {
		time.Sleep(1)
	}	

	
	buffer.WriteString("<html><head><script src='http://code.jquery.com/jquery-1.8.0.js'></script>")
	buffer.WriteString(javascript())
	buffer.WriteString("</head><body><table border = '1'><tr>")
	for i := range lengths {
		buffer.WriteString(fmt.Sprintf("<th> Population %d </th>", i+1))
	}
	buffer.WriteString("</tr><tr>")
	for i := range lengths {
		buffer.WriteString(fmt.Sprintf("<td id ='length_%d'> %.3f </td>", i, tour_length(status_map[fmt.Sprintf("Population: %d", i+1)].best_tour)))
	}
	buffer.WriteString("</tr><tr>")
	for i :=0; i <len(lengths); i++ {
		buffer.WriteString(fmt.Sprintf("<td><canvas id = 'graph_%d' width = '400' height = '400'> No support </canvas></td>",i))
	}
	buffer.WriteString("</tr></table>")
	buffer.WriteString("<input type= 'submit' value = 'Please Stop' id = 'stopButton'/>")
	buffer.WriteString("</body>")
	for i :=0; i <len(lengths); i++ {
		buffer.WriteString("</script>")
	}
	
	buffer.WriteString("</html>")
	
	
	w.Write([]byte(buffer.String()))
	

	for i, _ := range req.Form{
		if i == "x" {
			curtains <- true
		}
	}	
}

func makeserver() {
	var addr = flag.String("addr", ":1718", "http service address")
	http.Handle("/", http.HandlerFunc(serv))
	http.ListenAndServe(*addr, nil)
}


var lengths []float64
var status_map map[string] population_result

var curtains chan bool

func main() {

	running = false

	go makeserver()	
				
	for ;;{
		time.Sleep(1)
	}
				
}

func launch(count int, cpus int) {
														
	runtime.GOMAXPROCS(cpus)
	
	lengths = make([] float64, cpus)
	status_map = make(map[string] population_result) 
	curtains = make(chan bool)
	for i := range lengths {
		lengths[i] = -1.0
	}
	
	rand.Seed(100)
	curr_tour := create_tour(count, 400)
	
	population_results := make(chan population_result)
	stop := make(chan bool)
	results := make([] population_result, cpus)
	solutions := make([] population, cpus)

	solution_map := map[string] population {}
			
			
	for i := range solutions {
		solutions[i] = create_population(curr_tour, 30)
		name := fmt.Sprintf("Population: %d", i+1)
		solution_map[name] = solutions[i]
		go iterate_n(solutions[i], name, stop, population_results)
	}


	for i:=0 ; i<1.1e9; i++ {
		time.Sleep(1)

		select {
			case curr_result := <- population_results:
				status_map[curr_result.name] = curr_result	
			case <-curtains:
				i = 1.2e9
			default:
		}
	}
	
	running = false
	
	
	for i:=0; i<cpus; i++ {
		stop <- true
	}		

	for count:=0; count<cpus;  {
		select {
			case results[count] = <- population_results:
				count++
			default:
				time.Sleep(1)
		}
	}
	

	best_length := results[0].length
	best_result := results[0]
	
	for i := 1; i<cpus; i++ {
		if results[i].length < best_length {
			best_result = results[i]
			best_length = results[i].length
		}
	}
	
	best_solution := solution_map[best_result.name]


	fmt.Println()
	for _, value := range best_solution.tours[0].order {
		fmt.Printf("%0.3f\t%0.3f\n", best_solution.tours[0].locations[value].X, best_solution.tours[0].locations[value].Y)
	}
	fmt.Println("Distance ", tour_length(best_solution.tours[0]))
	
	fmt.Printf("Tour length: %0.3f\n", best_solution.tours[0].tour_length)
	
	for i:=0; i<cpus; i++ {
		fmt.Printf("%s: %.3f\n", results[i].name, results[i].length)
	}
}
