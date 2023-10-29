package lsystem

type TokenStateId uint8

func (bt TokenStateId) TokenId() uint8 {
	return uint8(bt & 0x7F)
}

func (bt TokenStateId) HasParam() bool {
	return bt&0x80 != 0
}

func NewTokenStateId(tokenId uint8, hasParam bool) TokenStateId {
	if hasParam {
		tokenId |= uint8(0x80)
	}

	return TokenStateId(tokenId)
}
