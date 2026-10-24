package main

import "testing"

func TestFulfillRejectsUnpaidOrder(t *testing.T) {
	err := Fulfill(Order{ID: "o-7", TotalCents: 1200, State: "pending"}, nil)
	if err == nil {
		t.Fatal("expected pending order to be rejected")
	}
}
