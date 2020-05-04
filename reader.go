package fifd

import (
	"encoding/binary"
	"errors"
	"io/ioutil"
	"sort"
)

type Reader struct {
	version                 uint16
	strings                 map[int]string
	componentsCount         int
	devicePropertiesCount   int
	properties              []Property
	requiredProperties      []int
	requiredPropertiesNames []string
	profiles                map[int]int
	devices                 map[int]int
	nodes                   []Node
	nodeIndexByOffset       map[int]int
}

func (r *Reader) MatchDevice(userAgent string) Device {
	matcher := NewMatcher(r, userAgent)
	return matcher.Match()
}

func NewReaderFromFile(filename string) (*Reader, error) {
	buffer, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return NewReader(buffer)
}

func NewReader(buffer []byte) (*Reader, error) {
	reader := &Reader{}
	err := initReader(reader, buffer)
	if err != nil {
		return nil, err
	}
	return reader, nil
}

func initReader(reader *Reader, buffer []byte) error {
	offset := 0
	reader.version = binary.LittleEndian.Uint16(buffer[offset : offset+2])
	if reader.version != 34 {
		return errors.New("invalid version")
	}
	offset += 2
	// formatOffset := int32(binary.LittleEndian.Uint32(buffer[offset : offset+4]))
	offset += 4
	// nameOffset := int32(binary.LittleEndian.Uint32(buffer[offset : offset+4]))
	offset += 4
	// tag := buffer[offset : offset+16]
	offset += 16

	// publishedYear := int16(binary.LittleEndian.Uint16(buffer[offset : offset+2]))
	offset += 2
	// publishedMonth := buffer[offset]
	offset += 1
	// publishedDay := buffer[offset]
	offset += 1

	// nextUpdateYear := int16(binary.LittleEndian.Uint16(buffer[offset : offset+2]))
	offset += 2
	// nextUpdateMonth := buffer[offset]
	offset += 1
	// nextUpdateDay := buffer[offset]
	offset += 1

	// copyrightOffset := int32(binary.LittleEndian.Uint32(buffer[offset : offset+4]))
	offset += 4
	// maxStringLength := binary.LittleEndian.Uint16(buffer[offset : offset+2])
	offset += 2

	stringsDataSize := binary.LittleEndian.Uint32(buffer[offset : offset+4])
	offset += 4
	stringsStartOffset := offset
	offset += int(stringsDataSize)

	reader.strings = map[int]string{}
	stringsOffset := stringsStartOffset
	for stringsOffset < offset {
		key := stringsOffset - stringsStartOffset
		length := int(binary.LittleEndian.Uint16(buffer[stringsOffset : stringsOffset+2]))
		stringsOffset += 2
		reader.strings[key] = string(buffer[stringsOffset : stringsOffset+length-1])
		stringsOffset += length
	}

	reader.componentsCount = int(binary.LittleEndian.Uint32(buffer[offset : offset+4]))
	offset += 4 + 4*reader.componentsCount

	httpHeadersCount := int(binary.LittleEndian.Uint32(buffer[offset : offset+4]))
	offset += 4 + 4*httpHeadersCount

	reader.devicePropertiesCount = int(binary.LittleEndian.Uint32(buffer[offset : offset+4]))
	offset += 4

	propertiesCount := int(binary.LittleEndian.Uint32(buffer[offset : offset+4]))
	offset += 4
	reader.properties = make([]Property, propertiesCount)
	for i := 0; i < propertiesCount; i++ {
		reader.properties[i].StringOffset = int(binary.LittleEndian.Uint32(buffer[offset : offset+4]))
		offset += 4
		reader.properties[i].ComponentIndex = int(binary.LittleEndian.Uint32(buffer[offset : offset+4]))
		offset += 4
		reader.properties[i].SubIndex = int(binary.LittleEndian.Uint32(buffer[offset : offset+4]))
		offset += 4
		// reader.properties[i].HeaderCount = int(binary.LittleEndian.Uint32(buffer[offset : offset+4]))
		offset += 4
		// reader.properties[i].HeaderFirstIndex = int(binary.LittleEndian.Uint32(buffer[offset : offset+4]))
		offset += 4
	}

	profilesDataSize := int(binary.LittleEndian.Uint32(buffer[offset : offset+4]))
	offset += 4
	profilesStartOffset := offset
	profilesCount := profilesDataSize / 4
	reader.profiles = make(map[int]int, profilesCount)
	for i := 0; i < profilesCount; i++ {
		reader.profiles[offset-profilesStartOffset] = int(binary.LittleEndian.Uint32(buffer[offset : offset+4]))
		offset += 4
	}

	devicesDataSize := int(binary.LittleEndian.Uint32(buffer[offset : offset+4]))
	offset += 4
	devicesStartOffset := offset
	devicesCount := devicesDataSize / 4
	reader.devices = make(map[int]int, devicesCount)
	for i := 0; i < devicesCount; i++ {
		reader.devices[offset-devicesStartOffset] = int(binary.LittleEndian.Uint32(buffer[offset : offset+4]))
		offset += 4
	}

	nodeDataSize := binary.LittleEndian.Uint32(buffer[offset : offset+4])
	offset += 4
	nodeOffset := offset
	offset += int(nodeDataSize)
	nodesStartOffset := nodeOffset
	nodeIndex := 0
	reader.nodeIndexByOffset = map[int]int{}
	for nodeOffset < offset {
		key := nodeOffset - nodesStartOffset
		node := Node{}
		node.UnmatchedNodeOffset = int32(binary.LittleEndian.Uint32(buffer[nodeOffset : nodeOffset+4]))
		nodeOffset += 4
		node.FirstIndex = int16(binary.LittleEndian.Uint16(buffer[nodeOffset : nodeOffset+2]))
		nodeOffset += 2
		node.LastIndex = int16(binary.LittleEndian.Uint16(buffer[nodeOffset : nodeOffset+2]))
		nodeOffset += 2
		node.Length = int(buffer[nodeOffset])
		nodeOffset += 1
		node.HashesCount = int32(binary.LittleEndian.Uint32(buffer[nodeOffset : nodeOffset+4]))
		nodeOffset += 4
		node.Modulo = binary.LittleEndian.Uint32(buffer[nodeOffset : nodeOffset+4])
		nodeOffset += 4
		node.Hashes = make([]NodeHash, node.HashesCount)
		for i := 0; i < int(node.HashesCount); i++ {
			node.Hashes[i].HashCode = binary.LittleEndian.Uint32(buffer[nodeOffset : nodeOffset+4])
			nodeOffset += 4
			node.Hashes[i].NodeOffset = int32(binary.LittleEndian.Uint32(buffer[nodeOffset : nodeOffset+4]))
			nodeOffset += 4
		}
		reader.nodes = append(reader.nodes, node)
		reader.nodeIndexByOffset[key] = nodeIndex
		nodeIndex++
	}

	reader.requiredProperties = make([]int, len(reader.properties))
	for i := 0; i < len(reader.properties); i++ {
		reader.requiredProperties[i] = i
	}
	reader.requiredPropertiesNames = make([]string, len(reader.requiredProperties))
	for i := 0; i < len(reader.requiredProperties); i++ {
		reader.requiredPropertiesNames[i] = reader.strings[reader.properties[i].StringOffset]
	}
	sort.Slice(reader.requiredPropertiesNames, func(i, j int) bool {
		return reader.requiredPropertiesNames[i] < reader.requiredPropertiesNames[j]
	})
	for i := 0; i < len(reader.requiredProperties); i++ {
		reader.requiredProperties[i] = -1
		name := reader.requiredPropertiesNames[i]
		for j := 0; j < len(reader.properties); j++ {
			if name == reader.strings[reader.properties[j].StringOffset] {
				reader.requiredProperties[i] = j
			}
		}
	}

	return nil
}
