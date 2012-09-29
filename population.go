package main

import (
	"math/rand"
	"sort"
)

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


