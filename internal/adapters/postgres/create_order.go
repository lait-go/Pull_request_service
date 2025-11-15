// package postgres

// import (
// 	"context"
// 	"fmt"
// 	"revivers/internal/domain"
// )

// func (p *Pool) CreateOrder(ctx context.Context, data domain.Order) error {
// 	tx, err := p.DB.BeginTxx(ctx, nil)
// 	if err != nil {
// 		return fmt.Errorf("failed to begin tx: %w", err)
// 	}
// 	defer tx.Rollback()

// 	if _, err := tx.NamedExecContext(ctx, `
//         INSERT INTO orders (
//             order_uid, track_number, entry, locale, internal_signature,
//             customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard
//         ) VALUES (
//             :order_uid, :track_number, :entry, :locale, :internal_signature,
//             :customer_id, :delivery_service, :shardkey, :sm_id, :date_created, :oof_shard
//         )`, data); err != nil {
// 		return fmt.Errorf("error in insert order: %w", err)
// 	}

// 	if _, err := tx.NamedExecContext(ctx, `
//         INSERT INTO deliverys (
//             order_uid, name, phone, zip, city, address, region, email
//         ) VALUES (
//             :order_uid, :name, :phone, :zip, :city, :address, :region, :email
//         )`, data.Delivery); err != nil {
// 		return fmt.Errorf("error in insert delivery: %w", err)
// 	}

// 	if _, err := tx.NamedExecContext(ctx, `
//         INSERT INTO payments (
//             order_uid, transaction, request_id, currency, provider, amount,
//             payment_dt, bank, delivery_cost, goods_total, custom_fee
//         ) VALUES (
//             :order_uid, :transaction, :request_id, :currency, :provider, :amount,
//             :payment_dt, :bank, :delivery_cost, :goods_total, :custom_fee
//         )`, data.Payment); err != nil {
// 		return fmt.Errorf("error in insert payment: %w", err)
// 	}

// 	for _, item := range data.Items {
// 		if _, err := tx.NamedExecContext(ctx, `
// 			INSERT INTO items (
// 				order_uid, chrt_id, track_number, price, rid, name,
// 				sale, size, total_price, nm_id, brand, status
// 			) VALUES (
// 				:order_uid, :chrt_id, :track_number, :price, :rid, :name,
// 				:sale, :size, :total_price, :nm_id, :brand, :status
// 			)`, item); err != nil {
// 			return fmt.Errorf("error in insert item: %w", err)
// 		}
// 	}

// 	return tx.Commit()
// }