package gomp4dur

import (
	"math"
	"os"
	"testing"
)

func Test_Get_MOOV(t *testing.T) {
	moov_file, err := os.OpenFile("resources/big-buck-bunny.mp4", os.O_RDONLY, 0644)

	if err != nil {
		t.Fatalf("%v", err)
	}

	defer moov_file.Close()

	duration, err := Get(moov_file)

	if err != nil {
		t.Fatalf("%v", err)
	}

	if math.Abs(duration-float64(10.0)) >= 0.0000001 {
		t.Fatalf("The duration didn't match, expected: %f, found: %f", 10.0, duration)
	}
}

func Test_Get_MOOF(t *testing.T) {
	moof_file, err := os.OpenFile("resources/big-buck-bunny-frag.mp4", os.O_RDONLY, 0644)

	if err != nil {
		t.Fatalf("%v\n", err)
	}

	defer moof_file.Close()

	duration, err := Get(moof_file)

	if err != nil {
		t.Fatal(err)
	}

	if math.Abs(duration-float64(10.0)) >= 0.0000001 {
		t.Fatalf("The duration didn't match, expected: %f, found: %f", 10.0, duration)
	}
}

func Test_Stringify(t *testing.T) {
	file, err := os.OpenFile("resources/big-buck-bunny.mp4", os.O_RDONLY, 0644)

	if err != nil {
		t.Error(file)
	}

	defer file.Close()

	duration, err := Get(file)

	if err != nil {
		t.Error(err)
	}

	str_dur := Stringify(duration)

	if str_dur != "00:10" {
		t.Errorf("Stringification of obtained duration failed, expected - \"%s\", found: \"%s\"", "00:10", str_dur)
	}
}
