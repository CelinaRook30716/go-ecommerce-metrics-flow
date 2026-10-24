package main

import "fmt"

type Order struct {
	ID         string
	TotalCents int
	State      string
}

func Fulfill(order Order, infrai *InfraiClient) error {
	if order.State != "paid" {
		return fmt.Errorf("order %s is not ready for fulfillment", order.ID)
	}
	if err := infrai.ReportMetric("orders.fulfilled", "counter", 1, map[string]string{"state": "shipped"}); err != nil {
		return err
	}
	if err := infrai.ReportMetric("orders.revenue_cents", "gauge", float64(order.TotalCents), map[string]string{"currency": "USD"}); err != nil {
		return err
	}
	if err := infrai.IngestLog("receipt queued for "+order.ID, "info", map[string]string{"order_id": order.ID}); err != nil {
		return err
	}
	return nil
}
