package main

import (
	"testing"
)

func TestContainsWikiURL(t *testing.T) {

	type addTest struct {
		arg, expected string
	}

	var addTests = []addTest{
		addTest{"https://en.m.wikipedia.org/wiki/Bamse", "https://en.wikipedia.org/wiki/Bamse"},
		addTest{"https://de.m.wikipedia.org/wiki/Bamse", "https://de.wikipedia.org/wiki/Bamse"},
		addTest{"https://en.wikipedia.org/wiki/Bamse", "https://en.wikipedia.org/wiki/Bamse"},
		addTest{"https://de.wikipedia.org/wiki/Bamse", "https://de.wikipedia.org/wiki/Bamse"},
		addTest{"https://de.wikipedia.org/wiki/Bamse?wprov=sfla1", "https://de.wikipedia.org/wiki/Bamse"},
		addTest{"https://en.wikipedia.org/wiki/Bamse?useskin=vector", "https://en.wikipedia.org/wiki/Bamse"},
	}

	for _, test := range addTests {
		if output, _ := containsWikiURL(test.arg); output != test.expected {
			t.Errorf("Output %q not equal to expected %q", output, test.expected)
		}
	}

}
