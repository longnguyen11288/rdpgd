package tasks

import (
	"fmt"
	"os"
	"syscall"
	"time"

	consulapi "github.com/hashicorp/consul/api"
	"github.com/wayneeseguin/rdpgd/log"
	"github.com/wayneeseguin/rdpgd/rdpg"
)

var (
	lock   *consulapi.Lock
	lockCh <-chan struct{}
)

type Schedule struct {
	Id       string `db:"id" json:"id"`
	Schedule string `db:"schedule" json:"schedule"`
	Role     string `db:"role" json:"role"`
	Action   string `db:"action" json:"action"`
	Data     string `db:"data" json:"data"`
	TTL      string `db:"ttl" json:"ttl"`
}

/*
Task Scheduler
- Task TTL: "Task type X should take no more than this long"
- accounting history stored in database.
- TTL based cleanup of task Queue for workers that may have imploded.

Thoughts on scheduler.tasks table:
 - The table stores regularly scheduled tasks which occur at intervals
 - "frequency" is the the amount in time in seconds between when a task should be scheduled
 - After each clock tick a single process looks over this table and grabs any rows
 where NOW() >= (frequency::interval + last_scheduled_at), when a record is found:
   - look at the "task" and with a giant case statement, puts entries onto the queue
   - Update the scheduler.tasks table entry "next_start" = "next_start" + duration

Interval + start_at clock time of day

*/
func Scheduler() {
	r := rdpg.NewRDPG()
	err := r.OpenDB("rdpg")
	if err != nil {
		log.Error(fmt.Sprintf(`tasks.Scheduler() Opening rdpg database ! %s`, err))
		SchedulerUnlock()

		log.Error(fmt.Sprintf("tasks.Scheduler() Failed connecting to %s err: %s", r.URI, err))
		proc, _ := os.FindProcess(os.Getpid())
		proc.Signal(syscall.SIGTERM)
	}
	defer r.DB.Close()

	for {
		err := SchedulerLock()
		if err != nil {
			continue
		}

		schedules := []Schedule{}

		sq := fmt.Sprintf(`SELECT schedule, role, action, data, ttl, last_scheduled_at FROM tasks.schedules WHERE CURRENT_TIMESTAMP >= (last_scheduled_at + frequency::interval);`)
		err = r.DB.Select(&schedules, sq)
		if err != nil {
			log.Error(fmt.Sprintf(`tasks.Scheduler() Selecting Schedules ! %s`, err))
			SchedulerUnlock()
			continue
		}
		for _, schedule := range schedules {
			sq = fmt.Sprintf(`UPDATE tasks.schedules SET last_scheduled_at = CURRENT_TIMESTAMP WHERE id=%s;`, schedule.Id)
			_, err = r.DB.Exec(sq)
			if err != nil {
				log.Error(fmt.Sprintf(`tasks.Scheduler() Selecting Schedules ! %s`, err))
			}

			task := NewTask()
			task.Role = schedule.Role
			task.Action = schedule.Action
			task.Data = schedule.Data
			task.TTL = schedule.TTL
			task.Enqueue()
		}

		SchedulerUnlock() // Release Consul K/V Lock.

		time.Sleep(10 * time.Second) // Wait before attempting to grab the lock
	}
}

func NewSchedule() (s *Schedule) {
	return &Schedule{}
}

func SchedulerLock() (err error) {
	// Acquire consul schedulerLock to aquire right to schedule tasks.
	r := rdpg.NewRDPG()
	key := fmt.Sprintf("rdpg/%s/tasks/scheduler", r.ClusterID)
	client, _ := consulapi.NewClient(consulapi.DefaultConfig())
	lock, err = client.LockKey(key)
	if err != nil {
		log.Error(fmt.Sprintf("tasks.SchedulerLock() Error Locking Scheduler Key %s ! %s", key, err))
		return
	}

	lockCh, err = lock.Lock(nil) // Acquire Consul K/V Lock
	if err != nil {
		log.Error(fmt.Sprintf("tasks.SchedulerLock() Error Aquiring Scheduler Key lock %s ! %s", key, err))
		return
	}

	if lockCh == nil {
		err = fmt.Errorf(`tasks.SchedulerLock() Scheduler Lock not aquired.`)
	}

	return
}

func SchedulerUnlock() (err error) {
	if lock != nil {
		err = lock.Unlock()
		if err != nil {
			log.Error(fmt.Sprintf("tasks.SchedulerUnlock() Error Unlocking Scheduler ! %s", err))
		}
	}
	return
}
