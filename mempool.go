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

type BufferPool struct {
	active   *Buffer
	inactive *Buffer

	swap bool
}

func NewBufferPool(capacity int) *BufferPool {
	return &BufferPool{
		active: &Buffer{
			BytePairs: make([]BytePair, capacity),
			Len:       0,
			Cap:       capacity,
		},
		inactive: &Buffer{
			BytePairs: make([]BytePair, capacity),
			Len:       0,
			Cap:       capacity,
		},

		swap: false,
	}
}

func (m *BufferPool) Reset() {
	m.active.Len = 0
	m.inactive.Len = 0
	m.swap = false
}

func (m *BufferPool) GetActive() *Buffer {
	if m.swap {
		return m.inactive
	}
	return m.active
}

func (m *BufferPool) Append(bp BytePair) {
	active := m.GetActive()

	if active.Len >= m.active.Cap {
		m.Grow()
	}

	active.BytePairs[active.Len] = bp
	active.Len++
}

func (m *BufferPool) AppendSlice(bps []BytePair) {
	active := m.GetActive()

	if active.Len+len(bps) > m.active.Cap {
		m.Grow()
	}

	copy(active.BytePairs[active.Len:], bps)
	active.Len += len(bps)
}

func (m *BufferPool) GetLen() int {
	return m.GetActive().Len
}

func (m *BufferPool) GetCap() int {
	return m.GetActive().Cap
}

func (m *BufferPool) GetSwap() *Buffer {
	if m.swap {
		return m.active
	}
	return m.inactive
}

func (m *BufferPool) Grow() {
	newCap := m.active.Cap * 2
	m.active.Cap = newCap
	m.inactive.Cap = newCap

	if m.swap {
		newSlice := make([]BytePair, newCap)
		copy(newSlice, m.inactive.BytePairs)
		m.inactive.BytePairs = newSlice
		m.active.BytePairs = make([]BytePair, newCap)
	} else {
		newSlice := make([]BytePair, newCap)
		copy(newSlice, m.active.BytePairs)
		m.active.BytePairs = newSlice
		m.inactive.BytePairs = make([]BytePair, newCap)
	}
}

func (m *BufferPool) Swap() {
	m.swap = !m.swap
}

func (m *BufferPool) ResetWritingHead() {
	m.GetActive().Len = 0
}

type MemPool struct {
	readBuffers  [4]*Buffer
	writeBuffers [4]*Buffer

	swap bool
}

func NewMemPool(capacity int) *MemPool {
	readBuffers := [4]*Buffer{}
	writeBuffers := [4]*Buffer{}

	for i := 0; i < 4; i++ {
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

		swap: false,
	}
}

func (m *MemPool) GetReadBuffer(idx int) *Buffer {
	if m.swap {
		return m.writeBuffers[idx]
	}
	return m.readBuffers[idx]
}

func (m *MemPool) GetWriteBuffer(idx int) *Buffer {
	if m.swap {
		return m.readBuffers[idx]
	}
	return m.writeBuffers[idx]
}

func (m *MemPool) Swap() {
	m.swap = !m.swap
}
