package main

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"revivers/internal/domain"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

func main() {
	for range 10 {
		data := CreateRandomOrder()
		jsonData, err := json.Marshal(data)
		if err != nil {
			log.Fatal("failed to marshal data:", err)
		}

		writer := kafka.NewWriter(kafka.WriterConfig{
			Brokers:  []string{"localhost:9092"},
			Topic:    "my-order",
			Balancer: &kafka.LeastBytes{},
		})
		defer writer.Close()

		err = writer.WriteMessages(context.Background(),
			kafka.Message{
				Key:   []byte("order" + strconv.Itoa(rand.Intn(1000))),
				Value: jsonData,
			},
		)
		if err != nil {
			log.Fatal("failed to write message:", err)
		}
	}

	log.Println("Message sent successfully")
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func randomPhone() string {
	return "+1" + strconv.Itoa(1000000000+rand.Intn(900000000))
}

func randomEmail() string {
	return randomString(6) + "@example.com"
}

func randomAmount(min, max int) int {
	return rand.Intn(max-min+1) + min
}

func CreateRandomOrder() domain.Order {
	rand.Seed(time.Now().UnixNano())
	ordUID := uuid.New()

	itemsCount := rand.Intn(3) + 1 // 1-3 случайных предмета
	items := make([]domain.Item, itemsCount)
	for i := 0; i < itemsCount; i++ {
		items[i] = domain.Item{
			OrderUID:    ordUID.String(),
			ChrtID:      rand.Int63n(1_000_000),
			TrackNumber: "TN" + strconv.Itoa(rand.Intn(999999999)),
			Price:       randomAmount(10, 500),
			Rid:         randomString(8),
			Name:        "Item " + randomString(5),
			Sale:        rand.Intn(50),
			Size:        []string{"S", "M", "L", "XL"}[rand.Intn(4)],
			TotalPrice:  randomAmount(10, 500),
			NmID:        rand.Int63n(1_000_000_000),
			Brand:       "Brand " + randomString(5),
			Status:      rand.Intn(5),
		}
	}

	return domain.Order{
		OrderUID:          ordUID.String(),
		TrackNumber:       "TN" + strconv.Itoa(rand.Intn(999999999)),
		Entry:             "entry" + randomString(3),
		Locale:            "en-US",
		InternalSignature: randomString(10),
		CustomerID:        "cust" + strconv.Itoa(rand.Intn(1000)),
		DeliveryService:   "service" + randomString(3),
		ShardKey:          "shard" + strconv.Itoa(rand.Intn(10)),
		SmID:              rand.Intn(10),
		DateCreated:       time.Now().Add(-time.Duration(rand.Intn(1000)) * time.Hour),
		OofShard:          "oof" + randomString(3),
		Delivery: domain.Delivery{
			OrderUID: ordUID.String(),
			Name:     "User " + randomString(5),
			Phone:    randomPhone(),
			Zip:      strconv.Itoa(10000 + rand.Intn(89999)),
			City:     "City" + randomString(3),
			Address:  strconv.Itoa(rand.Intn(999)) + " " + randomString(5) + " St",
			Region:   "Region" + randomString(3),
			Email:    randomEmail(),
		},
		Payment: domain.Payment{
			OrderUID:     ordUID.String(),
			Transaction:  "txn" + strconv.Itoa(rand.Intn(1_000_000)),
			RequestID:    "req" + strconv.Itoa(rand.Intn(1_000_000)),
			Currency:     "USD",
			Provider:     "provider" + randomString(3),
			Amount:       randomAmount(100, 1000),
			PaymentDt:    time.Now().Unix(),
			Bank:         "bank" + randomString(3),
			DeliveryCost: randomAmount(10, 50),
			GoodsTotal:   randomAmount(50, 900),
			CustomFee:    randomAmount(0, 100),
		},
		Items: items,
	}
}
