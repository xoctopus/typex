package testdata

type (
	AliasInt                         = int
	AliasUnion                       = SignedInteger
	AliasSerialized                  = Serialized[[]byte]
	AliasWithTArg[X CanBeSerialized] = Serialized[X]
)
