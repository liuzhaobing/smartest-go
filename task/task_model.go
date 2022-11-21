package task

import (
	"fmt"
	"time"
)

type TaskModel interface {
	pre()
	run()
	end()
	stop()
}

type backendTask struct {
	BaseTask
}

func (b *backendTask) pre() {
	b.starTime = time.Now()
}

func (b *backendTask) run() {
}

func (b *backendTask) end() {
	b.endTime = time.Now()
}

func (b *backendTask) stop() {
	b.endTime = time.Now()
}

var _ TaskModel = &backendTask{}

type BaseTask struct {
	T        *TaskModel
	starTime time.Time
	endTime  time.Time
}

func (b *BaseTask) Run() {
	defer func() {
		err := recover()
		if err != nil {
			fmt.Println(err)
			return
		}
	}()
	(*b.T).pre()
	(*b.T).run()
	(*b.T).end()
}
