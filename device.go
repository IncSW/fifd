package fiftyonedegrees

type Device struct {
	reader *Reader
	offset int
}

func (d *Device) GetValue(propertyName string) string {
	for i, name := range d.reader.requiredPropertiesNames {
		if name == propertyName {
			property := d.reader.properties[d.reader.requiredProperties[i]]
			profile := d.reader.profiles[d.reader.devices[(d.offset+property.ComponentIndex)*4]+property.SubIndex*4]
			return d.reader.strings[profile]
		}
	}
	return ""
}
