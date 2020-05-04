package fifd

type Matcher struct {
	reader       *Reader
	userAgent    string
	currentNode  *Node
	hash         uint32
	power        uint32
	firstIndex   int
	lastIndex    int
	currentIndex int
	deviceIndex  int
}

func (m *Matcher) getMatchingHashFromListNodeSearch() *NodeHash {
	hashes := m.currentNode.Hashes
	upper := m.currentNode.HashesCount - 1
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
	hashes := m.currentNode.Hashes
	index := int(m.hash % m.currentNode.Modulo)
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
	if m.currentNode.Modulo == 0 {
		return m.getMatchingHashFromListNodeSearch()
	}
	return m.getMatchingHashFromListNodeTable()
}

func (m *Matcher) advanceHash() bool {
	nextAddIndex := 0
	if m.currentIndex < m.lastIndex {
		nextAddIndex = m.currentIndex + m.currentNode.Length
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
	if m.firstIndex+m.currentNode.Length <= len(m.userAgent) {
		m.power = POWERS[m.currentNode.Length]
		for i := m.firstIndex; i < m.firstIndex+m.currentNode.Length; i++ {
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
			m.currentNode = nil
			return
		}
		m.currentNode = &m.reader.nodes[index]
		m.firstIndex += int(m.currentNode.FirstIndex)
		m.lastIndex += int(m.currentNode.LastIndex)
	} else if offset <= 0 {
		m.deviceIndex = -int(offset)
		m.currentNode = nil
	}
}

func (m *Matcher) evaluateBinaryNode() {
	nodeHash := m.currentNode.Hashes[0]
	found := false
	if m.setInitialHash() {
		for m.hash != nodeHash.HashCode && m.advanceHash() {
		}
	}
	if m.hash == nodeHash.HashCode {
		found = true
	}
	if found {
		m.setNextNode(nodeHash.NodeOffset)
	} else {
		m.setNextNode(m.currentNode.UnmatchedNodeOffset)
	}
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
	}
	if nodeHash != nil {
		m.setNextNode(nodeHash.NodeOffset)
	} else {
		m.setNextNode(m.currentNode.UnmatchedNodeOffset)
	}
}

func (m *Matcher) Match() Device {
	for m.currentNode != nil {
		if m.currentNode.HashesCount == 1 {
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

func NewMatcher(reader *Reader, userAgent string) Matcher {
	return Matcher{
		reader:      reader,
		userAgent:   userAgent,
		currentNode: &reader.nodes[0],
	}
}
