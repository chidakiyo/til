package a

import (
	"context"
	"fmt"

	"a/a-sub"
	"cloud.google.com/go/spanner"
)

type testStruct struct {
}

func (testStruct) NotUseTx(ctx context.Context, txn *spanner.ReadWriteTransaction) error { // want `this function never calls \*spanner.ReadWriteTransaction's BufferWrite or Update method`
	return nil
}

func (testStruct) UseTx(ctx context.Context, txn *spanner.ReadWriteTransaction) error {
	{
		txn := txn
		txn.BufferWrite([]*spanner.Mutation{ /* set your mutations */ }) // OK
	}
	return nil
}

func main() {
	ctx := context.Background()
	client, err := spanner.NewClient(ctx, "projects/P/instances/I/databases/D")
	if err != nil {
		panic(err)
	}
	defer client.Close()

	// NG!
	{
		_, err = client.ReadWriteTransaction(ctx, func(ctx context.Context, txn *spanner.ReadWriteTransaction) error { // want `this function never calls \*spanner.ReadWriteTransaction's BufferWrite or Update method`
			return nil
		})
		if err != nil {
			panic(err)
		}
	}
	{
		container := struct{ cli *spanner.Client }{cli: client}
		_, err = container.cli.ReadWriteTransaction(ctx, func(ctx context.Context, txn *spanner.ReadWriteTransaction) error { // want `this function never calls \*spanner.ReadWriteTransaction's BufferWrite or Update method`
			return nil
		})
		if err != nil {
			panic(err)
		}
	}
	{
		rwtx := client.ReadWriteTransaction
		_, err = rwtx(ctx, func(ctx context.Context, txn *spanner.ReadWriteTransaction) error { // want `this function never calls \*spanner.ReadWriteTransaction's BufferWrite or Update method`
			return nil
		})
		if err != nil {
			panic(err)
		}
	}
	{
		callback := func(ctx context.Context, txn *spanner.ReadWriteTransaction) error { // want `this function never calls \*spanner.ReadWriteTransaction's BufferWrite or Update method`
			return nil
		}
		_, err = client.ReadWriteTransaction(ctx, callback)
		if err != nil {
			panic(err)
		}
	}
	{
		callbackGenerator := func() func(ctx context.Context, txn *spanner.ReadWriteTransaction) error {
			return func(ctx context.Context, txn *spanner.ReadWriteTransaction) error { // want `this function never calls \*spanner.ReadWriteTransaction's BufferWrite or Update method`
				return nil
			}
		}
		_, err = client.ReadWriteTransaction(ctx, callbackGenerator())
		if err != nil {
			panic(err)
		}
	}
	{
		container := struct {
			callback func(ctx context.Context, txn *spanner.ReadWriteTransaction) error
		}{
			callback: func(ctx context.Context, txn *spanner.ReadWriteTransaction) error { // want `this function never calls \*spanner.ReadWriteTransaction's BufferWrite or Update method`
				return nil
			},
		}
		_, err = client.ReadWriteTransaction(ctx, container.callback)
		if err != nil {
			panic(err)
		}
	}
	{
		container := &testStruct{}
		_, err = client.ReadWriteTransaction(ctx, container.NotUseTx)
		if err != nil {
			panic(err)
		}
	}
	{
		_, err = client.ReadWriteTransaction(ctx, func(ctx context.Context, txn *spanner.ReadWriteTransaction) error { // want `this function never calls \*spanner.ReadWriteTransaction's BufferWrite or Update method`
			f := func() {
				txn.BufferWrite([]*spanner.Mutation{ /* set your mutations */ }) // OK
			}
			// don't call f()
			fmt.Println(f)
			return nil
		})
		if err != nil {
			panic(err)
		}
	}
	{
		_, err = client.ReadWriteTransaction(ctx, func(ctx context.Context, txn *spanner.ReadWriteTransaction) error { // want `this function never calls \*spanner.ReadWriteTransaction's BufferWrite or Update method`
			return a_sub.NotUseTx(txn)
		})
		if err != nil {
			panic(err)
		}
	}

	// OK!
	{
		_, err = client.ReadWriteTransaction(ctx, func(ctx context.Context, txn *spanner.ReadWriteTransaction) error {
			txn.BufferWrite([]*spanner.Mutation{ /* set your mutations */ }) // OK
			return nil
		})
		if err != nil {
			panic(err)
		}
	}
	{
		_, err = client.ReadWriteTransaction(ctx, func(ctx context.Context, txn *spanner.ReadWriteTransaction) error {
			txn.Update(ctx, spanner.Statement{ /* write your good DML */ }) // OK
			return nil
		})
		if err != nil {
			panic(err)
		}
	}
	{ // TODO まだ検出できない
		_, err = client.ReadWriteTransaction(ctx, func(ctx context.Context, txn *spanner.ReadWriteTransaction) error {
			bw := txn.BufferWrite
			bw([]*spanner.Mutation{ /* set your mutations */ }) // OK
			return nil
		})
		if err != nil {
			panic(err)
		}
	}
	{ // TODO まだ検出できない
		_, err = client.ReadWriteTransaction(ctx, func(ctx context.Context, txn *spanner.ReadWriteTransaction) error {
			container := struct{ txn *spanner.ReadWriteTransaction }{txn}
			container.txn.BufferWrite([]*spanner.Mutation{ /* set your mutations */ }) // OK
			return nil
		})
		if err != nil {
			panic(err)
		}
	}
	{ // TODO まだ検出できない
		useTx := func(tx *spanner.ReadWriteTransaction) error {
			tx.BufferWrite([]*spanner.Mutation{ /* set your mutations */ }) // OK
			return nil
		}
		_, err = client.ReadWriteTransaction(ctx, func(ctx context.Context, txn *spanner.ReadWriteTransaction) error {
			return useTx(txn)
		})
		if err != nil {
			panic(err)
		}
	}
	{
		container := &testStruct{}
		_, err = client.ReadWriteTransaction(ctx, container.UseTx)
		if err != nil {
			panic(err)
		}
	}
	{ // TODO まだ検出できない
		_, err = client.ReadWriteTransaction(ctx, func(ctx context.Context, txn *spanner.ReadWriteTransaction) error {
			f := func() {
				txn.BufferWrite([]*spanner.Mutation{ /* set your mutations */ }) // OK
			}
			f()
			return nil
		})
		if err != nil {
			panic(err)
		}
	}
	{ // TODO まだ検出できない
		_, err = client.ReadWriteTransaction(ctx, func(ctx context.Context, txn *spanner.ReadWriteTransaction) error {
			return a_sub.UseTx(txn)
		})
		if err != nil {
			panic(err)
		}
	}
}
