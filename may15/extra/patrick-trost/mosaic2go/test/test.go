package test

import (
	"reflect"
	"testing"
)

// Assert fails when a and b are different
func Assert(subject string, a interface{}, b interface{}, t *testing.T) {
	if a != b {
		t.Errorf("%s: Expected %v (type %v), got %v (type %v)", subject, a, reflect.TypeOf(a), b, reflect.TypeOf(b))
	}
}

// AssertNot fails when a and b are equal
func AssertNot(subject string, a interface{}, b interface{}, t *testing.T) {
	if a == b {
		t.Errorf("%s: Did not expect %v (type %v), got %v (type %v)", subject, a, reflect.TypeOf(a), b, reflect.TypeOf(b))
	}
}

// AssertNil fails when a is not null
func AssertNil(subject string, a interface{}, t *testing.T) {
	if !reflect.ValueOf(a).IsNil() {
		t.Errorf("%s: Expected nil, got %v (type %v)", subject, a, reflect.TypeOf(a))
	}
}

// AssertNotNil fails when a is nil
func AssertNotNil(subject string, a interface{}, t *testing.T) {
	if reflect.ValueOf(a).IsNil() {
		t.Errorf("%s: Expected not nil, got %v (type %v)", subject, a, reflect.TypeOf(a))
	}
}

// AssertNotErr checks the provided err object and fails when it has an error
func AssertNotErr(subject string, err error, t *testing.T) {
	if err != nil {
		t.Fatalf("%s: Failed with error: \"%s\".", subject, err.Error())
	}
}
