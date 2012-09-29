package main

import ("math")

type location struct {
	X float64
	Y float64
}

func distance(location1 location, location2 location) float64 {
	return math.Max(math.Abs(location1.X-location2.X), math.Abs(location1.Y-location2.Y))
}