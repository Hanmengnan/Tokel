package gotools

import (
	"errors"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

const (
	RUNNING uint64 = iota
	STOP
)

type Task struct {
	Params  []interface{}
	Handler func(params ...interface{})
}

type GoroutinesPool struct {
	capacity  uint64 // 容量
	workerNum uint64 // 目前的协程数
	status    uint64 // 状态

	taskChan chan *Task //任务池

	sync.Mutex // 锁
}

func NewPool(capacity uint64) (*GoroutinesPool, error) {
	if capacity <= 0 {
		return nil, errors.New("invalid pool capacity")
	} else {
		pool := &GoroutinesPool{
			capacity:  capacity,
			workerNum: 0,
			status:    RUNNING,
			taskChan:  make(chan *Task, capacity),
		}
		return pool, nil
	}
}

func (p *GoroutinesPool) getPoolCapacity() uint64 {
	return p.capacity
}

func (p *GoroutinesPool) getPoolWorkerNum() uint64 {
	return atomic.LoadUint64(&p.workerNum)
}

func (p *GoroutinesPool) addNewWorker() {
	atomic.AddUint64(&p.workerNum, 1)
}

func (p *GoroutinesPool) decWorker() {
	atomic.AddUint64(&p.workerNum, ^uint64(0))
}

func (p *GoroutinesPool) getPoolStatus() uint64 {
	return atomic.LoadUint64(&p.status)
}

func (p *GoroutinesPool) setPoolStatus(status uint64) bool {
	defer p.Unlock()
	p.Lock()
	if p.status == status {
		return false
	} else {
		p.status = status
		return true
	}
}

func (p *GoroutinesPool) newTaskGoroutine() {
	p.addNewWorker()

	go func() {
		defer func() {
			p.decWorker()
			if r := recover(); r != nil {
				log.Printf("worker %s has panic\n", r)
			}
		}()

		for {
			select {
			case task, ok := <-p.taskChan:
				if !ok {
					return
				}
				task.Handler(task.Params...)
			}
		}
	}()
}

func (p *GoroutinesPool) NewTask(t *Task) error {
	defer p.Unlock()
	p.Lock()

	if p.getPoolStatus() == RUNNING {
		if p.getPoolCapacity() > p.getPoolWorkerNum() {
			p.newTaskGoroutine()
		}
		p.taskChan <- t
		return nil
	} else {
		return errors.New("goroutines pool has already closed")
	}
}

func (p *GoroutinesPool) ClosePool() {
	p.setPoolStatus(STOP)
	for len(p.taskChan) > 0 {
		time.Sleep(time.Second * 60)
	}
	close(p.taskChan)
}
