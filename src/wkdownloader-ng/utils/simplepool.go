package utils

import "sync"

type SimplePool struct {
	queue chan int
	wg    *sync.WaitGroup
}

func NewPool(size int) *SimplePool {
	if size <= 0 {
		size = 1
	}
	return &SimplePool{
		queue: make(chan int, size),
		wg:    &sync.WaitGroup{},
	}
}

func (p *SimplePool) Add(delta int) {
	for i := 0; i < delta; i++ {
		p.queue <- 1
	}
	for i := 0; i > delta; i-- {
		<-p.queue
	}
	p.wg.Add(delta)
}

func (p *SimplePool) Done() {
	<-p.queue
	p.wg.Done()
}

func (p *SimplePool) Wait() {
	p.wg.Wait()
}
