package tests

import (
	"testing"
	"github.com/ravendb/ravendb-go-client"
)

type WithID struct {
	N  int
	B  bool
	ID string
}

type WithId struct {
	N  int
	B  bool
	Id string
}

type Withid struct {
	N  int
	B  bool
	id string
}

type NoID struct {
	N int
	B bool
}

type WithIntID struct {
	N  int
	B  bool
	ID int
}

func TestTryGetSetIDFromInstance(t *testing.T) {
	{
		// verify can get/set field name ID of type string
		exp := "hello"
		s := WithID{ID: exp}
		got, ok := ravendb.TryGetIDFromInstance(s)
		if !ok {
			t.Fatalf("TryGetIDFromInstance on %#v failed", s)
		}
		if got != exp {
			t.Fatalf("got %v expected %v", got, exp)
		}
		got, ok = ravendb.TryGetIDFromInstance(&s)
		if !ok {
			t.Fatalf("TryGetIDFromInstance on %#v failed", &s)
		}
		if got != exp {
			t.Fatalf("got %v expected %v", got, exp)
		}

		exp = "new"
		ok = ravendb.TrySetIDOnEntity(s, exp)
		// can't set on structs, only on pointer to structs
		if ok || s.ID == exp {
			t.Fatalf("TrySetIDOnEntity should not succeed on %#v", s)
		}
		ok = ravendb.TrySetIDOnEntity(&s, exp)
		if !ok {
			t.Fatalf("TrySetIDOnEntity failed on %#v", s)
		}
		if s.ID != exp {
			t.Fatalf("TrySetIDOnEntity didn't set ID field to %v on %#v", exp, s)
		}
	}

	{
		// verify can't get/set field name Id of type string
		exp := "hello"
		s := WithId{Id: exp}
		got, ok := ravendb.TryGetIDFromInstance(s)
		if ok || got != "" {
			t.Fatalf("got %v expected %v, ok: %v", got, exp, ok)
		}
		got, ok = ravendb.TryGetIDFromInstance(&s)
		if ok || got != "" {
			t.Fatalf("got %v expected %v, ok: %v", got, exp, ok)
		}
		exp = "new"
		ok = ravendb.TrySetIDOnEntity(s, exp)
		// can't set on structs, only on pointer to structs
		if ok || s.Id == exp {
			t.Fatalf("TrySetIDOnEntity should not succeed on %#v", s)
		}
		ok = ravendb.TrySetIDOnEntity(&s, exp)
		if ok {
			t.Fatalf("TrySetIDOnEntity should fail on %#v", s)
		}
	}

	{
		// verify doesn't get/set unexported field
		exp := "hello"
		s := Withid{id: exp}
		got, ok := ravendb.TryGetIDFromInstance(s)
		if ok || got != "" {
			t.Fatalf("got %v expected %v, ok: %v", got, exp, ok)
		}
		exp = "new"
		ok = ravendb.TrySetIDOnEntity(s, exp)
		if ok {
			t.Fatalf("TrySetIDOnEntity should fail on %#v", s)
		}
	}

	{
		// verify doesn't get/set if there's no ID field
		exp := "new"
		s := NoID{}
		got, ok := ravendb.TryGetIDFromInstance(s)
		if ok || got != "" {
			t.Fatalf("got %v expected %v, ok: %v", got, exp, ok)
		}
		ok = ravendb.TrySetIDOnEntity(s, exp)
		if ok {
			t.Fatalf("TrySetIDOnEntity should fail on %#v", s)
		}
	}

	{
		// verify doesn't get/set if ID is not string
		exp := "new"
		s := WithIntID{ID: 5}
		got, ok := ravendb.TryGetIDFromInstance(s)
		if ok || got != "" {
			t.Fatalf("got %v expected %v, ok: %v", got, exp, ok)
		}
		ok = ravendb.TrySetIDOnEntity(s, exp)
		if ok {
			t.Fatalf("TrySetIDOnEntity should fail on %#v", s)
		}
	}

}
