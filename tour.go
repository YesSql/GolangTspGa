package main

import ("math/rand")

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

func tour_length(tour_in tour) float64 {

	length := 0.0
	for i := 1; i < len(tour_in.order); i++ {
		length += distance(tour_in.locations[tour_in.order[i-1]], tour_in.locations[tour_in.order[i]])
	}
	length += distance(tour_in.locations[tour_in.order[0]], tour_in.locations[tour_in.order[len(tour_in.order)-1]])
	return length
}
