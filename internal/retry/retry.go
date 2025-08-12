package retry

import (
	"log"
	"time"
)

func sleepTime(i int64) int64 {
	if i == 0 {
		i += 1
		return i
	}
	i += 2
	return i

}

// Do пытается выполнить fn с повтором при ошибке
func Do(attempts int, fn func() error, isRetriable func(error) bool) error {
	var err error
	var delays int64
	for i := 0; i < attempts; i++ {
		err = fn()
		if err == nil {
			return nil // Успех!
		}
		log.Printf("Ошибка: %v — retriable? %v", err, isRetriable(err))

		if !isRetriable(err) {
			return err
		}

		delays = sleepTime(delays)

		if i < attempts {
			log.Printf("Попытка %d неудачна: %v, спим %v", i+1, err, delays)
			time.Sleep(time.Duration(delays) * time.Second)
		}
	}

	// Все попытки исчерпаны
	return err
}
