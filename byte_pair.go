package lsystem

type BytePair uint16

func (bt BytePair) First() uint8 {
	return uint8(bt >> 8)
}

func (bt BytePair) Second() uint8 {
	return uint8(bt & 0xFF)
}

func NewBytePair(first, second uint8) BytePair {
	return BytePair(uint16(first)<<8 | uint16(second))
}
