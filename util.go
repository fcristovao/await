package main

import "sort"

// Returns the string sorted, with no duplicates
func unique(strings []string) []string {
	sort.Strings(strings)
	var uniques []string
	current, next := 0, 1
	if len(strings) > 0 {
		uniques = append(uniques, strings[current])
		for next < len(strings) {
			if strings[current] != strings[next] {
				uniques = append(uniques, strings[next])
				current = next
			}
			next = next + 1
		}
	}
	return uniques
}

// It assumes both inputs are already sorted and distinct
// Returns not only if it found all, but also which ones were missing
func containsAll(haystack []string, needles []string) (bool, []string) {
	switch {
	case len(needles) == 0:
		return true, needles
	case len(haystack) == 0:
		return false, needles
	case haystack[0] == needles[0]:
		return containsAll(haystack[1:], needles[1:])
	default:
		return containsAll(haystack[1:], needles)
	}
}
