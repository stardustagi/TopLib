package uuid

import "fmt"

var uuidWorker *UuidWorker

func init() {
	uuidWorker, _ = NewUuidWorker(0)
}

func GetUuidString() string {
	return fmt.Sprintf("%d", uuidWorker.Get())
}
