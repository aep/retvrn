package kv

import (
	"context"
	"github.com/tikv/client-go/v2/txnkv"
	"github.com/tikv/client-go/v2/txnkv/txnsnapshot"
	"iter"
)

type Tikv struct {
	k *txnkv.Client
}

type TikvWrite struct {
	txn *txnkv.KVTxn
	err error
}

func (w *TikvWrite) Commit(ctx context.Context) error {
	if w.err != nil {
		return w.err
	}
	return w.txn.Commit(ctx)
}

func (w *TikvWrite) Rollback() error {
	if w.err != nil {
		return w.err
	}
	return w.txn.Rollback()
}

func (w *TikvWrite) Set(key []byte, value []byte) error {
	if w.err != nil {
		return w.err
	}
	err := w.txn.Set(key, value)
	if err != nil {
		w.Rollback()
		w.err = err
	}
	return w.err
}

func (w *TikvWrite) Get(ctx context.Context, key []byte) ([]byte, error) {
	if w.err != nil {
		return nil, w.err
	}
	return w.txn.Get(ctx, key)
}

type TikvRead struct {
	txn *txnsnapshot.KVSnapshot
	err error
}

func (r *TikvRead) Get(ctx context.Context, key []byte) ([]byte, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.txn.Get(ctx, key)
}

func (r *TikvRead) Close() {
}

func (r *TikvRead) Iter(ctx context.Context, start []byte, end []byte) iter.Seq2[KeyAndValue, error] {

	return func(yield func(KeyAndValue, error) bool) {

		it, err := r.txn.Iter(start, end)
		if err != nil {
			yield(KeyAndValue{}, err)
			return
		}

		for it.Valid() {

			if !yield(KeyAndValue{K: it.Key(), V: it.Value()}, nil) {
				return
			}

			err := it.Next()
			if err != nil {
				if !yield(KeyAndValue{}, err) {
					return
				}
			}
		}
	}
}

func (t *Tikv) Close() {
	t.k.Close()
}

func (t *Tikv) Write() Write {
	txn, err := t.k.Begin()
	return &TikvWrite{txn, err}
}

func (t *Tikv) Read() Read {
	ts, err := t.k.CurrentTimestamp("global")
	if err != nil {
		return &TikvRead{nil, err}
	}

	txn := t.k.GetSnapshot(ts)
	return &TikvRead{txn, nil}
}

func NewTikv() (KV, error) {
	k, err := txnkv.NewClient([]string{"127.0.0.1:2379"})
	if err != nil {
		return nil, err
	}

	return &Tikv{k}, nil
}
