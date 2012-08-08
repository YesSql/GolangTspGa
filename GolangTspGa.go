package main

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"runtime"
	"time"
	"net/http"
	"flag"
	"bytes"
	"strconv"
)

type location struct {
	X float64
	Y float64
}

type tour struct {
	locations   []location
	order       []int
	count       int
	tour_length float64
}

func (t tour) swapMutation() float64 {
	swap1 := (int)(rand.Float64() * (float64)(t.count))
	swap2 := (int)(rand.Float64() * (float64)(t.count))

	t.order[swap1], t.order[swap2] = t.order[swap2], t.order[swap1]
	return tour_length(t)
}

func (t tour) insertionMutation() float64 {
	element := (int)(rand.Float64() * (float64)(t.count))
	insert_after := (int)(rand.Float64() * (float64)(t.count))

	if element != insert_after {

		if element > insert_after {
			element, insert_after = insert_after, element
		}

		element_copy := t.order[element]

		for i := element; i < insert_after; i++ {
			t.order[i] = t.order[i+1]
		}

		t.order[insert_after] = element_copy

		t.tour_length = tour_length(t)
	}

	return t.tour_length
}

func (t tour) branchSwapMutation() float64 {

	start := (int)(rand.Float64() * (float64)(t.count))
	stop := (int)(rand.Float64() * (float64)(t.count))

	if start > stop {
		start, stop = stop, start
	}

	if stop-start > 1 {
		swapper := make([]int, stop-start+1)
		swapper_count := 0
		for i := stop; i > (start - 1); i-- {
			swapper[swapper_count] = t.order[i]
			swapper_count++
		}
		swapper_count = 0
		for i := start; i < (stop + 1); i++ {
			t.order[i] = swapper[swapper_count]
			swapper_count++
		}
	}
	return tour_length(t)
}

func (p population) Len() int {
	return len(p.tours)
}

func (p population) Less(i, j int) bool {
	return p.tours[i].tour_length < p.tours[j].tour_length
}

func (p population) Swap(i, j int) {
	p.tours[i], p.tours[j] = p.tours[j], p.tours[i]
}

type population struct {
	tours     []tour
	best_tour tour
}

func create_population(starting_tour tour, size int) population {
	var ret population
	ret.tours = make([]tour, size)
	ret.best_tour = copy_tour(starting_tour)
	for i := range ret.tours {
		ret.tours[i] = copy_tour(starting_tour)
	}
	return ret
}

func copy_tour(tour_to_copy tour) tour {
	var ret tour
	ret.count = tour_to_copy.count
	ret.locations = make([]location, ret.count)
	ret.order = make([]int, ret.count)
	for i := range ret.order {
		ret.locations[i] = tour_to_copy.locations[i]
		ret.order[i] = tour_to_copy.order[i]
	}
	ret.tour_length = tour_to_copy.tour_length
	return ret
}

func create_tour(count int, bounds float64) tour {
	var ret tour

	ret.count = count
	ret.locations = make([]location, count)
	ret.order = make([]int, count)
	for i := 0; i < count; i++ {
		ret.locations[i] = location{rand.Float64() * bounds, rand.Float64() * bounds}
		ret.order[i] = i
	}

	ret.tour_length = tour_length(ret)

	return ret
}

func distance(location1 location, location2 location) float64 {
	return math.Max(math.Abs(location1.X-location2.X), math.Abs(location1.Y-location2.Y))
}

func tour_length(tour_in tour) float64 {

	length := 0.0
	for i := 1; i < len(tour_in.order); i++ {
		length += distance(tour_in.locations[tour_in.order[i-1]], tour_in.locations[tour_in.order[i]])
	}
	length += distance(tour_in.locations[tour_in.order[0]], tour_in.locations[tour_in.order[len(tour_in.order)-1]])
	return length
}

func iterate(solutions population) {
	
	solutions.best_tour = solutions.tours[0]

	deadSolutions := 0

	for i := 1; i < len(solutions.tours); i++ {
		probability := (float64)(i) / (float64)(len(solutions.tours))
		if rand.Float64() < probability {
			solutions.tours = append(solutions.tours[:i-deadSolutions], solutions.tours[i+1-deadSolutions:]...)
			deadSolutions++
		}
	}

	survivors := len(solutions.tours)

	for i := 0; i < deadSolutions; i++ {
		mom := (int)(rand.Float64() * (float64)(survivors))
		baby := copy_tour(solutions.tours[mom])

		var newLength float64
		switch (int)(rand.Float64() * (float64(4))) {
		case 0:
			newLength = baby.insertionMutation()
		case 1:
			newLength = baby.swapMutation()
		case 2:
			newLength = baby.branchSwapMutation()
		case 3:
			newLength = baby.branchSwapMutation()
		}
		baby.tour_length = newLength
		solutions.tours = append(solutions.tours, baby)
	}
	sort.Sort(solutions)
}

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
			//fmt.Printf("%s\tGen: %d\tNew Best: %.3f\n", name, i, current_best)
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


func print_update()

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

func serv(w http.ResponseWriter, req *http.Request) {
	
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
	
	var buffer bytes.Buffer
	
	buffer.WriteString("<html><body><table border = '1'><tr>")
	for i := range lengths {
		buffer.WriteString(fmt.Sprintf("<th> Population %d </th>", i+1))
	}
	buffer.WriteString("</tr><tr>")
	for i := range lengths {
		buffer.WriteString(fmt.Sprintf("<td> %.3f </td>", tour_length(status_map[fmt.Sprintf("Population: %d", i+1)].best_tour)))
	}
	buffer.WriteString("</tr><tr>")
	for i :=0; i <len(lengths); i++ {
		buffer.WriteString(fmt.Sprintf("<td><canvas id = 'graph_%d' width = '400' height = '400'> No support </canvas></td>",i))
	}
	buffer.WriteString("</tr></table>")
	buffer.WriteString("<form action = '.' type = 'post'><input type= 'submit' value = 'Refresh'/></form>")
	buffer.WriteString("<form action = '.' type = 'post'><input type= 'submit' value = 'Please Stop'/><input type = 'hidden' name = 'x' /></form>")
	buffer.WriteString("</body>")
	for i :=0; i <len(lengths); i++ {
		
		the_tour := status_map[fmt.Sprintf("Population: %d", i+1)].best_tour
		
		buffer.WriteString(fmt.Sprintf("<script> var graph = document.getElementById('graph_%d');\n",i))
		buffer.WriteString("var context = graph.getContext('2d');\n")
		buffer.WriteString("context.fillStyle = 'blue';\n")
		for j := range the_tour.order {
			buffer.WriteString(fmt.Sprintf("context.fillRect(%d,%d,%d,%d);\n", (int)(the_tour.locations[the_tour.order[j]].X-2), (int)(the_tour.locations[the_tour.order[j]].Y-2), 4,4))
		}
		
		
		for j := 0; j< len(the_tour.order) -1; j++ {
			buffer.WriteString("context.beginPath();\n")
			x := (int)(the_tour.locations[the_tour.order[j]].X)
			y := (int)(the_tour.locations[the_tour.order[j]].Y)
			buffer.WriteString(fmt.Sprintf("context.moveTo(%d,%d);\n", x,y))
			
			x = (int)(the_tour.locations[the_tour.order[j+1]].X)
			y = (int)(the_tour.locations[the_tour.order[j+1]].Y)
			buffer.WriteString(fmt.Sprintf("context.lineTo(%d,%d);\n",x,y))
			buffer.WriteString("context.stroke();\n")
		}
		
		
		buffer.WriteString("</script>")
	}
	
	buffer.WriteString("</html>")
	
	
	w.Write([]byte(buffer.String()))
	
	req.ParseForm()
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
