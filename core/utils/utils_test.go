package utils

import "testing"

func TestRemoveDuplicates(t *testing.T) {
	testSlice := []string{"aaaa", "bbb", "aaaa", "bbb"}
	expectedSlice := []string{"aaaa", "bbb"}
	testSlice = RemoveDuplicatesUnordered(testSlice)

	for _, x := range testSlice {
		for i, expectedX := range expectedSlice {
			if x == expectedX {
				break
			}
			if i == len(expectedSlice) {
				t.Fail()
			}
		}
	}

}
