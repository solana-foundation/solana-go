package solana

import orderedmap "github.com/wk8/go-ordered-map/v2"

func newAddressTableMap(capacity int) *orderedmap.OrderedMap[PublicKey, PublicKeySlice] {
	return orderedmap.New[PublicKey, PublicKeySlice](orderedmap.WithCapacity[PublicKey, PublicKeySlice](capacity))
}

func addressTableMapFromMap(tables map[PublicKey]PublicKeySlice) *orderedmap.OrderedMap[PublicKey, PublicKeySlice] {
	om := newAddressTableMap(len(tables))
	for k, v := range tables {
		om.Set(k, v)
	}
	return om
}

func addressTableMapFromSlice(tables []AddressTableEntry) *orderedmap.OrderedMap[PublicKey, PublicKeySlice] {
	om := newAddressTableMap(len(tables))
	for _, entry := range tables {
		om.Set(entry.TableKey, entry.Addresses)
	}
	return om
}

func addressTableMapToMap(om *orderedmap.OrderedMap[PublicKey, PublicKeySlice]) map[PublicKey]PublicKeySlice {
	out := make(map[PublicKey]PublicKeySlice, om.Len())
	for pair := om.Oldest(); pair != nil; pair = pair.Next() {
		out[pair.Key] = pair.Value
	}
	return out
}

func addressTableMapToSlice(om *orderedmap.OrderedMap[PublicKey, PublicKeySlice]) []AddressTableEntry {
	out := make([]AddressTableEntry, 0, om.Len())
	for pair := om.Oldest(); pair != nil; pair = pair.Next() {
		out = append(out, AddressTableEntry{TableKey: pair.Key, Addresses: pair.Value})
	}
	return out
}
