package deliveryservicedata // TODO rename?

import (
	"errors"
	"github.com/Comcast/traffic_control/traffic_monitor/experimental/traffic_monitor/enum"
)

// New, more structured format:
type StatMeta struct {
	Time int `json:"time"`
}
type StatFloat struct {
	StatMeta
	Value float64 `json:"value"`
}
type StatBool struct {
	StatMeta
	Value bool `json:"value"`
}
type StatInt struct {
	StatMeta
	Value int64 `json:"value"`
}
type StatString struct {
	StatMeta
	Value string `json:"value"`
}

// TODO add `AddCacheStat` func, so it isn't necessary to cast, to add stats
type Stat interface {
	StatType() enum.DSType
	CommonData() *StatCommon
	Copy() Stat
}

type StatCommon struct {
	CachesConfigured StatInt                 `json:"caches_configured"`
	CachesReporting  map[enum.CacheName]bool `json:"caches_reporting"`
	ErrorString      StatString              `json:"error_string"`
	Status           StatString              `json:"status"`
	IsHealthy        StatBool                `json:"is_healthy"`
	IsAvailable      StatBool                `json:"is_available"`
	CachesAvailable  StatInt                 `json:"caches_available"`
}

func (a StatCommon) Copy() StatCommon {
	b := a
	for k, v := range a.CachesReporting {
		b.CachesReporting[k] = v
	}
	return b
}

// StatCacheStats is all the stats generated by a cache.
// This may also be used for aggregate stats, for example, the summary of all cache stats for a cache group, or delivery service.
// Each stat is an array, in case there are multiple data points at different times. However, a single data point i.e. a single array member is common.
type StatCacheStats struct {
	OutBytes    StatInt    `json:"out_bytes"`
	IsAvailable StatBool   `json:"is_available"`
	Status5xx   StatInt    `json:"status_5xx"`
	Status4xx   StatInt    `json:"status_4xx"`
	Status3xx   StatInt    `json:"status_3xx"`
	Status2xx   StatInt    `json:"status_2xx"`
	InBytes     StatFloat  `json:"in_bytes"`
	Kbps        StatFloat  `json:"kbps"`
	Tps5xx      StatInt    `json:"tps_5xx"`
	Tps4xx      StatInt    `json:"tps_4xx"`
	Tps3xx      StatInt    `json:"tps_3xx"`
	Tps2xx      StatInt    `json:"tps_2xx"`
	ErrorString StatString `json:"error_string"`
	TpsTotal    StatInt    `json:"tps_total"`
}

func (a StatCacheStats) Sum(b StatCacheStats) StatCacheStats {
	return StatCacheStats{
		OutBytes:    StatInt{Value: a.OutBytes.Value + b.OutBytes.Value},
		IsAvailable: StatBool{Value: a.IsAvailable.Value || b.IsAvailable.Value},
		Status5xx:   StatInt{Value: a.Status5xx.Value + b.Status5xx.Value},
		Status4xx:   StatInt{Value: a.Status4xx.Value + b.Status4xx.Value},
		Status3xx:   StatInt{Value: a.Status3xx.Value + b.Status3xx.Value},
		Status2xx:   StatInt{Value: a.Status2xx.Value + b.Status2xx.Value},
		InBytes:     StatFloat{Value: a.InBytes.Value + b.InBytes.Value},
		Kbps:        StatFloat{Value: a.Kbps.Value + b.Kbps.Value},
		Tps5xx:      StatInt{Value: a.Tps5xx.Value + b.Tps5xx.Value},
		Tps4xx:      StatInt{Value: a.Tps4xx.Value + b.Tps4xx.Value},
		Tps3xx:      StatInt{Value: a.Tps3xx.Value + b.Tps3xx.Value},
		Tps2xx:      StatInt{Value: a.Tps2xx.Value + b.Tps2xx.Value},
		ErrorString: StatString{Value: a.ErrorString.Value + b.ErrorString.Value},
		TpsTotal:    StatInt{Value: a.TpsTotal.Value + b.TpsTotal.Value},
	}
}

type StatHTTP struct {
	Common      StatCommon
	CacheGroups map[enum.CacheGroupName]StatCacheStats
	Type        map[enum.CacheType]StatCacheStats
	Total       StatCacheStats
}

func (d StatHTTP) StatType() enum.DSType {
	return enum.DSTypeHTTP
}

func (d *StatHTTP) CommonData() *StatCommon {
	return &d.Common
}

func NewStatHTTP() *StatHTTP {
	return &StatHTTP{CacheGroups: map[enum.CacheGroupName]StatCacheStats{}, Type: map[enum.CacheType]StatCacheStats{}, Common: StatCommon{CachesReporting: map[enum.CacheName]bool{}}}
}

func (a StatHTTP) Copy() Stat {
	b := StatHTTP{Common: a.Common.Copy(), Total: a.Total, CacheGroups: map[enum.CacheGroupName]StatCacheStats{}, Type: map[enum.CacheType]StatCacheStats{}}
	for k, v := range a.CacheGroups {
		b.CacheGroups[k] = v
	}
	for k, v := range a.Type {
		b.Type[k] = v
	}
	return &b
}

type StatDNS struct {
	Common StatCommon
}

func (d StatDNS) StatType() enum.DSType {
	return enum.DSTypeDNS
}

func (d *StatDNS) CommonData() *StatCommon {
	return &d.Common
}

func NewStatDNS() *StatDNS {
	return &StatDNS{Common: StatCommon{CachesReporting: map[enum.CacheName]bool{}}}
}

func (s StatDNS) Copy() Stat {
	return &StatDNS{Common: s.Common.Copy()}
}

func (a *StatDNS) Sum(b *StatDNS) {
	a.CommonData().CachesConfigured.Value += b.CommonData().CachesConfigured.Value
	for cache, reporting := range b.CommonData().CachesReporting {
		if reporting {
			a.CommonData().CachesReporting[cache] = true
		}
	}
	a.CommonData().ErrorString.Value += b.CommonData().ErrorString.Value
	//	a.CommonData().Status += b.CommonData().Status // TODO decide what to do about 'summing' the status text
	a.CommonData().IsHealthy.Value = a.CommonData().IsHealthy.Value || b.CommonData().IsHealthy.Value
	a.CommonData().IsAvailable.Value = a.CommonData().IsAvailable.Value || b.CommonData().IsAvailable.Value
	a.CommonData().CachesAvailable.Value += b.CommonData().CachesAvailable.Value
}

var ErrNotProcessedStat = errors.New("This stat is not used.")
