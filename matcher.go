package fifd

type Matcher struct {
	reader       *Reader
	userAgent    string
	node         *Node
	power        uint32
	hash         uint32
	drift        int
	difference   int
	currentIndex int
	firstIndex   int
	lastIndex    int
	deviceIndex  int
}

func (m *Matcher) getMatchingHashFromListNodeSearch() *NodeHash {
	hashes := m.node.Hashes
	upper := m.node.HashesCount - 1
	lower := int32(0)
	for lower <= upper {
		middle := lower + (upper-lower)/2
		if hashes[middle].HashCode == m.hash {
			return &hashes[middle]
		} else if hashes[middle].HashCode > m.hash {
			upper = middle - 1
		} else {
			lower = middle + 1
		}
	}
	return nil
}

func (m *Matcher) getMatchingHashFromListNodeTable() *NodeHash {
	hashes := m.node.Hashes
	index := int(m.hash % m.node.Modulo)
	if m.hash == hashes[index].HashCode {
		return &hashes[index]
	}
	if hashes[index].HashCode == 0 && hashes[index].NodeOffset > 0 {
		index = int(hashes[index].NodeOffset)
		for hashes[index].HashCode != 0 {
			if m.hash == hashes[index].HashCode {
				return &hashes[index]
			}
			index++
		}
	}
	return nil
}

func (m *Matcher) getMatchingHashFromListNode() *NodeHash {
	if m.node.Modulo == 0 {
		return m.getMatchingHashFromListNodeSearch()
	}
	return m.getMatchingHashFromListNodeTable()
}

func (m *Matcher) advanceHash() bool {
	nextAddIndex := 0
	if m.currentIndex < m.lastIndex {
		nextAddIndex = m.currentIndex + m.node.Length
		if nextAddIndex < len(m.userAgent) {
			m.hash *= rkPrime
			m.hash += uint32(m.userAgent[nextAddIndex])
			m.hash -= m.power * uint32(m.userAgent[m.currentIndex])
			m.currentIndex++
			return true
		}
	}
	return false
}

func (m *Matcher) setInitialHash() bool {
	m.hash = 0
	if m.firstIndex+m.node.Length <= len(m.userAgent) {
		m.power = POWERS[m.node.Length]
		for i := m.firstIndex; i < m.firstIndex+m.node.Length; i++ {
			m.hash *= rkPrime
			m.hash += uint32(m.userAgent[i])
		}
		m.currentIndex = m.firstIndex
		return true
	}
	return false
}

func (m *Matcher) setNextNode(offset int32) {
	if offset > 0 {
		index, ok := m.reader.nodeIndexByOffset[int(offset)]
		if !ok {
			m.node = nil
			return
		}
		m.node = &m.reader.nodes[index]
		m.firstIndex += int(m.node.FirstIndex)
		m.lastIndex += int(m.node.LastIndex)
	} else if offset <= 0 {
		m.deviceIndex = -int(offset)
		m.node = nil
	}
}

func (m *Matcher) applyDrift() {
	if m.firstIndex >= m.drift {
		m.firstIndex = m.firstIndex - m.drift
	} else {
		m.firstIndex = 0
	}
	if m.lastIndex+m.drift < len(m.userAgent) {
		m.lastIndex = m.lastIndex + m.drift
	} else {
		m.lastIndex = len(m.userAgent) - 1
	}
}

func absDiff(a uint32, b uint32) uint32 {
	if a > b {
		return a - b
	}
	return b - a
}

func (m *Matcher) evaluateBinaryNode() {
	nodeHash := m.node.Hashes[0]
	found := false
	if m.setInitialHash() {
		for m.hash != nodeHash.HashCode && m.advanceHash() {
		}
	}
	if m.hash == nodeHash.HashCode {
		found = true
	}
	if !found && m.difference > 0 {
		if m.setInitialHash() {
			for absDiff(m.hash, nodeHash.HashCode) <= uint32(m.difference) && m.advanceHash() {
			}
			if absDiff(m.hash, nodeHash.HashCode) <= uint32(m.difference) {
				found = true
			}
		}
	}
	if !found && m.drift > 0 {
		m.applyDrift()
		if m.setInitialHash() {
			for m.hash != nodeHash.HashCode && m.advanceHash() {
			}
			if m.hash == nodeHash.HashCode {
				found = true
			}
		}
	}
	if !found && m.drift > 0 && m.difference > 0 {
		if m.setInitialHash() {
			for absDiff(m.hash, nodeHash.HashCode) <= uint32(m.difference) && m.advanceHash() {
			}
			if absDiff(m.hash, nodeHash.HashCode) <= uint32(m.difference) {
				found = true
			}
		}
	}
	if found {
		m.setNextNode(nodeHash.NodeOffset)
	} else {
		m.setNextNode(m.node.UnmatchedNodeOffset)
	}
}

func (m *Matcher) getMatchingHashFromListNodeWithinDifference() *NodeHash {
	var nodeHash *NodeHash
	originalHashCode := m.hash
	for m.hash = originalHashCode + uint32(m.difference); m.hash >= originalHashCode-uint32(m.difference) && nodeHash == nil; m.hash-- {
		nodeHash = m.getMatchingHashFromListNode()
	}
	m.hash = originalHashCode
	return nodeHash
}

func (m *Matcher) evaluateListNode() {
	var nodeHash *NodeHash
	if m.setInitialHash() {
		for {
			nodeHash = m.getMatchingHashFromListNode()
			if nodeHash != nil || !m.advanceHash() {
				break
			}
		}

		if nodeHash == nil && m.difference > 0 {
			if m.setInitialHash() {
				for {
					nodeHash = m.getMatchingHashFromListNodeWithinDifference()
					if nodeHash != nil || !m.advanceHash() {
						break
					}
				}
			}
		}
		if nodeHash == nil && m.drift > 0 {
			m.applyDrift()
			if m.setInitialHash() {
				for {
					nodeHash = m.getMatchingHashFromListNode()
					if nodeHash != nil || !m.advanceHash() {
						break
					}
				}
			}
		}

		if nodeHash == nil && m.difference > 0 && m.drift > 0 {
			if m.setInitialHash() {
				for {
					nodeHash = m.getMatchingHashFromListNodeWithinDifference()
					if nodeHash != nil || !m.advanceHash() {
						break
					}
				}
			}
		}
	}
	if nodeHash != nil {
		m.setNextNode(nodeHash.NodeOffset)
	} else {
		m.setNextNode(m.node.UnmatchedNodeOffset)
	}
}

func (m *Matcher) Match() Device {
	for m.node != nil {
		if m.node.HashesCount == 1 {
			m.evaluateBinaryNode()
		} else {
			m.evaluateListNode()
		}
	}
	return Device{
		reader: m.reader,
		offset: m.deviceIndex * (m.reader.devicePropertiesCount + m.reader.componentsCount),
	}
}

func NewMatcher(reader *Reader, userAgent string, drift int, difference int) Matcher {
	node := &reader.nodes[0]
	return Matcher{
		reader:     reader,
		userAgent:  userAgent,
		node:       node,
		drift:      reader.BaseDrift + drift,
		difference: reader.BaseDifference + difference,
		firstIndex: int(node.FirstIndex),
		lastIndex:  int(node.LastIndex),
	}
}
