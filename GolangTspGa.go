package main

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
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
	return length
}

func iterate(solutions population) population {
	sort.Sort(solutions)
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

	return solutions
}

func main() {
	rand.Seed(100)
	curr_tour := create_tour(50, 100)

	solutions := create_population(curr_tour, 30)

	for i := 0; i < 1000000; i++ {
		solutions = iterate(solutions)
		if i%80000 == 0 {
			fmt.Println("")
		}
		if i%1000 == 0 {
			fmt.Print("*")
		}
	}

	sort.Sort(solutions)

	for j := 0; j < 1; j++ {
		fmt.Println()
		for _, value := range solutions.tours[j].order {
			fmt.Printf("%0.3f\t%0.3f\n", solutions.tours[j].locations[value].X, solutions.tours[j].locations[value].Y)
		}
		fmt.Println("Distance ", tour_length(solutions.tours[j]))
	}
	fmt.Printf("Tour length: %0.3f\n", solutions.tours[0].tour_length)
}
