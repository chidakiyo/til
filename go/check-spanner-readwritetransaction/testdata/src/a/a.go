package a

import (
	"cloud.google.com/go/spanner"
	"context"
)

func main() {
	ctx := context.Background()
	client, err := spanner.NewClient(ctx, "projects/P/instances/I/databases/D")
	if err != nil {
		panic(err)
	}
	defer client.Close()

	_, err = client.ReadWriteTransaction(ctx, func(ctx context.Context, txn *spanner.ReadWriteTransaction) error {
		return nil
	}) // want `Foobar`
	if err != nil {
		panic(err)
	}

	container := struct{ cli *spanner.Client }{cli: client}
	_, err = container.cli.ReadWriteTransaction(ctx, func(ctx context.Context, txn *spanner.ReadWriteTransaction) error {
		return nil
	}) // want `Foobar`
	if err != nil {
		panic(err)
	}

	_, err = client.ReadWriteTransaction(ctx, func(ctx context.Context, txn *spanner.ReadWriteTransaction) error {
		txn.BufferWrite([]*spanner.Mutation{ /* set your mutations */ }) // OK
		return nil
	})
	if err != nil {
		panic(err)
	}

	_, err = client.ReadWriteTransaction(ctx, func(ctx context.Context, txn *spanner.ReadWriteTransaction) error {
		txn.Update(ctx, spanner.Statement{ /* write your good DML */ }) // OK
		return nil
	})
	if err != nil {
		panic(err)
	}
}
