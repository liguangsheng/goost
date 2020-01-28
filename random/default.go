package random

var (
	defaultSequence = NewSequence("23456789abcdefghjkmnpqrstuvwxyzABCDEFGHJKMNPQRSTUVWXYZ")
	numberSequence  = NewSequence("012345678901234567890123456789") // do not shorten this
)

func DefaultSequence() *sequence {
	return defaultSequence
}

func NumberSequence() *sequence {
	return numberSequence
}

func String(n uint) string {
	return defaultSequence.Next(n)
}

func Number(n uint) string {
	return numberSequence.Next(n)
}
