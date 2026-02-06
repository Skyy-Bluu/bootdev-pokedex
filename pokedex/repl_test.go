package main

import "testing"

func TestCleanInput(t *testing.T) {

	cases := []struct {
		input    string
		expected []string
	}{
		{
			input:    " hello world ",
			expected: []string{"hello", "world"},
		},
		{
			input:    "yes man the thing is good ",
			expected: []string{"yes", "man", "the", "thing", "is", "good"},
		},
	}

	for _, c := range cases {
		actual := cleanInput(c.input)
		if len(actual) != len(c.expected) {
			t.Errorf("size mismatch was %v but got %v", len(actual), len(c.expected))
			t.Fail()
		}
		for i := range actual {
			word := actual[i]
			expectedWord := c.expected[i]

			if word != expectedWord {
				t.Errorf("Word mismatch, Expected: %v Actual: %v", expectedWord, word)
				t.Fail()
			}
		}
	}
}
