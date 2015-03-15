package main

import "testing"

func TestParseDuration(t *testing.T) {
	v, err := parseDuration("55s")
	if err != nil {
		t.Errorf("something wrong")
	}
	if v != 55 {
		t.Errorf("parse failed!")
	}

	v, err = parseDuration("55")
	if err != nil {
		t.Errorf("something wrong")
	}
	if v != 55 {
		t.Errorf("parse failed!")
	}

	v, err = parseDuration("10m")
	if err != nil {
		t.Errorf("something wrong")
	}
	if v != 600 {
		t.Errorf("parse failed!")
	}

	v, err = parseDuration("1h")
	if err != nil {
		t.Errorf("something wrong")
	}
	if v != 3600 {
		t.Errorf("parse failed!")
	}

	v, err = parseDuration("1d")
	if err != nil {
		t.Errorf("something wrong")
	}
	if v != 86400 {
		t.Errorf("parse failed!")
	}

	_, err = parseDuration("1w")
	if err == nil {
		t.Errorf("something wrong")
	}
}
