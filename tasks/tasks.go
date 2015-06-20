package tasks

import (
	"fmt"

	"code.google.com/p/go-uuid/uuid"

	consulapi "github.com/hashicorp/consul/api"
	"github.com/wayneeseguin/rdpg-agent/log"
	"github.com/wayneeseguin/rdpg-agent/rdpg"
)

type Task struct {
	TaskId string `db:"task_id" json:"task_id"`
	Role   string `db:"data" json:"data"`
	Action string `db:"action" json:"action"`
	Data   string `db:"data" json:"data"`
	TTL    string `db:"ttl" json:"ttl"`
	lock   *consulapi.Lock
	lockCh <-chan struct{}
}

func NewTask() *Task {
	return &Task{TaskId: uuid.NewUUID().String()}
}

// Insert start/stop/(status stuff) into history.backups:
//   kind {backup,restore,s3upload,...},
//   action {start,stop}
//   file location/status,
//   s3 bucket location
// Insert start/stop/(status stuff) into history.restores
// host role/type that Task applies to eg. write/read

// Attempt to obtain a consul Lock, return err if lock if fail to obtain lock.
func (t *Task) Lock() (err error) {
	client, _ := consulapi.NewClient(consulapi.DefaultConfig())
	t.lock, err = client.LockKey(fmt.Sprintf("rdpg/work/tasks/%s/lock", t.TaskId))
	if err != nil {
		log.Error(fmt.Sprintf("tasks.Lock() Error aquiring lock for task %s ! %s", t.TaskId, err))
		return
	}
	t.lockCh, err = t.lock.Lock(nil)
	if err != nil {
		log.Error(fmt.Sprintf("tasks.Lock() Error aquiring lock for task %s ! %s", t.TaskId, err))
	}
	log.Trace(fmt.Sprintf(`tasks.Dequeue() Aquired leader lock for task %s`, t.TaskId))
	return
}

func (t *Task) Unlock() (err error) {
	if t.lock != nil {
		err = t.lock.Unlock()
	}
	return
}

func (t *Task) Enqueue() (err error) {
	// Save Task to database queue.
	r := rdpg.NewRDPG()
	err = r.OpenDB("rdpg")
	if err != nil {
		log.Error(fmt.Sprintf(`tasks.Enqueue() Opening rdpg database ! %s`, err))
	}
	sq := fmt.Sprintf(`INSERT INTO work.tasks(task_id,action,data,ttl) VALUES ('%s','%s','%s','%s');`, t.TaskId, t.Action, t.Data, t.TTL)
	_, err = r.DB.Exec(sq)
	if err != nil {
		log.Error(fmt.Sprintf(`tasks.Enqueue() Insert Task %+v ! %s`, t, err))
	}
	log.Trace(fmt.Sprintf(`tasks.Enqueue() Task Enqueued > %+v`, t))
	return
}

func (t *Task) Dequeue() (err error) {
	if t.lock == nil {
		log.Trace(fmt.Sprintf(`tasks.Dequeue() Not Leader, denied dequeue for task %s`, t.TaskId))
		return
	}

	r := rdpg.NewRDPG()
	err = r.OpenDB("rdpg")
	if err != nil {
		t.Unlock()
		log.Error(fmt.Sprintf(`tasks.Dequeue() Opening rdpg database ! %s`, err))
		return
	}

	sq := fmt.Sprintf(`SELECT task_id,action,data,ttl FROM work.tasks WHERE task_id = '%s' LIMIT 1;`, t.TaskId)
	err = r.DB.Select(&t, sq)
	if err != nil {
		t.Unlock()
		log.Error(fmt.Sprintf(`tasks.Dequeue() Selecting Task %s ! %s`, t.TaskId, err))
		return
	}

	// TODO: locked_by...
	sq = fmt.Sprintf(`UPDATE work.tasks SET processing_at=CURRENT_TIMESTAMP WHERE task_id = '%s';`, t.TaskId)
	err = r.DB.Select(&t, sq)
	if err != nil {
		t.Unlock()
		log.Error(fmt.Sprintf(`tasks.Dequeue() Updating Task %s processing_at ! %s`, t.TaskId, err))
		return
	}
	log.Trace(fmt.Sprintf(`tasks.Dequeue() Task Dequeued > %+v`, t))
	return
}
