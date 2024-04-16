package main

import (
	"cmp"
	"slices"
)

func Unique[S ~[]E, E cmp.Ordered](s S) S {
	slices.Sort(s)
	compact := slices.Compact(s)
	unique := slices.Clip(compact)
	return unique
}
