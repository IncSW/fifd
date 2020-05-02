package test

import (
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"

	go51degrees "github.com/IncSW/go-51degrees"
	cgo51degrees "github.com/IncSW/go-51degrees/test/51degrees"
)

func TestReader(t *testing.T) {
	reader, err := go51degrees.NewReaderFromFile("./51Degrees-EnterpriseV3.4.trie")
	if err != nil {
		t.Fatal(err)
	}
	provider := cgo51degrees.NewProvider("./51Degrees-EnterpriseV3.4.trie")

	for i, userAgent := range userAgents {
		device := reader.MatchDevice(userAgent)
		match := provider.GetMatch(userAgent)
		if !assert.Equal(t, device.GetValue("HardwareVendor"), match.GetValue("HardwareVendor"), userAgent) ||
			!assert.Equal(t, device.GetValue("HardwareModel"), match.GetValue("HardwareModel"), userAgent) ||
			!assert.Equal(t, device.GetValue("PlatformName"), match.GetValue("PlatformName"), i, userAgent) ||
			!assert.Equal(t, device.GetValue("PlatformVersion"), match.GetValue("PlatformVersion"), userAgent) ||
			!assert.Equal(t, device.GetValue("BrowserName"), match.GetValue("BrowserName"), userAgent) ||
			!assert.Equal(t, device.GetValue("BrowserVersion"), match.GetValue("BrowserVersion"), userAgent) ||
			!assert.Equal(t, device.GetValue("IsCrawler"), match.GetValue("IsCrawler"), userAgent) ||
			!assert.Equal(t, device.GetValue("DeviceType"), match.GetValue("DeviceType"), userAgent) {
			cgo51degrees.DeleteMatch(match)
			return
		}
		cgo51degrees.DeleteMatch(match)
	}
}

func BenchmarkReader(b *testing.B) {
	reader, err := go51degrees.NewReaderFromFile("./51Degrees-EnterpriseV3.4.trie")
	if err != nil {
		b.Fatal(err)
	}
	provider := cgo51degrees.NewProvider("./51Degrees-EnterpriseV3.4.trie")

	b.Log("fixed ua:", userAgents[0])
	b.ReportAllocs()

	b.Run("go", func(b *testing.B) {
		b.Run("fixed", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				device := reader.MatchDevice(userAgents[0])
				device.GetValue("HardwareVendor")
				device.GetValue("HardwareModel")
				device.GetValue("PlatformName")
				device.GetValue("PlatformVersion")
				device.GetValue("BrowserName")
				device.GetValue("BrowserVersion")
				device.GetValue("IsCrawler")
				device.GetValue("DeviceType")
			}
		})
		b.Run("fixed-parallel", func(b *testing.B) {
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					device := reader.MatchDevice(userAgents[0])
					device.GetValue("HardwareVendor")
					device.GetValue("HardwareModel")
					device.GetValue("PlatformName")
					device.GetValue("PlatformVersion")
					device.GetValue("BrowserName")
					device.GetValue("BrowserVersion")
					device.GetValue("IsCrawler")
					device.GetValue("DeviceType")
				}
			})
		})
		b.Run("range", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				device := reader.MatchDevice(userAgents[i%len(userAgents)])
				device.GetValue("HardwareVendor")
				device.GetValue("HardwareModel")
				device.GetValue("PlatformName")
				device.GetValue("PlatformVersion")
				device.GetValue("BrowserName")
				device.GetValue("BrowserVersion")
				device.GetValue("IsCrawler")
				device.GetValue("DeviceType")
			}
		})
		b.Run("range-parallel", func(b *testing.B) {
			i := int64(0)
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					device := reader.MatchDevice(userAgents[int(atomic.AddInt64(&i, 1))%len(userAgents)])
					device.GetValue("HardwareVendor")
					device.GetValue("HardwareModel")
					device.GetValue("PlatformName")
					device.GetValue("PlatformVersion")
					device.GetValue("BrowserName")
					device.GetValue("BrowserVersion")
					device.GetValue("IsCrawler")
					device.GetValue("DeviceType")
				}
			})
		})
	})

	b.Run("cgo", func(b *testing.B) {
		b.Run("fixed", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				match := provider.GetMatch(userAgents[0])
				match.GetValue("HardwareVendor")
				match.GetValue("HardwareModel")
				match.GetValue("PlatformName")
				match.GetValue("PlatformVersion")
				match.GetValue("BrowserName")
				match.GetValue("BrowserVersion")
				match.GetValue("IsCrawler")
				match.GetValue("DeviceType")
				cgo51degrees.DeleteMatch(match)
			}
		})
		b.Run("fixed-parallel", func(b *testing.B) {
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					match := provider.GetMatch(userAgents[0])
					match.GetValue("HardwareVendor")
					match.GetValue("HardwareModel")
					match.GetValue("PlatformName")
					match.GetValue("PlatformVersion")
					match.GetValue("BrowserName")
					match.GetValue("BrowserVersion")
					match.GetValue("IsCrawler")
					match.GetValue("DeviceType")
					cgo51degrees.DeleteMatch(match)
				}
			})
		})
		b.Run("range", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				match := provider.GetMatch(userAgents[i%len(userAgents)])
				match.GetValue("HardwareVendor")
				match.GetValue("HardwareModel")
				match.GetValue("PlatformName")
				match.GetValue("PlatformVersion")
				match.GetValue("BrowserName")
				match.GetValue("BrowserVersion")
				match.GetValue("IsCrawler")
				match.GetValue("DeviceType")
				cgo51degrees.DeleteMatch(match)
			}
		})
		b.Run("range-parallel", func(b *testing.B) {
			i := int64(0)
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					match := provider.GetMatch(userAgents[int(atomic.AddInt64(&i, 1))%len(userAgents)])
					match.GetValue("HardwareVendor")
					match.GetValue("HardwareModel")
					match.GetValue("PlatformName")
					match.GetValue("PlatformVersion")
					match.GetValue("BrowserName")
					match.GetValue("BrowserVersion")
					match.GetValue("IsCrawler")
					match.GetValue("DeviceType")
					cgo51degrees.DeleteMatch(match)
				}
			})
		})
	})
}
