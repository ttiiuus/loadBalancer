package balancer

import (
	"sync"

	"github.com/google/uuid"
)

// структура одного бэкенда для которого будет реализовываться балансировка
type Backend struct {
	Url       string
	ActiveCon int64
	mu        sync.RWMutex
	State     bool
	ID        uuid.UUID
}

type Balancer interface {
	AllBackend() []*Backend
	RemoveBackend(url string)
	NextBackendRR() *Backend
	AddBackend(backend *Backend)
	MarkBackendUp(url string)
	MarkBackendDown(url string)
}
