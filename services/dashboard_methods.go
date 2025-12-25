package services

import (
	"context"

	api "github.com/sunfmin/shadcn-admin-go/api/gen/admin"
)

// GetDashboardStats implements api.Handler.
func (s *AdminService) GetDashboardStats(ctx context.Context) (*api.DashboardStats, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// In production, these would be calculated from actual data
	return &api.DashboardStats{
		TotalRevenue: api.DashboardStatsTotalRevenue{
			Value:  api.NewOptFloat64(45231.89),
			Change: api.NewOptString("+20.1% from last month"),
		},
		Subscriptions: api.DashboardStatsSubscriptions{
			Value:  api.NewOptInt(2350),
			Change: api.NewOptString("+180.1% from last month"),
		},
		Sales: api.DashboardStatsSales{
			Value:  api.NewOptInt(12234),
			Change: api.NewOptString("+19% from last month"),
		},
		ActiveNow: api.DashboardStatsActiveNow{
			Value:  api.NewOptInt(573),
			Change: api.NewOptString("+201 since last hour"),
		},
	}, nil
}

// GetDashboardOverview implements api.Handler.
func (s *AdminService) GetDashboardOverview(ctx context.Context) (*api.DashboardOverview, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// In production, these would be calculated from actual data
	data := []api.DashboardOverviewDataItem{
		{Name: api.NewOptString("Jan"), Total: api.NewOptFloat64(4500)},
		{Name: api.NewOptString("Feb"), Total: api.NewOptFloat64(3200)},
		{Name: api.NewOptString("Mar"), Total: api.NewOptFloat64(5100)},
		{Name: api.NewOptString("Apr"), Total: api.NewOptFloat64(4800)},
		{Name: api.NewOptString("May"), Total: api.NewOptFloat64(6200)},
		{Name: api.NewOptString("Jun"), Total: api.NewOptFloat64(5800)},
		{Name: api.NewOptString("Jul"), Total: api.NewOptFloat64(4900)},
		{Name: api.NewOptString("Aug"), Total: api.NewOptFloat64(5500)},
		{Name: api.NewOptString("Sep"), Total: api.NewOptFloat64(6100)},
		{Name: api.NewOptString("Oct"), Total: api.NewOptFloat64(5300)},
		{Name: api.NewOptString("Nov"), Total: api.NewOptFloat64(4700)},
		{Name: api.NewOptString("Dec"), Total: api.NewOptFloat64(6800)},
	}

	return &api.DashboardOverview{
		Data: data,
	}, nil
}

// GetRecentSales implements api.Handler.
func (s *AdminService) GetRecentSales(ctx context.Context) (*api.RecentSalesResponse, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// In production, these would be fetched from actual data
	data := []api.RecentSale{
		{
			Name:   "Olivia Martin",
			Email:  "olivia.martin@email.com",
			Avatar: api.NewOptString("/avatars/01.png"),
			Amount: 1999.00,
		},
		{
			Name:   "Jackson Lee",
			Email:  "jackson.lee@email.com",
			Avatar: api.NewOptString("/avatars/02.png"),
			Amount: 39.00,
		},
		{
			Name:   "Isabella Nguyen",
			Email:  "isabella.nguyen@email.com",
			Avatar: api.NewOptString("/avatars/03.png"),
			Amount: 299.00,
		},
		{
			Name:   "William Kim",
			Email:  "will@email.com",
			Avatar: api.NewOptString("/avatars/04.png"),
			Amount: 99.00,
		},
		{
			Name:   "Sofia Davis",
			Email:  "sofia.davis@email.com",
			Avatar: api.NewOptString("/avatars/05.png"),
			Amount: 39.00,
		},
	}

	return &api.RecentSalesResponse{
		Data:       data,
		TotalSales: 2475,
	}, nil
}
