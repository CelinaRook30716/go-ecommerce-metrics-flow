package main

import (
	"fmt"
	"log"
)

func main() {
	client, err := NewInfraiClient()
	if err != nil {
		log.Fatal(err)
	}
	order := Order{ID: "order-1042", TotalCents: 8499, State: "paid"}
	if err := Fulfill(order, client); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("order %s shipped; receipt event recorded\n", order.ID)
}
