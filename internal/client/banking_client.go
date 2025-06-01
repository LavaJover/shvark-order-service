package client

import (
	"context"
	"time"

	bankingpb "github.com/LavaJover/shvark-banking-service/proto/gen"
	"github.com/LavaJover/shvark-order-service/internal/domain"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type BankingClient struct {
	conn *grpc.ClientConn
	service bankingpb.BankingServiceClient
}

func NewbankingClient(addr string) (*BankingClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(
		ctx,
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`),
	)

	if err != nil {
		return nil, err
	}

	return &BankingClient{
		conn: conn,
		service: bankingpb.NewBankingServiceClient(conn),
	}, nil
}

func (c *BankingClient) GetEligibleBankDetails(query *domain.BankDetailQuery) (*bankingpb.GetEligibleBankDetailsResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return c.service.GetEligibleBankDetails(
		ctx,
		&bankingpb.GetEligibleBankDetailsRequest{
			Currency: query.Currency,
			Country: query.Country,
			Amount: float64(query.Amount),
			PaymentSystem: query.PaymentSystem,
		},
	)
}

func (c *BankingClient) GetBankDetailByID(bankDetailID string) (*bankingpb.GetBankDetailByIDResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return c.service.GetBankDetailByID(
		ctx,
		&bankingpb.GetBankDetailByIDRequest{
			BankDetailId: bankDetailID,
		},
	)
}