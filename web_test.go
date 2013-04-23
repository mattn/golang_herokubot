package main

import (
	"regexp"
	"testing"
)

type botTest struct {
	in string
	out string
}

var botTests = []botTest {
	{
		"!heroke",
		"^$",
	},
	{
		"!heroku",
		"^へろくー$",
	},
	{
		" !heroku",
		"^$",
	},
	{
		"!heroku ",
		"^$",
	},
	{
		"!weather",
		"^$",
	},
	{
		"!weather osaka",
		"^http://openweathermap\\.org/img/w/\\d+\\w\\.png\n",
	},
	{
		"!uranai 牡羊座",
		"^\\d+位 牡羊座\n",
	},
}

func TestReadAll(t *testing.T) {
	for _, test := range botTests {
		expected := regexp.MustCompile(test.out)
		events := []Event {
			{
				111,
				&Message{
					"222",
					"foo",
					"XXX",
					"http://mattn.kaoriya.net/images/logo.png",
					"human",
					"333",
					"mattn",
					test.in,
				},
			},
		}
		got := handleEvents(events)
		if !expected.MatchString(got) {
			t.Errorf("Unexpected response for request %q; got %q; expected %q",
				test.in, got, expected.String())
		}
	}
}
