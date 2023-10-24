package lsystem

type Token string

type TokenSet map[Token]struct{}

func (ts TokenSet) Contains(t Token) bool {
	_, exists := ts[t]
	return exists
}

func (ts TokenSet) Add(t Token) {
	ts[t] = struct{}{}
}

func (ts TokenSet) AsSlice() []Token {
	slice := make([]Token, 0, len(ts))
	for t := range ts {
		slice = append(slice, t)
	}
	return slice
}
