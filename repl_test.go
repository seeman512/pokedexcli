package main

import "testing"

func TestCleanInput(t *testing.T) {
	cases := []struct {
		input    string
		expected []string
	}{
		{
			input:    " hello World",
			expected: []string{"hello", "world"},
		},
		{
			input:    " Test1, 200 TEST100 ff",
			expected: []string{"test1,", "200", "test100", "ff"},
		},
	}

	for _, c := range cases {
		actual := cleanInput(c.input)
		for i, word := range actual {
			if word != c.expected[i] {
				t.Errorf("Expecting '%v' got '%v'", c.expected[i], word)
			}
		}
	}
}
