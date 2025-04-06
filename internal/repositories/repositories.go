package repositories

// Repository - интерфейс для работы с данными
type Repositories interface {
	SetGauge(name string, value float64)
	AddCounter(name string, value int64)
}
