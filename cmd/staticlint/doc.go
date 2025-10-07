// Package staticlint implements a multichecker tool for static analysis.
//
// Запуск:
//
//	go run ./cmd/staticlint ./...
//
// Анализаторы:
//   - printf: проверка вызовов Printf на корректность.
//   - shadow: выявление затенённых переменных.
//   - structtag: проверка корректности struct tag.
//   - SAxxx: весь класс ошибок из staticcheck.
//   - ST1000: пример из других классов staticcheck.
//   - ineffassign: проверка неиспользованных присваиваний.
//   - noexit: собственный анализатор, запрещающий os.Exit в main.
package main
