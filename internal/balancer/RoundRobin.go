package balancer

import (
	"sync"
	"sync/atomic"

	"github.com/google/uuid"
)

// логика по добавлению и удалению активных соединений
func (back *Backend) IncrementConn() {
	atomic.AddInt64(&back.ActiveCon, 1)
}

func (back *Backend) DecrementConn() {
	atomic.AddInt64(&back.ActiveCon, -1)
}

type RoundRobin struct {
	backendsMap map[string]*Backend // мапа для быстрого доступа по url
	backends    []*Backend          // слайс необходим для сохранения логики Round Robin
	index       uint32              //индекс бэкенда
	mu          sync.RWMutex        //для того чтобы синхронизировать данные
}

// инициализируем слайс бэкендов
func NewRR(backends []*Backend) *RoundRobin {
	backendsMap := make(map[string]*Backend)
	for _, b := range backends {
		backendsMap[b.Url] = b
	}
	return &RoundRobin{
		backends:    backends,
		backendsMap: backendsMap,
	}
}

// функция нужна для того чтобы проставить стейт бэкенда
func (back *Backend) SetAlive(state bool) {
	back.mu.Lock()
	defer back.mu.Unlock()
	back.State = state //будет проставляться false как базовое значение для bool
}

// если бэкенд недоступен то помечаем его стейт как false
func (r *RoundRobin) MarkBackendDown(url string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.backendsMap[url].SetAlive(false)
}

func (r *RoundRobin) MarkBackendUp(url string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.backendsMap[url].SetAlive(true)
}

func (r *RoundRobin) AddBackend(backend *Backend) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.backendsMap[backend.Url]; !exists {
		r.backendsMap[backend.Url] = backend
		r.backends = append(r.backends, backend)
		backend.SetAlive(false)
		uuidOrig := uuid.New()
		r.backendsMap[backend.Url].ID = uuidOrig //хорошо было бы добавить здесь проверку на повторяющиеся uuid`ы или реализовать хранилку uuid`ов откуда я мог бы забирать точно актуальные значения
	}
}

// реализуем саму логику балансировки роунд робин которая из себя представляет алгоритм
// который балансирует нагрузку по кругу, т.е. это будет похоже на очередь, которая замкнута
func (r *RoundRobin) NextBackendRR() *Backend {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.backends) == 0 {
		return nil
	}

	next := atomic.AddUint32(&r.index, 1) % uint32(len(r.backends)) //используем % для того чтобы не выходить за границы массива
	for i := uint32(0); i < uint32(len(r.backends)); i++ {
		nowIndex := (next + i) % uint32(len(r.backends))

		backend := r.backends[nowIndex]

		backend.mu.RLock()
		state := backend.State
		backend.mu.RUnlock()

		if state {
			backend.IncrementConn()
			return backend
		}
	}
	return nil
}

func (r *RoundRobin) RemoveBackend(url string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.backendsMap, url)

	for i, b := range r.backends {
		if b.Url == url {
			r.backends = append(r.backends[:i], r.backends[i+1:]...)
			break
		}
	}
}

func (r *RoundRobin) AllBackend() []*Backend {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.backends
}
