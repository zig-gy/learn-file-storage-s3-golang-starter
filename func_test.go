package main

import "testing"

func TestAspectRatio(t *testing.T) {
	expected := [3]string{"16:9", "9:16", "other"}
	width := [3]int{1920, 1080, 100}
	height := [3]int{1080, 1920, 100}
	for i := range expected {
		actual := aspectRatioFromHeightWidth(height[i], width[i])
		if actual != expected[i] {
			t.Errorf("expected, %v, does not match actual, %v, aspect ratio", expected[i], actual)
		}
	}
}
