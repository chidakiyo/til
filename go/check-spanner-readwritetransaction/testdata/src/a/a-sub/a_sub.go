package a_sub

import "cloud.google.com/go/spanner"

func NotUseTx(txn *spanner.ReadWriteTransaction) error {
	return nil
}

func UseTx(txn *spanner.ReadWriteTransaction) error { // OK
	return txn.BufferWrite([]*spanner.Mutation{ /* set your mutations */ })
}
