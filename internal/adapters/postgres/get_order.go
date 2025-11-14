// package postgres

// import (
// 	"context"
// 	"fmt"
// 	"revivers/internal/domain"
// )

// func (p *Pool) GetOrder(ctx context.Context, user_id string) (domain.Order, error) {
// 	var sub domain.Order

// 	query := `SELECT *
// 			FROM orders
// 			WHERE orders.order_uid = $1`

// 	if err := p.DB.GetContext(ctx, &sub, query, user_id); err != nil {
// 		return domain.Order{}, fmt.Errorf("Error in get orders: %w", err)
// 	}

// 	query = `SELECT *
// 		FROM deliverys
// 		WHERE deliverys.order_uid = $1`

// 	if err := p.DB.GetContext(ctx, &sub.Delivery, query, user_id); err != nil {
// 		return domain.Order{}, fmt.Errorf("Error in get orders: %w", err)
// 	}

// 	query = `SELECT *
// 		FROM payments
// 		WHERE payments.order_uid = $1`

// 	if err := p.DB.GetContext(ctx, &sub.Payment, query, user_id); err != nil {
// 		return domain.Order{}, fmt.Errorf("Error in get orders: %w", err)
// 	}

// 	query = `SELECT *
// 		FROM items
// 		WHERE items.order_uid = $1`

// 	if err := p.DB.SelectContext(ctx, &sub.Items, query, user_id); err != nil {
// 		return domain.Order{}, fmt.Errorf("Error in get items: %w", err)
// 	}

// 	return sub, nil
// }
