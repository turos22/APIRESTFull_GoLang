package orders

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	repo "github.com/turos22/APIRESTFull_GoLang/internal/adapters/postgresql/sqlc"
)

var (
	ErrProductNotFound = fmt.Errorf("product not found")
	ErrProductOutOfStock = fmt.Errorf("product out of stock")
)

type svc struct {
	repo repo.Queries
	db *pgxpool.Pool
	rdb *redis.Client
}

func NewService(repo *repo.Queries, db *pgxpool.Pool, rdb *redis.Client) Service {
	return &svc{
		repo: *repo,
		db: db,
		rdb: rdb,
	}
}


func (svc *svc) PlaceOrder(ctx context.Context, tempOrder createOrderParams) (repo.Order, []repo.OrderItem, error) {
	if tempOrder.CustomerID == 0{
		return repo.Order{},[]repo.OrderItem{}, fmt.Errorf("CostumerID is required")
	}
	if len(tempOrder.Items) == 0{
		return repo.Order{},[]repo.OrderItem{}, fmt.Errorf("Items is required")
	}

	tx, err := svc.db.Begin(ctx)
	if err != nil {
		return repo.Order{},[]repo.OrderItem{}, err
	}
	defer tx.Rollback(ctx)

	qtx := svc.repo.WithTx(tx)

	order, err := qtx.CreateOrder(ctx, tempOrder.CustomerID)
	if err != nil {
		return repo.Order{},[]repo.OrderItem{}, err
	}

	var itensOrder []repo.OrderItem
	//ver se produto existe
	for _, item := range tempOrder.Items {
		product, err := qtx.FindProductByID(ctx, item.ProductID)
		if err != nil {
			return repo.Order{},[]repo.OrderItem{}, ErrProductNotFound
		}

		if product.Quantity < item.Quantity {
			return repo.Order{},[]repo.OrderItem{}, ErrProductOutOfStock
		}

		itemzinho, err := qtx.CreateOrderItem(ctx, repo.CreateOrderItemParams{
			OrderID: order.ID,
			ProductID: item.ProductID,
			Quantity: item.Quantity,
			PriceCents: product.PriceInCents,
		})
	
		if err != nil {
			return repo.Order{},[]repo.OrderItem{}, err
		}

		itensOrder = append(itensOrder, itemzinho)

		_, err = qtx.UpdateStock(ctx, repo.UpdateStockParams{
			ID: product.ID,
			Quantity: product.Quantity - item.Quantity,
		})

		if err != nil {
			return repo.Order{},[]repo.OrderItem{}, err
		}

		
	}
	tx.Commit(ctx)

	//criar evento de Stream no Redis
	if err := svc.rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: "orders.created",
		Values: map[string]interface{}{
			"order_id":    order.ID,
			"customer_id": order.CustomerID,
		},
	}).Err(); err != nil {
		fmt.Println("falha ao publicar orders.created:", err)
	}

	return order, itensOrder, nil

}

func (svc *svc) GetOrderId(ctx context.Context, id int64) (repo.Order, error){
	return svc.repo.OrdersId(ctx, id)
}

func (svc *svc) MeOrder(ctx context.Context, id int64) ([]repo.Order, error){
	return svc.repo.OrdersMe(ctx, id)
}

func (svc *svc) GetItemsOrder(ctx context.Context, id int64) ([]repo.OrderItem, error){
	return svc.repo.OrderItemsOrderId(ctx, id)
}

func (svc *svc) OrderMeId(ctx context.Context, id int64, customerId int64) (repo.Order, error){
	return svc.repo.OrderMeId(ctx, repo.OrderMeIdParams{
		ID: id,
		CustomerID: customerId,
	})
}