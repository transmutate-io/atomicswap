package db

import "context"

func OpenLevelDB(path string) (Database, error) { return &swapperLDB{}, nil }

type swapperLDB struct{}

func (s *swapperLDB) Close(ctx context.Context) error { return nil }
