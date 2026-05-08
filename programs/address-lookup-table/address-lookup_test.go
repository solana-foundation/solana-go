package addresslookuptable

import (
	"bytes"
	"encoding/base64"
	"math"
	"testing"

	bin "github.com/gagliardetto/binary"
	"github.com/solana-foundation/solana-go/v2"
	"github.com/stretchr/testify/require"
)

func TestDecodeTable(t *testing.T) {
	tableAccountBase64 := "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"

	tableAccountBytes, err := base64.StdEncoding.DecodeString(tableAccountBase64)
	require.NoError(t, err)

	table, err := DecodeAddressLookupTableState(tableAccountBytes)
	require.NoError(t, err)

	require.Equal(t, uint64(math.MaxUint64), table.DeactivationSlot)
	require.Equal(t, uint64(154742572), table.LastExtendedSlot)
	require.Equal(t, uint8(232), table.LastExtendedSlotStartIndex)
	require.Equal(t, solana.MPK("9FRhPDoDk9JrpCqc4r51qTWgdBTxM892TdjexeErQUNs"), *table.Authority)
	require.Equal(t, uint8(232), table.LastExtendedSlotStartIndex)

	expectedKeys := solana.PublicKeySlice{
		solana.MPK("9W959DqEETiGZocYWCQPaJ6sBmUzgfxXfqGeTEdp3aQP"),
		solana.MPK("TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"),
		solana.MPK("GW1Xt9HHtvcnky8X7aBA3BoTgiirJKP5XwC5REFcZSsc"),
		solana.MPK("GXueH9K1MzRncoTYbpLiXXC3WrKkmHUFxV5JEu8oADbw"),
		solana.MPK("F7XioZaGe99nosYJQCahx25TKgdUGufYf6sudm1JSgu"),
		solana.MPK("BT14DfFyNS7qcBGc8TY4HAzDev4vvqsoFBJgjtQpdM2Z"),
		solana.MPK("C2YzN6MymD5HM2kPaH7bzcbqciyjfmpqyVaR3KA5V6z1"),
		solana.MPK("BhHd49JYH3Hk6TV5kCjmUgf7fQSQKDjaWTokMmBhTx9o"),
		solana.MPK("9W959DqEETiGZocYWCQPaJ6sBmUzgfxXfqGeTEdp3aQP"),
		solana.MPK("TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"),
		solana.MPK("8JPid6GtND2tU3A7x7GDfPPEWwS36rMtzF7YoHU44UoA"),
		solana.MPK("749y4fXb9SzqmrLEetQdui5iDucnNiMgCJ2uzc3y7cou"),
		solana.MPK("ErcxwkPgLdyoVL6j2SsekZ5iysPZEDRGfAggh282kQb8"),
		solana.MPK("EFYW6YEiCGpavuMPS1zoXhgfNkPisWkQ3bQz1b4UfKek"),
		solana.MPK("BVWwyiHVHZQMPHsiW7dZH7bnBVKmbxdeEjWqVRciHCyo"),
		solana.MPK("E6aTzkZKdCECgpDtBZtVpqiGjxRDSAFh1SC9CdSoVK7a"),
		solana.MPK("675kPX9MHTjS2zt1qfr1NYHuzeLXfQM9H24wFSUt1Mp8"),
		solana.MPK("TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"),
		solana.MPK("ABrn4ED4AvkQ79VAXqf7ooqicJPHhZDAbC9rqcQ8ePzz"),
		solana.MPK("5Q544fKrFoe6tsEbD7S8EmxGTJYAKtTVhAW5Q5pge4j1"),
		solana.MPK("D7CHbxSFSiEW3sPc486AGDwuwsmyZqhP7stG4Yo9ZHTC"),
		solana.MPK("5o8dopjEKEy491bVHShtG6KSSHKm2JUugVqKEK7Jw7YF"),
		solana.MPK("FN3wMZUuWkM65ZtcnAoYpsq773YxrnMfM5iAroSGttBo"),
		solana.MPK("9xQeWvG816bUx9EPjHmaT23yvVM2ZWbrrpZb9PusVFin"),
		solana.MPK("6uGUx583UHvFKKCnoMfnGNEFxhSWy5iXXyea4o5E9dx7"),
		solana.MPK("Gp7wpKu9mXpxdykMD9JKW5SK2Jw1h2fttxukvcL2dnW6"),
		solana.MPK("4mkSxT9MaUsUd5uSkZxohf1pbPByk7b5ptWpu4ZABvto"),
		solana.MPK("4dDEjb4JZejtweFEJjjqqC5wwZi3jqtzoS7cPNRyPoT6"),
		solana.MPK("Geoh8p8j48Efupens8TqJKj491aqk5VhPXABFAqGtAjr"),
		solana.MPK("EVv4jPvUxbugw8EHTDwkNBboE26DiN4Zy1CQrd5j3Sd4"),
		solana.MPK("3ceGkbGkqQwjJsZEYzjykDcWM1FjzHGMNTyKHD1c7kqW"),
		solana.MPK("9W959DqEETiGZocYWCQPaJ6sBmUzgfxXfqGeTEdp3aQP"),
		solana.MPK("TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"),
		solana.MPK("3Ukqqshh3kZ8UPbcYYFSRaeJcsgttcmShtNNn12F1rj2"),
		solana.MPK("7NP8DTzPdpbQofhNyhLW3j2khutmfy1kuFp2AjaD8rrp"),
		solana.MPK("F6nCAMYEFxsyRDVonQXLNufXgAHsgAa1Br8DhBoX3KAV"),
		solana.MPK("HWCTHmQppFSsKQEk1bHUqPC2WLaidgnfTG9MQGD4XKEt"),
		solana.MPK("GHuoeq9UnFBsBhMwH43eL3RWX5XVXbSRYJymmyMYpT7n"),
		solana.MPK("CCuSVbnnq8SUj7cpPe7BbHLuKanyxfvfrwypzCBnaDdf"),
		solana.MPK("9W959DqEETiGZocYWCQPaJ6sBmUzgfxXfqGeTEdp3aQP"),
		solana.MPK("TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"),
		solana.MPK("2Y1Jmpkf5wt1X5zcFHNBDoHxqjTXbMJfj1UFtrYQwUbG"),
		solana.MPK("8K4eRHeyPhBGB9zCjKtyBHoPPZ75zLN64oxBF6GyF4Qg"),
		solana.MPK("DRYADMQevoJHDFCYbDQeS4p551MpsDN2d7CJU3LxfNHa"),
		solana.MPK("HzsECCX6RZ2ccbR3FarRSEfc5rkuETfywXJnRZut5JzU"),
		solana.MPK("ELfBngAgvLEHVBuJQhhE7AW6eqLX7id2sfrBngVNVAUW"),
		solana.MPK("Bx3ZhEBFedDqCBzuzKVS4eMKTtW1MmHkcMGU45FcyxRT"),
		solana.MPK("9W959DqEETiGZocYWCQPaJ6sBmUzgfxXfqGeTEdp3aQP"),
		solana.MPK("TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"),
		solana.MPK("6MxUhBLXHCqpdYaFPTmw1D9fQ7zYnm9grZyJvpGiqr15"),
		solana.MPK("SvWmpVVUkv8cwoRnBQ5Gqt2FFYjdpWLS665gE2ZLNQp"),
		solana.MPK("686KiYDMMkbredNoWx8yqvAdKSiHuWSG3dnbL6yWYmZp"),
		solana.MPK("9i14ZKzaDmzKCAQb8hCv4h5GCo2Xiq83JcL7bofk4Ddj"),
		solana.MPK("EorFh8siFyLF1QTZ7cCXQaPGqyo7eb4SAgKtRH8Jcxjd"),
		solana.MPK("6vWYnRDEHu7kRLbA2dnBgEfbdba72iDMDD9k3munyPaP"),
		solana.MPK("675kPX9MHTjS2zt1qfr1NYHuzeLXfQM9H24wFSUt1Mp8"),
		solana.MPK("TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"),
		solana.MPK("vVXfY15WdPsCmLvbiP4hWWECPFeAvPTuPNq3Q4BXfhy"),
		solana.MPK("5Q544fKrFoe6tsEbD7S8EmxGTJYAKtTVhAW5Q5pge4j1"),
		solana.MPK("A7RFkvmDFN4Qev8XgGAqSr5W75sNhhtCY3ZcGHZiDDo1"),
		solana.MPK("BKqBnj1TLpW4UEBbZn6aVoPLLBHDB6NTEL5nFNRqX7e7"),
		solana.MPK("AN7XxHrrcFL7629WySWVA2Tq9inczxkbE6YqgZ31rDnG"),
		solana.MPK("9xQeWvG816bUx9EPjHmaT23yvVM2ZWbrrpZb9PusVFin"),
		solana.MPK("72aW3Sgp1hMTXUiCq8aJ39DX2Jr7sZgumAvdLrLuCMLe"),
		solana.MPK("F3PQsAGiFf8fSySjUGgP3NQdAGSnioAThncyfd26GKZ3"),
		solana.MPK("6KyB4XprAw7Mgp1YMMsxRGx8T59Y5Lcu6s1FcwFrXy3i"),
		solana.MPK("Due4ZmGX2u7an9DPMvk3uX3sXYgngRatP1XmwzEgk1tT"),
		solana.MPK("8FMjC6yopBVYTXcYSGdFgoh6AFpwTdkJAGXxBeoV8xSq"),
		solana.MPK("5vgxuCqMn7DUt6Le6EGhdMzZjPQrtD1x4TD9zGw3mPte"),
		solana.MPK("FCZkJzztVTx6qKVec25jA3m4XjeGBH1iukGdDqDBHPvG"),
		solana.MPK("72aW3Sgp1hMTXUiCq8aJ39DX2Jr7sZgumAvdLrLuCMLe"),
		solana.MPK("7vtb8cULCnAdsfUKpex5v4CiFS2GwcsTzcK22m9BiDD5"),
		solana.MPK("Due4ZmGX2u7an9DPMvk3uX3sXYgngRatP1XmwzEgk1tT"),
		solana.MPK("F3PQsAGiFf8fSySjUGgP3NQdAGSnioAThncyfd26GKZ3"),
		solana.MPK("6KyB4XprAw7Mgp1YMMsxRGx8T59Y5Lcu6s1FcwFrXy3i"),
		solana.MPK("8FMjC6yopBVYTXcYSGdFgoh6AFpwTdkJAGXxBeoV8xSq"),
		solana.MPK("5vgxuCqMn7DUt6Le6EGhdMzZjPQrtD1x4TD9zGw3mPte"),
		solana.MPK("FCZkJzztVTx6qKVec25jA3m4XjeGBH1iukGdDqDBHPvG"),
		solana.MPK("9xQeWvG816bUx9EPjHmaT23yvVM2ZWbrrpZb9PusVFin"),
		solana.MPK("TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"),
		solana.MPK("SysvarRent111111111111111111111111111111111"),
		solana.MPK("BHzPYvC5J38kUeqkcUXwfraLWJ68cmGWm43ksF3i8bmk"),
		solana.MPK("9W959DqEETiGZocYWCQPaJ6sBmUzgfxXfqGeTEdp3aQP"),
		solana.MPK("TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"),
		solana.MPK("rxwsjytcEBvXpXrXBL1rpsjhoh78imBn8WbxjKmLRge"),
		solana.MPK("AcaxutE6Rh9vRxipTLdqinEdRK6R4ayUAAv2bZPh6UU9"),
		solana.MPK("6FRxhbY7bvSiDojPiqoidjTyDjxaUyCoPQk3ifEdfFbm"),
		solana.MPK("8aTapFecZRZmC2bTeKr2voHFW2twNvbrh8nWYdXYQWkZ"),
		solana.MPK("GMzPbaCuQmeMUm1opH3oSCgKUjVgJUW14myq99RVPGX5"),
		solana.MPK("7pPJGwd8Vq7aYmHaocQpQSfTn3UWYGKUgFkFhpMmRdDF"),
		solana.MPK("whirLbMiicVdio4qvUfM5KAg6Ct8VwpYzGff3uctyCc"),
		solana.MPK("TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"),
		solana.MPK("F7qyox3dAegTNfd8oBQD97LuCHWzQ9hSjbsF7Kv8kTNc"),
		solana.MPK("3NpsRa7H93FeGyT53KgeFNF4vX5m1YT5hxpUZJSpeUy1"),
		solana.MPK("4tS4d1j8vBeBU8zeHHo8sP7DUoNzVG24SZkHKGRNKXiT"),
		solana.MPK("DhU9gMt4gpnqseda43nXJjCssz266ivmTiJSRVu6P8Re"),
		solana.MPK("4VwhkiNBu2YpgvJjY6tANLXdtFtNY6jp5WPq4gY2PeUi"),
		solana.MPK("4rgdoZhrEbYrf9ZMXZmugMoYZ2XWiNPEAiKKujnEcjSv"),
		solana.MPK("faihchv9g9RwjcficyZVRPrQkzKA2NHUDL2PhgmxXPS"),
		solana.MPK("675kPX9MHTjS2zt1qfr1NYHuzeLXfQM9H24wFSUt1Mp8"),
		solana.MPK("TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"),
		solana.MPK("GQJjrG6f8HbxkE3ZVSRpzoyWhQ2RiivT68BybVK9DxME"),
		solana.MPK("5Q544fKrFoe6tsEbD7S8EmxGTJYAKtTVhAW5Q5pge4j1"),
		solana.MPK("AwHZdJrEDWAFhxsdsErvVPjWyE5JEY5Xq6cq4JjZX73L"),
		solana.MPK("3zrQ9od43vB9sV1MNbM68VnkLCfq9dVUvM1hmp8tcJNz"),
		solana.MPK("5odFuHq8jhqtNBKtCu4F2GvUiH5hB1zVfpS9XXbLf35d"),
		solana.MPK("9xQeWvG816bUx9EPjHmaT23yvVM2ZWbrrpZb9PusVFin"),
		solana.MPK("FrR9FBmiBjm2GjLZbfnCcgkbueUJ78NbBx1qcQKPUQe8"),
		solana.MPK("4W6ZoBB2QNBe6AYM6ofpWjerAsnJad93hVfdC5WMjRsX"),
		solana.MPK("64yfFmc7ivEknLRT2nvUmWkASGwz8MPxtcPdaiWUffro"),
		solana.MPK("GgJ8bQSZ6Lt2mEurrhzLMWFMzTgVFq8ax91QzmZzYiS6"),
		solana.MPK("9yg6VjgPUbojGn9d2n3UpX7B6gz7todGfTcV8apV5wkL"),
		solana.MPK("BDdh4ane6wXkRdbqUuMGYYR4ggf3GufUbjT2TxpHiAzU"),
		solana.MPK("A3LkbNQUjz1q3Ux5kQKCzNMFJw3yxk9qx1RtuQBXbZZe"),
		solana.MPK("9W959DqEETiGZocYWCQPaJ6sBmUzgfxXfqGeTEdp3aQP"),
		solana.MPK("TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"),
		solana.MPK("HdeYs4bpJKN2oTb7PHxbqq4kzKiLr772A5N2gWjY57ZT"),
		solana.MPK("2KRcBDQJWEPygxcMMFMvR6dMTVtMkJV6kbxr5e9Kdj5Q"),
		solana.MPK("DBckbD9CoRBFE8WdbbnFLDz6WdDDSZ7ReEeqdjL62fpG"),
		solana.MPK("B252w7ZkUX4WyLUJKLEymEpRkYMqJhgv2PSj2Z2LWH34"),
		solana.MPK("DRknxb4ZFxXUTG6UJ5HupNHG1SmvBSCPzsZ1o9gAhyBi"),
		solana.MPK("5XuLrZqpX9gW3pJw7274EYwAft1ciTXndU4on96ERi9J"),
		solana.MPK("675kPX9MHTjS2zt1qfr1NYHuzeLXfQM9H24wFSUt1Mp8"),
		solana.MPK("TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"),
		solana.MPK("5zUBVuXM3pwcrfi2Nj1mkT4RLKJjmBTjd4AsGs3biZBY"),
		solana.MPK("5Q544fKrFoe6tsEbD7S8EmxGTJYAKtTVhAW5Q5pge4j1"),
		solana.MPK("C5JCYfp6YE6JrpNkRoAGhARSUSMLMP7paPBHsiK1E5tb"),
		solana.MPK("DBMA8CUKosdnNvXT7phVDk8u9QCyNWnG4Z2twDS7ET17"),
		solana.MPK("EjBkXsDPGmyMQavnAQQsuMAMncDYTUAL35MhvyzEX4Kx"),
		solana.MPK("9xQeWvG816bUx9EPjHmaT23yvVM2ZWbrrpZb9PusVFin"),
		solana.MPK("9oXkdAWFyjDH8BbYrDVJ77r6GWPmUWo9ZYYpE25SZ2td"),
		solana.MPK("BdAZ2q9Ct9atgYnENKTsNKyLFhPXWiKuy3god4zEKMQW"),
		solana.MPK("6GBaKj1LncmZGS2B4uWjCM7pRZ9gUZWNTK7K8VCBvxaG"),
		solana.MPK("8aVP9P8cPzCSK4hdsVVk1E2nEf53f7iWkTKmsbidp4Fm"),
		solana.MPK("8N9HsqECLfZ7wHg7DW5WqYzN9UnWEgRhChG8ByNJ828Q"),
		solana.MPK("kvNtTHZU6vofnfdzYjN8G9gFqfjjf6yGYQJzwHb4m7h"),
		solana.MPK("5XKsQrPiQh1YznQFs9x8zMcqMSeJZBiGe7FmGfyQgC9N"),
		solana.MPK("9W959DqEETiGZocYWCQPaJ6sBmUzgfxXfqGeTEdp3aQP"),
		solana.MPK("TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"),
		solana.MPK("B76e3wtCDTKBgKQjvx87EBkDLPGcCY9w1SGiwjD5kaK7"),
		solana.MPK("FRUmMZDiZrDrwioiUYi3tdqF7SEBeT219bBu54PGxoCo"),
		solana.MPK("8voSogytL9jLgE73GS3WuujBinKFRQJjvUFsVGYexWZd"),
		solana.MPK("HEP7mACuN13cT95eDAYTNjgwriqJnMQVhnyRctqnBRe4"),
		solana.MPK("Df6XNHMF3uRVZnz7LCEGiZVax6rXgz76owtVkBHEjSb6"),
		solana.MPK("BCuRKfsM99LJFCchKUBLBZ26UuziDewJDRkkKMwx2qnd"),
		solana.MPK("675kPX9MHTjS2zt1qfr1NYHuzeLXfQM9H24wFSUt1Mp8"),
		solana.MPK("TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"),
		solana.MPK("2Tv6eMih3iqxHrLAWn372Nba4A8FT8AxFSbowBmmTuAd"),
		solana.MPK("5Q544fKrFoe6tsEbD7S8EmxGTJYAKtTVhAW5Q5pge4j1"),
		solana.MPK("GNHftHYD7WRG5HYdyWjd9KsxjUgUALrLcSG2AZvv5ahU"),
		solana.MPK("9ZQNgn9zAc9oLKST5yW9PNjCCqSfJVwnFpfgZnd88Xn1"),
		solana.MPK("HLtqBqwgdbGdFfd5UZtKkvrdxLLcpaMnAJ5aZAzDjFdT"),
		solana.MPK("9xQeWvG816bUx9EPjHmaT23yvVM2ZWbrrpZb9PusVFin"),
		solana.MPK("9cuBrXXSH9Uw51JB9odLqEyeF5RQSeRpcfXbEW2L8X6X"),
		solana.MPK("EmqbfgZFSQxAeJRWKrrBVST4oLsq8aMt4WtcufPARcd7"),
		solana.MPK("GZqx3xX1PjNpmw2qDGhiUSa6PsM5tWYY7cMmKzYFCCLD"),
		solana.MPK("8w8JzuqcRUm9QAC3YWJm2mBCVjWDLXh8b7ktSouJKMUd"),
		solana.MPK("8DGcP5Z8M878mguFLohaK9jFrrShDCREF3qa7JhMfgib"),
		solana.MPK("CLS4WFje2PbV3MmV4v7CGxu3bNFqx2sYewq95rzGR8t8"),
		solana.MPK("FBLtcfAXmm5PpJLLr95L5cjfgbpJiGHsWdBXDpC2TBQ2"),
		solana.MPK("9W959DqEETiGZocYWCQPaJ6sBmUzgfxXfqGeTEdp3aQP"),
		solana.MPK("TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"),
		solana.MPK("5DnwMqYAGEKekYXJdN8Bue6vN1p5zrEnBpmd53jEK61S"),
		solana.MPK("4NfadURWeSDPJBGcKQRt39mPhbG9M7EJx6FZDwwcFB9f"),
		solana.MPK("6d19CQA1FP2MLLAzA7XoZEc9Agc32FaKUS175UVWLGtv"),
		solana.MPK("HpPnUHyo19VjmXbP6FbbKXu7WQCUEn6h7be76fZdHVmf"),
		solana.MPK("qJxKN9BhxbYvRNbjfK2uAVWboto6sonj8XC1ZEW5XTB"),
		solana.MPK("57L2bEFecsAv4jnaM2PBaeAVyPZEYtTmXBi7eaG2xWXw"),
		solana.MPK("675kPX9MHTjS2zt1qfr1NYHuzeLXfQM9H24wFSUt1Mp8"),
		solana.MPK("TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"),
		solana.MPK("FEFzBbbEK8yDigqyJPgJKMR5X1xZARC25QTCskvudjuK"),
		solana.MPK("5Q544fKrFoe6tsEbD7S8EmxGTJYAKtTVhAW5Q5pge4j1"),
		solana.MPK("8PAAfUWoVsSotWUGrL6CJCT2sApMpE2hn8DGWXq4y9Gs"),
		solana.MPK("AZPsv6tY1HQjmeps2sMje5ysNtPKsfbtxj5Qw3jcya1a"),
		solana.MPK("9D6JfNjyi6dXBYGErxmXmezkauPJdHW4KjMr2RGyD86Y"),
		solana.MPK("9xQeWvG816bUx9EPjHmaT23yvVM2ZWbrrpZb9PusVFin"),
		solana.MPK("BHqcTEDhCoZgvXcsSbwnTuzPdxv1HPs6Kz4AnPpNrGuq"),
		solana.MPK("F61FtHm4R4F1gszB3FuwDPvXeSPQwNmHTofoYCnrV4FY"),
		solana.MPK("5tYcHCW3ZZK4TMUSYiTi4dEE7iefyQ9dE17XDDAmDf92"),
		solana.MPK("C5gcq3kmmXJ6ADWvH3Pc8bpiBQCL5cx4ypRwPg5xxFFx"),
		solana.MPK("6sF1TAJjfrNucAqaQFRrMD78z2RinTGeyo4KsXPbwiqh"),
		solana.MPK("5iXoDYXGnMxEwL65XTJHWdr6Z2UD5qq47ZijW24VSSSQ"),
		solana.MPK("BuRLkxJffwznEsxXEqmXZJdLh4vQ1BRXc41sT6BtPV4X"),
		solana.MPK("9W959DqEETiGZocYWCQPaJ6sBmUzgfxXfqGeTEdp3aQP"),
		solana.MPK("TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"),
		solana.MPK("ErWwp9HKjk5ZPLDt8SrHKH5PvSKTwFDdFo5E3zDuE5Be"),
		solana.MPK("GYY1t5d4pZnJC4rMXGY9yKMyCzLqxRqbtSguD2KkxghH"),
		solana.MPK("GEtZSc8188t2cCAv21UGCyjvxCeyU5Co99GtRtyTkpdh"),
		solana.MPK("Bi95f8H7o7zHWuYysxDHEubPv4c3NhsHWhaesXJu91NC"),
		solana.MPK("GBijunwxa4Ni3JmYC6q6zgaVhSUJU6hVX5qTyJDRpNTc"),
		solana.MPK("6NhybmW42rdWj5TcobNKQT6JaZispgngcfTDrCsgVq4Q"),
		solana.MPK("9W959DqEETiGZocYWCQPaJ6sBmUzgfxXfqGeTEdp3aQP"),
		solana.MPK("TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"),
		solana.MPK("v51xWrRwmFVH6EKe8eZTjgK5E4uC2tzY5sVt5cHbrkG"),
		solana.MPK("3Kk8rpjxpc9qv2pJPr1CbmyQqrTDPntpryXActLogQeD"),
		solana.MPK("5c4tzhRVaCxpmu8o3HrEZ8PWBDKSR6QNkBdQrUo9oe3e"),
		solana.MPK("AFNaWHH7ZGFjB7y7jmPM7jVs7QBAciffu7Z5tZidRHPR"),
		solana.MPK("5a6Y1ephcbKSoyLMQyD1JWbtqawCy8p2FtRL9v3zhaG5"),
		solana.MPK("ACKiRmbiMaPEc73pz4dVMuJGPaa74Vx9sfYADjnHuzvo"),
		solana.MPK("whirLbMiicVdio4qvUfM5KAg6Ct8VwpYzGff3uctyCc"),
		solana.MPK("TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"),
		solana.MPK("CeGZNd2ap2ke8JreVqDXikSrxhnSERh9G8WAt4yqMcBs"),
		solana.MPK("GhZVE3cbEtbAJMb4QbYEEK4BQXcDeyPCWmr8yzZL9NH7"),
		solana.MPK("BJW4t9VsS7W5W7UxXhjGj3C3r4sB9ZzTXMcUD3WVq511"),
		solana.MPK("AjYCGnsaCsoif9LU4Dr3PRN2EJucjsn1DduycybVwzvL"),
		solana.MPK("GC83zEFAxq6mR6HZgPMwTf4oThc7RPsjQFfopEn2R1Rf"),
		solana.MPK("3v3QYAnnGjeGo8K2rsqeKYegCB1YWb7QC58JXKdaGGoD"),
		solana.MPK("4BKxA3d67Sfa3xhjg6waEo58ACB2XkFESXDL5RYFHsP4"),
		solana.MPK("9W959DqEETiGZocYWCQPaJ6sBmUzgfxXfqGeTEdp3aQP"),
		solana.MPK("TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"),
		solana.MPK("7c2CLgatf2TU36PgpS65WLmvWk94rmaHVf1Z1peZ7mcA"),
		solana.MPK("Mq46N9EknnxHL9fRkJhS4Eg9YXRifHiWzFJTD11ePWC"),
		solana.MPK("5rU6M2jAXQMSmgrsn14BPoVVhoBdCU6y5cP7XMjN4ZYy"),
		solana.MPK("D28rzq246bcXBrYiCeALY86y8NwvCUmuJGNggvKsh4WR"),
		solana.MPK("Hmfrtmo93DpSDmVNLQKcBS5D1ia5JatiRSok9ososubz"),
		solana.MPK("7JH76Kw4dHyC5szRXkx6MFkJ3BEViodfNy15uFJst1cX"),
		solana.MPK("9W959DqEETiGZocYWCQPaJ6sBmUzgfxXfqGeTEdp3aQP"),
		solana.MPK("TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"),
		solana.MPK("C3b5AWQJiyar5g8EWu75zgDE26F55ZJWpqtFVCCVDQQQ"),
		solana.MPK("3SphkwoHx3d13Eu9RehVVg4gGMZv7FEaDXvPqWbQF9bm"),
		solana.MPK("5AhPVbtyiTV3SiNRJuq5z9xeaqqwoHQWqohR9HvjJkKS"),
		solana.MPK("6mtcbtTAadVEdnWJZmsq8woqLea7ef7k5WumVXSHr5KQ"),
		solana.MPK("7vnps4VE5RTGAr5fmPZu7fSrk2VnM4Up838grZfqmxqE"),
		solana.MPK("CT95CSNqi4nttNW84dDuA8Um7FLAC52PVUvuVRKeCHVK"),
		solana.MPK("9W959DqEETiGZocYWCQPaJ6sBmUzgfxXfqGeTEdp3aQP"),
		solana.MPK("TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"),
		solana.MPK("GumfURfQvPaJ2E5ueCEKYJmymNQbV34gU9TmiKZYRkiv"),
		solana.MPK("EjoLNSDggfWWE7BxwoL4tJHBEg1cFpdiyKeYTYCec2o2"),
		solana.MPK("AGNHgSQuPd4EqjLTLJrXEVb3KCkjRxGVDTaag4drV1XX"),
		solana.MPK("FESKk2kj9oqdYR4dcaP4LyqDyWZt3NttgypRVFoyUQNs"),
		solana.MPK("55r9txzQtmjTykmTXmBYZCVMg5z9squB8b5cSw2AhxA4"),
		solana.MPK("F6xCTe256cA6HTX5CYBkDtXoruHvjfbxeHNeqR9kR7oJ"),
		solana.MPK("9W959DqEETiGZocYWCQPaJ6sBmUzgfxXfqGeTEdp3aQP"),
		solana.MPK("TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"),
		solana.MPK("9tK2LaapwjxaUmfcAzY9zgC39M3wnaFX558y2Bb4oxWG"),
		solana.MPK("EGuBsx6HAgAtf1ogzF1uXTUQgwRex61hnhvuZcMwQKUJ"),
		solana.MPK("GgfTGZ5DnAotnXKFM86vqffKQZ9nGgHaX1PDS7RTcKjQ"),
		solana.MPK("A3rzsPGtqowjKXfscYrPo1jvv2EVYpJwXQPGKxgvvStf"),
		solana.MPK("DSiHyHDn96bUQSZtizyCRLcQzrwohZeMpVu8rYJN1HzG"),
		solana.MPK("Acom6ebnmbFKQk3XeX5VHPiz8bd7kzfpUMsqHKJDJnry"),
		solana.MPK("675kPX9MHTjS2zt1qfr1NYHuzeLXfQM9H24wFSUt1Mp8"),
		solana.MPK("TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"),
		solana.MPK("9xUWbM7zXsccied2jNXama1Z1Wh9mwn9APX1drRTPtvh"),
		solana.MPK("5Q544fKrFoe6tsEbD7S8EmxGTJYAKtTVhAW5Q5pge4j1"),
		solana.MPK("JD51bY2uLtwgzYQNYjF7m1UvWX5HdHE7orMrxogPQk1G"),
		solana.MPK("7TeWuw6WxwqLkadHGRsLFVWoe4zb9snMRZHH5nQPpUPV"),
		solana.MPK("A59Pg8yemxDqUqfvfmh6e9Wmkr74v7uGeygcUkQCSoLJ"),
		solana.MPK("9xQeWvG816bUx9EPjHmaT23yvVM2ZWbrrpZb9PusVFin"),
		solana.MPK("9fe1MWiKqUdwift3dEpxuRHWftG72rysCRHbxDy6i9xB"),
		solana.MPK("4ZwKetX2m7fS3gigLa215xjveVNwLWAVJeh1zaQUbpuF"),
		solana.MPK("vL2N5k5PS67MctE1Tj5u3sivBNMj6EvejskPiqtDP6n"),
		solana.MPK("ApWLqV2xjdn2FEjYvVgf7Ltp5by9TDVEnpg5dXrZzY8k"),
		solana.MPK("26h5i4vYPinyUZ6kUp8tzhzvQtP3cNzhzaBMqySybNMF"),
		solana.MPK("7E5CzVnTFsTnwPqoJ8uUA8RNqCgsYy6ZEnRVmz7LURaA"),
		solana.MPK("FPC75yXyJwF3NFEmgHrJRDNmXnukpVQgXayZVsmpEDKo"),
	}

	require.Equal(t, len(expectedKeys), len(table.Addresses), "decoded addresses length mismatch")

	for i := range expectedKeys {
		require.Equal(t, expectedKeys[i], table.Addresses[i], "key %d; extected %s, got %s", i, expectedKeys[i], table.Addresses[i])
	}
	{
		buf := new(bytes.Buffer)
		enc := bin.NewBinEncoder(buf)
		err := table.MarshalWithEncoder(enc)
		require.NoError(t, err)
		require.Equal(t, buf.Bytes(), tableAccountBytes)
	}
	{
		// https://explorer.solana.com/tx/24jRjMP3medE9iMqVSPRbkwfe9GdPmLfeftKPuwRHZdYTZJ6UyzNMGGKo4BHrTu2zVj4CgFF3CEuzS79QXUo2CMC
		txBase64 := "ATU8IfVPwTFbXJvq0sO8w7Xc6/nc/4RFux7ehibunps+JjNSczZbue4bPn6uR9s6aWSZQCP8brf8RyYUwZQ0DQeAAQACBVTKAWVY3LzWDZTdfErhZpxix74Qyp+LjLmlvAPS/l4z4Rt91U7dxP1bjyHkoY3vBWo/XAqjvIzK8DTSvetevZWcVdOLU3j3+e/xY8bnWyZ3eMohRZiHw3bH7GrsEzYzowR51S3tv2vF7NCdhFNKNK6ll1BDs2/QKyRlC7WEQ1lcAwZGb+UhFzL/7K26csOb57yM5bvF9xJrLEObOkAAAACgpxPJ+57U2CTjvooWuPN91CltQch0oj/O5i3Fa45CgQMDIBMAARQTBRUGBwgWCQoLDA0OFwECABgTGRoAAg8QARESJOUXy5d6460qAAIAAAACBwIDgJaYAAAAAACAlpgAAAAAAAAAAAQABQLAJwkABAAJA0CcAAAAAAAAAaC/OwMGMFZGentK87meSfdqah4dRR/iw+wKg/yvZNw5DhIUFRYYGRobHB0MDQ4PCAEQExceAAoL"
		tx := &solana.Transaction{}
		err := tx.UnmarshalBase64(txBase64)
		require.NoError(t, err)
		err = tx.Message.SetAddressTables(map[solana.PublicKey]solana.PublicKeySlice{
			solana.MPK("BpVMhYJB14QX5pXfbHRxB8vmpW4AFodWjBTDvfCJwsfv"): table.Addresses,
		})
		require.NoError(t, err)
		require.True(t, tx.Message.IsSigner(solana.MPK("6hyuGqKQyhAEipjtaquiNHfd1dVjrNT3FzzanXurbK4W")))

		require.Equal(t, solana.PublicKeySlice{
			solana.MPK("BpVMhYJB14QX5pXfbHRxB8vmpW4AFodWjBTDvfCJwsfv"),
		}, tx.Message.GetAddressTableLookups().GetTableIDs())

		metas, err := tx.Message.AccountMetaList()
		require.NoError(t, err)

		{
			writable := solana.PublicKeySlice{
				//
				solana.MPK("G9j3eZtj5pprqxvYz6v5WoyqhEMbnV6m9KTtoZuNrvCk"),
				solana.MPK("BXGWMYBEsuXsqt4R19ihHsjhL81KvVXbDvw5TxTfD7sx"),
				// from lookups:
				solana.MPK("ABrn4ED4AvkQ79VAXqf7ooqicJPHhZDAbC9rqcQ8ePzz"),
				solana.MPK("D7CHbxSFSiEW3sPc486AGDwuwsmyZqhP7stG4Yo9ZHTC"),
				solana.MPK("5o8dopjEKEy491bVHShtG6KSSHKm2JUugVqKEK7Jw7YF"),
				solana.MPK("FN3wMZUuWkM65ZtcnAoYpsq773YxrnMfM5iAroSGttBo"),
				solana.MPK("6uGUx583UHvFKKCnoMfnGNEFxhSWy5iXXyea4o5E9dx7"),
				solana.MPK("Gp7wpKu9mXpxdykMD9JKW5SK2Jw1h2fttxukvcL2dnW6"),
				solana.MPK("4mkSxT9MaUsUd5uSkZxohf1pbPByk7b5ptWpu4ZABvto"),
				solana.MPK("4dDEjb4JZejtweFEJjjqqC5wwZi3jqtzoS7cPNRyPoT6"),
				solana.MPK("Geoh8p8j48Efupens8TqJKj491aqk5VhPXABFAqGtAjr"),
				solana.MPK("EVv4jPvUxbugw8EHTDwkNBboE26DiN4Zy1CQrd5j3Sd4"),
				solana.MPK("ErcxwkPgLdyoVL6j2SsekZ5iysPZEDRGfAggh282kQb8"),
				solana.MPK("EFYW6YEiCGpavuMPS1zoXhgfNkPisWkQ3bQz1b4UfKek"),
				solana.MPK("BVWwyiHVHZQMPHsiW7dZH7bnBVKmbxdeEjWqVRciHCyo"),
				solana.MPK("E6aTzkZKdCECgpDtBZtVpqiGjxRDSAFh1SC9CdSoVK7a"),
			}
			for _, acc := range writable {
				is, err := tx.Message.IsWritable(acc)
				require.NoError(t, err, "account %s", acc)
				require.True(t, is, "account %s must be writable", acc)
			}
			for _, acc := range writable {
				has, err := tx.Message.HasAccount(acc)
				require.NoError(t, err)
				require.True(t, has)
			}
		}
		{
			readonly := solana.PublicKeySlice{
				solana.MPK("JUP4Fb2cqiRUcaTHdrPC8h2gNsA2ETXiPDD33WcGuJB"),
				solana.MPK("ComputeBudget111111111111111111111111111111"),
				// from lookups:
				solana.MPK("TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"),
				solana.MPK("675kPX9MHTjS2zt1qfr1NYHuzeLXfQM9H24wFSUt1Mp8"),
				solana.MPK("5Q544fKrFoe6tsEbD7S8EmxGTJYAKtTVhAW5Q5pge4j1"),
				solana.MPK("9xQeWvG816bUx9EPjHmaT23yvVM2ZWbrrpZb9PusVFin"),
				solana.MPK("3ceGkbGkqQwjJsZEYzjykDcWM1FjzHGMNTyKHD1c7kqW"),
				solana.MPK("9W959DqEETiGZocYWCQPaJ6sBmUzgfxXfqGeTEdp3aQP"),
				solana.MPK("8JPid6GtND2tU3A7x7GDfPPEWwS36rMtzF7YoHU44UoA"),
				solana.MPK("749y4fXb9SzqmrLEetQdui5iDucnNiMgCJ2uzc3y7cou"),
			}
			for _, acc := range readonly {
				is, err := tx.Message.IsWritable(acc)
				require.NoError(t, err, "account %s", acc)
				require.False(t, is, "account %s must be readonly", acc)
			}
			for _, acc := range readonly {
				has, err := tx.Message.HasAccount(acc)
				require.NoError(t, err)
				require.True(t, has)
			}
		}
		{
			ix := tx.Message.Instructions[0]
			got, err := tx.Message.ResolveProgramIDIndex(ix.ProgramIDIndex)
			require.NoError(t, err)
			require.Equal(t, solana.MPK("JUP4Fb2cqiRUcaTHdrPC8h2gNsA2ETXiPDD33WcGuJB"), got)
		}
		{
			ix := tx.Message.Instructions[1]
			got, err := tx.Message.ResolveProgramIDIndex(ix.ProgramIDIndex)
			require.NoError(t, err)
			require.Equal(t, solana.MPK("ComputeBudget111111111111111111111111111111"), got)
		}
		{
			ix := tx.Message.Instructions[2]
			got, err := tx.Message.Program(ix.ProgramIDIndex)
			require.NoError(t, err)
			require.Equal(t, solana.MPK("ComputeBudget111111111111111111111111111111"), got)
		}
		{
			has, err := tx.Message.HasAccount(solana.SysVarClockPubkey)
			require.NoError(t, err)
			require.False(t, has)
		}
		{
			acc, err := tx.Message.Account(0)
			require.NoError(t, err)
			require.Equal(t, solana.MPK("6hyuGqKQyhAEipjtaquiNHfd1dVjrNT3FzzanXurbK4W"), acc)
			require.Equal(t, solana.MPK("6hyuGqKQyhAEipjtaquiNHfd1dVjrNT3FzzanXurbK4W"), metas[0].PublicKey)
		}
		{
			acc, err := tx.Message.Account(1)
			require.NoError(t, err)
			require.Equal(t, solana.MPK("G9j3eZtj5pprqxvYz6v5WoyqhEMbnV6m9KTtoZuNrvCk"), acc)
			require.Equal(t, solana.MPK("G9j3eZtj5pprqxvYz6v5WoyqhEMbnV6m9KTtoZuNrvCk"), metas[1].PublicKey)
		}
		{
			acc, err := tx.Message.Account(15)
			require.NoError(t, err)
			require.Equal(t, solana.MPK("ErcxwkPgLdyoVL6j2SsekZ5iysPZEDRGfAggh282kQb8"), acc)
			require.Equal(t, solana.MPK("ErcxwkPgLdyoVL6j2SsekZ5iysPZEDRGfAggh282kQb8"), metas[15].PublicKey)
		}
		{
			acc, err := tx.Message.Account(18)
			require.NoError(t, err)
			require.Equal(t, solana.MPK("E6aTzkZKdCECgpDtBZtVpqiGjxRDSAFh1SC9CdSoVK7a"), acc)
			require.Equal(t, solana.MPK("E6aTzkZKdCECgpDtBZtVpqiGjxRDSAFh1SC9CdSoVK7a"), metas[18].PublicKey)
		}
		{
			acc, err := tx.Message.Account(21)
			require.NoError(t, err)
			require.Equal(t, solana.MPK("5Q544fKrFoe6tsEbD7S8EmxGTJYAKtTVhAW5Q5pge4j1"), acc)
			require.Equal(t, solana.MPK("5Q544fKrFoe6tsEbD7S8EmxGTJYAKtTVhAW5Q5pge4j1"), metas[21].PublicKey)
		}
		{
			acc, err := tx.Message.Account(26)
			require.NoError(t, err)
			require.Equal(t, solana.MPK("749y4fXb9SzqmrLEetQdui5iDucnNiMgCJ2uzc3y7cou"), acc)
			require.Equal(t, solana.MPK("749y4fXb9SzqmrLEetQdui5iDucnNiMgCJ2uzc3y7cou"), metas[26].PublicKey)
		}
		{
			acc, err := tx.Message.Account(9999)
			require.Error(t, err)
			require.Equal(t, solana.PublicKey{}, acc)
		}
		require.True(t, tx.Message.IsVersioned())
		{
			{
				lookups := tx.Message.GetAddressTableLookups()
				require.Equal(t, 1, len(lookups))
				first := lookups[0]
				require.Equal(t,
					solana.MessageAddressTableLookup{
						AccountKey:      solana.MPK("BpVMhYJB14QX5pXfbHRxB8vmpW4AFodWjBTDvfCJwsfv"),
						WritableIndexes: []uint8{18, 20, 21, 22, 24, 25, 26, 27, 28, 29, 12, 13, 14, 15},
						ReadonlyIndexes: []uint8{1, 16, 19, 23, 30, 0, 10, 11},
					}, first)
			}
		}
		{
			got, err := tx.Message.GetAllKeys()
			require.NoError(t, err)
			require.Equal(t, metas.GetKeys(), got)
			{
				// before resolution:
				static := solana.PublicKeySlice{
					solana.MPK("6hyuGqKQyhAEipjtaquiNHfd1dVjrNT3FzzanXurbK4W"),
					solana.MPK("G9j3eZtj5pprqxvYz6v5WoyqhEMbnV6m9KTtoZuNrvCk"),
					solana.MPK("BXGWMYBEsuXsqt4R19ihHsjhL81KvVXbDvw5TxTfD7sx"),
					solana.MPK("JUP4Fb2cqiRUcaTHdrPC8h2gNsA2ETXiPDD33WcGuJB"),
					solana.MPK("ComputeBudget111111111111111111111111111111"),
				}
				require.Equal(t, 5, len(tx.Message.AccountKeys))
				for i, acc := range static {
					require.Equal(t, acc, tx.Message.AccountKeys[i], "static key %d", i)
				}
				{ // serialization before resolution:
					{
						got, err := tx.ToBase64()
						require.NoError(t, err)
						require.Equal(t, txBase64, got)
					}
					{
						buf := new(bytes.Buffer)
						err := table.MarshalWithEncoder(bin.NewBinEncoder(buf))
						require.NoError(t, err)
						require.Equal(t, tableAccountBytes, buf.Bytes())
					}
					{
						require.Equal(t, 14, tx.Message.GetAddressTableLookups().NumWritableLookups())
						require.Equal(t, 22, tx.Message.GetAddressTableLookups().NumLookups())
						require.Equal(t, 22, tx.Message.NumLookups())
					}
				}
				// after resolution:
				err = tx.Message.ResolveLookups()
				require.NoError(t, err)
				require.Equal(t, 27, len(tx.Message.AccountKeys))
				{
					// serialization after resolution:
					{
						got, err := tx.ToBase64()
						require.NoError(t, err)
						require.Equal(t, txBase64, got)
					}
					{
						buf := new(bytes.Buffer)
						err := table.MarshalWithEncoder(bin.NewBinEncoder(buf))
						require.NoError(t, err)
						require.Equal(t, tableAccountBytes, buf.Bytes())
					}
					{
						require.Equal(t, 14, tx.Message.GetAddressTableLookups().NumWritableLookups())
						require.Equal(t, 22, tx.Message.GetAddressTableLookups().NumLookups())
						require.Equal(t, 22, tx.Message.NumLookups())
					}
				}
				// after resolution:
				err = tx.Message.ResolveLookups()
				require.NoError(t, err)
				require.Equal(t, 27, len(tx.Message.AccountKeys))
				{
					// same as before first resolution call:
					{
						got, err := tx.ToBase64()
						require.NoError(t, err)
						require.Equal(t, txBase64, got)
					}
					{
						buf := new(bytes.Buffer)
						err := table.MarshalWithEncoder(bin.NewBinEncoder(buf))
						require.NoError(t, err)
						require.Equal(t, tableAccountBytes, buf.Bytes())
					}
					{
						require.Equal(t, 14, tx.Message.GetAddressTableLookups().NumWritableLookups())
						require.Equal(t, 22, tx.Message.GetAddressTableLookups().NumLookups())
						require.Equal(t, 22, tx.Message.NumLookups())
					}
				}
			}
		}
	}
}
