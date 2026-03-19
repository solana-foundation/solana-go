package solana

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/require"
)

func TestTransactionV0(t *testing.T) {
	txB64 := "Alkhq/BfGdBeok4oBP21xAwT4oO/R5PvkKqbCTq4sHHRsto+uDQCFcdp8hXh1g5D3mTh8GAJW8xE+EDD27f9IweTkH2Afiu4h5aM+Xbo0mklc0/Vi1xawd7SZVbstXDLtWdoJaf4Zt+20F/SasURzw/P4dkD+Q6BjgUNHT+vg5gOgAIBAQgaJV0Ch/DG6XwNcizWbI7STLgSbIOrg0Dl67Oo30WU1uA/NIbYLPRmuLarIJ4J0CcN3IWEm4Gf8675KhnXef2LaDXzjFgWVSbAO2yyTF6dK1oO3gTExie957LXDwu6oJMAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAVKU1qZKSEGTSTocWDaOHx8NbXdvJK7geQfqEBBBUSN1LfoiB9oYLDSHJL9rjAlchZhn+fd/23ACfq0oIGla54pt5JT0MdBTJhQI+z7dnVsisw2xWwW+vFSTs97l0tJPxmv9kxpXbHYZFenDpT2s6CT75/9QNFVTkHFLMK+UG6VlyFnQmYh1aMkGtq3c6TIOsk32S6XMUnN9DQgFGQq4lwEAwIAAgwCAAAAgJaYAAAAAAADAgAFDAIAAACAlpgAAAAAAAMCAAYMAgAAAICWmAAAAAAABAAMSGVsbG8gRmFiaW8hAX5s37FH6IeB4QeMYxD4LtpXf1DaupH/ro7W+kEQnofaAgECAQA="

	tx := new(Transaction)
	err := tx.UnmarshalBase64(txB64)
	require.NoError(t, err)

	require.NotPanics(t, func() {
		spew.Dump(tx.Message)
	})
	require.NotPanics(t, func() {
		tx.MustToBase64()
	})

	require.True(t, tx.Message.IsVersioned())
	require.Equal(t, PublicKeySlice{MPK("9WWfC3y4uCNofr2qEFHSVUXkCxW99JiYkMWmSZvVt8j3")}, tx.Message.GetAddressTableLookups().GetTableIDs())
	require.False(t, tx.Message.resolved)

	// You would fetch the tables from the chain.
	tables := map[PublicKey]PublicKeySlice{
		MPK("9WWfC3y4uCNofr2qEFHSVUXkCxW99JiYkMWmSZvVt8j3"): {
			MPK("2jGpE3ADYRoJPMjyGC4tvqqDfobvdvwGr3vhd66zA1rc"),
			MPK("FKN5imdi7yadX4axe4hxaqBET4n6DBDRF5LKo5aBF53j"),
			MPK("3or4uF7ZyuQW5GGmcmdXDJasNiSZUURF2az1UrRPYQTg"),
			MPK("MemoSq4gqABAXKb96qnH8TysNcWxMyWCqXgDLGmfcHr"),
		},
	}
	// set the address tables
	err = tx.Message.SetAddressTables(
		tables,
	)
	require.NoError(t, err)
	require.Equal(t, tables, tx.Message.GetAddressTables())

	require.Equal(t, MessageVersionV0, tx.Message.GetVersion())
	require.Equal(t, uint8(2), tx.Message.Header.NumRequiredSignatures)
	require.Equal(t, uint8(1), tx.Message.Header.NumReadonlySignedAccounts)
	require.Equal(t, uint8(1), tx.Message.Header.NumReadonlyUnsignedAccounts)

	require.Equal(t, "2nMjR8mdczMJZZ1XeQ5Y37GxfrRQmaV74eypnD9ggpQMmaWfETq9C5DoGKha4bMamu9tFQQArBAgxzQ5vnng1ZdG", tx.Signatures[0].String())
	require.Equal(t, "3x7m4nDNGiZiDgadNtewvHKGcCEWe16QpHo197Azs5ybKNqjzbknuF7VFWeHJ6jowdSeDqVZ2EVgpoq9rNoHvPrM", tx.Signatures[1].String())
	require.Equal(t,
		PublicKeySlice{
			MPK("2m4eNwBVqu6SgFk23HgE3W5MW89yT5z1vspz2WsiFBHF"),
			MPK("G6NDx85GM481GPjT5kUBAvjLxzDMsgRMQ1EAxzGswEJn"),
			MPK("81o7hHYN5a8fc5wdjjfznK9ziJ9wcuKXwbZnuYpanxMQ"),
			MPK("11111111111111111111111111111111"),
			MPK("MemoSq4gqABAXKb96qnH8TysNcWxMyWCqXgDLGmfcHr"),
			MPK("FKN5imdi7yadX4axe4hxaqBET4n6DBDRF5LKo5aBF53j"),
			MPK("3or4uF7ZyuQW5GGmcmdXDJasNiSZUURF2az1UrRPYQTg"),
			MPK("2jGpE3ADYRoJPMjyGC4tvqqDfobvdvwGr3vhd66zA1rc"),
		},
		tx.Message.AccountKeys,
	)
	require.Equal(t,
		MustHashFromBase58("BAx74QRmMwhnTytrPoG5ogw2BQn4CdhB14jxJnbDMUS7"),
		tx.Message.RecentBlockhash,
	)
	require.NotPanics(t, func() {
		spew.Dump(tx.Message)
	})
	require.NotPanics(t, func() {
		tx.MustToBase64()
	})

	{
		err = tx.Message.ResolveLookups()
		require.NoError(t, err)
		require.True(t, tx.Message.resolved)
		// call again
		err = tx.Message.ResolveLookups()
		require.NoError(t, err)
		{
			spew.Dump(tx.Message.AccountKeys)
			require.Equal(t,
				PublicKeySlice{
					MPK("2m4eNwBVqu6SgFk23HgE3W5MW89yT5z1vspz2WsiFBHF"),
					MPK("G6NDx85GM481GPjT5kUBAvjLxzDMsgRMQ1EAxzGswEJn"),
					MPK("81o7hHYN5a8fc5wdjjfznK9ziJ9wcuKXwbZnuYpanxMQ"),
					MPK("11111111111111111111111111111111"),
					MPK("MemoSq4gqABAXKb96qnH8TysNcWxMyWCqXgDLGmfcHr"),
					MPK("FKN5imdi7yadX4axe4hxaqBET4n6DBDRF5LKo5aBF53j"),
					MPK("3or4uF7ZyuQW5GGmcmdXDJasNiSZUURF2az1UrRPYQTg"),
					MPK("2jGpE3ADYRoJPMjyGC4tvqqDfobvdvwGr3vhd66zA1rc"),
					// from tables:
					MPK("FKN5imdi7yadX4axe4hxaqBET4n6DBDRF5LKo5aBF53j"),
					MPK("3or4uF7ZyuQW5GGmcmdXDJasNiSZUURF2az1UrRPYQTg"),
					MPK("2jGpE3ADYRoJPMjyGC4tvqqDfobvdvwGr3vhd66zA1rc"),
				},
				tx.Message.AccountKeys,
			)
		}
	}
	{
		lookups := tx.Message.GetAddressTableLookups()
		require.Equal(t, 1, len(lookups))
		first := lookups[0]
		require.Equal(t,
			MessageAddressTableLookup{
				AccountKey:      MPK("9WWfC3y4uCNofr2qEFHSVUXkCxW99JiYkMWmSZvVt8j3"),
				WritableIndexes: []uint8{1, 2},
				ReadonlyIndexes: []uint8{0},
			}, first)
	}

	{
		encoded := tx.MustToBase64()
		require.Equal(t, txB64, encoded)
	}
}
