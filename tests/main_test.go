package tests

import (
	"fmt"
	"testing"

	"github.com/amirgamil/apollo/pkg/apollo/backend"
)

//compare arrays because deepEqual is inconsistent smh
func compareArrays(arr1 []string, arr2 []string) bool {
	equal := len(arr1) == len(arr2)
	for i := 0; i < len(arr1); i++ {
		if arr1[i] != arr2[i] {
			return false
		}
	}
	return true && equal
}

func TestTokenizer(t *testing.T) {
	sanityTest := "Hi, my name is John!"
	expected := []string{"hi", "name", "john"}
	result := backend.Tokenize(sanityTest)
	equal := compareArrays(expected, result)
	if !equal {
		t.Errorf(fmt.Sprintf("Uh oh, test 1 expected %s but got %s", expected, result))
	}

	//test punctuation, capital letters
	moreComplex := "help me! (I'm feeling luCKy) today"
	expected = []string{"help", "'m", "feeling", "lucky", "today"}
	result = backend.Tokenize(moreComplex)
	equal = compareArrays(expected, result)
	if !equal {
		t.Errorf(fmt.Sprintf("Uh oh, test 2 expected %s but got %s", expected, result))
	}

	//test quotes
	moreComplex = "but I \"don't know What it's going to be like\""
	expected = []string{"know", "'s", "going", "like"}
	result = backend.Tokenize(moreComplex)
	equal = compareArrays(expected, result)
	if !equal {
		t.Errorf(fmt.Sprintf("Uh oh, test 3 expected %s but got %s", expected, result))
	}

	moreComplex = "'Hello', you're [cool*!%] and funny too"
	//way we handle single quotes, should give us an empty
	expected = []string{"hello", "cool", "funny"}
	result = backend.Tokenize(moreComplex)
	equal = compareArrays(expected, result)
	if !equal {
		t.Errorf(fmt.Sprintf("Uh oh, test 3 expected %s but got %s", expected, result))
	}

	result = backend.Tokenize("")
	expected = make([]string, 0)
	equal = compareArrays(result, expected)
	if !equal {
		t.Errorf(fmt.Sprintf("Uh oh, test 4 expected %s but got %s", expected, result))
	}

	//same word different formations
	result = backend.Tokenize("Hello something SOMETHING someTHiNg")
	expected = []string{"hello", "something", "something", "something"}
	equal = compareArrays(result, expected)
	if !equal {
		t.Errorf(fmt.Sprintf("Uh oh, test 4 expected %s but got %s", expected, result))
	}

	result = backend.Tokenize("Test\nhello sir")
	expected = []string{"test", "hello", "sir"}
	equal = compareArrays(result, expected)
	if !equal {
		t.Errorf(fmt.Sprintf("Uh oh, test 4 expected %s but got %s", expected, result))
	}
}
