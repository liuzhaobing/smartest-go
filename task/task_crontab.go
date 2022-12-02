package task

import (
	"errors"
	"github.com/robfig/cron/v3"
	"time"
)

type CronMange struct {
	CornServer *cron.Cron
	maxTask    int
	taskMap    map[int]string
}

func (c *CronMange) AddCronTask(config *AddTask, job TaskModel) (int, error) {
	if c.GetCronTaskNum()+1 > c.maxTask {
		return 0, errors.New("max task error")
	}
	if job == nil {
		return 0, errors.New("nil task error")
	}
	var b = &BaseTask{
		T: &job,
	}
	for _, name := range c.taskMap {
		if name == config.TaskName {
			return 0, errors.New("已存在同名定时任务")
		}
	}

	res, err := c.CornServer.AddJob(config.CrontabString, b)
	if err != nil {
		return int(res), err
	}
	c.taskMap[int(res)] = config.TaskName
	return int(res), err
}

func (c *CronMange) RemoveCronTask(id int) {
	c.CornServer.Remove(cron.EntryID(id))
	delete(c.taskMap, id)
}

func (c *CronMange) RemoveCronTaskByName(taskName string) (bool, error) {
	for i, name := range c.taskMap {
		if name == taskName {
			c.RemoveCronTask(i)
			return true, nil
		}
	}
	return false, errors.New("not find this task")
}

func (c *CronMange) GetCronTaskList() (resList []*CronTaskInfo, err error) {
	for i, name := range c.taskMap {
		resList = append(resList, &CronTaskInfo{
			TaskId:      i,
			TaskName:    name,
			NextRunTime: c.CornServer.Entry(cron.EntryID(i)).Next,
		})
	}
	return
}

func (c *CronMange) GetCronTaskNum() int {
	return len(c.CornServer.Entries())
}

type CronTaskInfo struct {
	TaskId      int       `json:"task_id"`
	NextRunTime time.Time `json:"next_run_time"`
	TaskName    string    `json:"task_name"`
}

var CM = &CronMange{}

func init() {
	CM.CornServer = cron.New(cron.WithChain(cron.SkipIfStillRunning(cron.DiscardLogger)))
	CM.taskMap = make(map[int]string)
	CM.maxTask = 10000

	CM.CornServer.Start()
}
