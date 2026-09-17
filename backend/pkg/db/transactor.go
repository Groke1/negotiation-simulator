package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type DB interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

type transactorImpl struct {
	db DB
}

func NewTransactor(db DB) *transactorImpl {
	return &transactorImpl{db: db}
}

func (t *transactorImpl) WithTx(ctx context.Context, f func(ctx context.Context) error) (err error) {
	if _, err = ExtractTx(ctx); err == nil {
		return f(ctx)
	}

	tx, err := t.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	ctx = injectTx(ctx, tx)
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
			return
		}
		if commitErr := tx.Commit(ctx); commitErr != nil {
			err = fmt.Errorf("commit tx: %w", commitErr)
		}
	}()

	if err = f(ctx); err != nil {
		return fmt.Errorf("execute in tx: %w", err)
	}
	return nil
}

type txKey struct{}

var ErrTxNotFound = errors.New("tx not found in context")

func ExtractTx(ctx context.Context) (pgx.Tx, error) {
	tx, ok := ctx.Value(txKey{}).(pgx.Tx)
	if !ok {
		return nil, ErrTxNotFound
	}
	return tx, nil
}

func injectTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}
