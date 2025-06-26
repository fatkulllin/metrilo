package storage

type Repositories interface {
	SaveGauge(name string, value float64)
	SaveCounter(name string, value int64)
	GetCounter(name string)
	GetGauge(name string)
}
