package lsystem

type Buffer struct {
	BytePairs []BytePair
	Len       int
	Cap       int
}

func (m *Buffer) Append(bp BytePair) {
	if m.Len >= m.Cap {
		m.Grow()
	}

	m.BytePairs[m.Len] = bp
	m.Len++
}

func (m *Buffer) AppendSlice(bps []BytePair) {
	if m.Len+len(bps) > m.Cap {
		m.Grow()
	}

	copy(m.BytePairs[m.Len:], bps)
	m.Len += len(bps)
}

func (m *Buffer) Grow() {
	newCap := m.Cap * 2
	m.Cap = newCap

	newSlice := make([]BytePair, newCap)
	copy(newSlice, m.BytePairs)
	m.BytePairs = newSlice
}

const threadCount = 4

type MemPool struct {
	readBuffers  [threadCount]*Buffer
	writeBuffers [threadCount]*Buffer

	swap [threadCount]bool
}

func NewMemPool(capacity int) *MemPool {
	readBuffers := [threadCount]*Buffer{}
	writeBuffers := [threadCount]*Buffer{}
	swapValues := [threadCount]bool{}

	for i := 0; i < threadCount; i++ {
		readBuffers[i] = &Buffer{
			BytePairs: make([]BytePair, capacity),
			Len:       0,
			Cap:       capacity,
		}

		writeBuffers[i] = &Buffer{
			BytePairs: make([]BytePair, capacity),
			Len:       0,
			Cap:       capacity,
		}
	}

	return &MemPool{
		readBuffers:  readBuffers,
		writeBuffers: writeBuffers,

		swap: swapValues,
	}
}

func (m *MemPool) GetReadBuffer(idx int) *Buffer {
	if m.swap[idx] {
		return m.writeBuffers[idx]
	}
	return m.readBuffers[idx]
}

func (m *MemPool) GetWriteBuffer(idx int) *Buffer {
	if m.swap[idx] {
		return m.readBuffers[idx]
	}
	return m.writeBuffers[idx]
}

func (m *MemPool) SwapAll() {
	for i := 0; i < threadCount; i++ {
		m.swap[i] = !m.swap[i]
		writeBuf := m.GetWriteBuffer(i)
		writeBuf.Len = 0
	}
}

func (m *MemPool) Swap(idx int) {
	m.swap[idx] = !m.swap[idx]
	writeBuf := m.GetWriteBuffer(idx)
	writeBuf.Len = 0
}

func (m *MemPool) Reset() {
	for i := 0; i < threadCount; i++ {
		readBuf := m.GetReadBuffer(i)
		readBuf.Len = 0

		writeBuf := m.GetWriteBuffer(i)
		writeBuf.Len = 0

		m.swap[i] = false
	}
}

func (m *MemPool) ReadAll() []BytePair {
	tokens := []BytePair{}
	for i := 0; i < threadCount; i++ {
		buf := m.GetReadBuffer(i)
		tokens = append(tokens, buf.BytePairs[:buf.Len]...)
	}

	return tokens
}
