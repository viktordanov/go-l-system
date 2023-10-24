package lsystem

type Buffer struct {
	BytePairs []BytePair
	Len       int
	Cap       int
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
