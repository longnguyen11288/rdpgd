package scheduler

import (
	"fmt"
	"time"

	"code.google.com/p/go-uuid/uuid"

	consulapi "github.com/hashicorp/consul/api"
	"github.com/wayneeseguin/rdpg-agent/log"
	"github.com/wayneeseguin/rdpg-agent/rdpg"
)

type Schedule struct {
	ScheduleId string `db:"schedule_id" json:"schedule_id"`
	Schedule   string `db:"schedule" json:"schedule"`
	Action     string `db:"action" json:"action"`
	Data       string `db:"data" json:"data"`
	TTL        string `db:"ttl" json:"ttl"`
}

var (
	schedulerLock *consulapi.Lock
	lockCh        <-chan struct{}
)

/*
Task Scheduler
- Consul Lock & Unlock
- Task TTL: "Task type X should take no more than this long"
- accounting history stored in database.
- TTL based cleanup of task Queue for workers that may have imploded.

Thoughts on scheduler.tasks table:
 - The table stores regularly scheduled tasks which occur at intervals
 - "frequency" is the the amount in time in seconds between when a task should be scheduled
 - After each clock tick a single process looks over this table and grabs any rows
   where NOW() > next_start, when a record is found:
	  - look at the "task" and with a giant case statement, puts entries onto the queue
		- Update the scheduler.tasks table entry "next_start" = "next_start" + duration

*/
func Scheduler() {
	for {
		time.Sleep(10 * time.Second) // Wait 5 seconds before attempting to grab the schedulerLock

		err := Lock()
		if err != nil {
			continue
		}

		r := rdpg.NewRDPG()
		err = r.OpenDB("rdpg")
		if err != nil {
			log.Error(fmt.Sprintf(`scheduler.Scheduler() Opening rdpg database ! %s`, err))
			Unlock()
			continue
		}

		schedules := []Schedule{}

		// Leader, lets get Scheduling!
		// Duration + last_scheduled_at
		sq := fmt.Sprintf(`SELECT schedule_id,schedule,role,action,data,ttl,last_scheduled_at FROM rdpg.schedules WHERE scheduling_at IS NULL;`)
		err = r.DB.Select(&schedules, sq)
		if err != nil {
			log.Error(fmt.Sprintf(`scheduler.Dequeue() Selecting Schedules ! %s`, err))
			Unlock()
			continue
		}
		// TODO: Schedule tasks...
		// task := NewTask()
		//
		Unlock() // Release Consul K/V Lock.
	}
}

func NewSchedule() (s *Schedule) {
	return &Schedule{ScheduleId: uuid.NewUUID().String()}
}

func Lock() (err error) {
	// Acquire consul schedulerLock to aquire right to schedule tasks.
	key := "rdpg/manager/scheduler/leader"
	client, _ := consulapi.NewClient(consulapi.DefaultConfig())
	schedulerLock, err := client.LockKey(key)
	if err != nil {
		log.Error(fmt.Sprintf("scheduler.Schedule() Error Locking Scheduler Key %s ! %s", key, err))
		return
	}

	lockCh, err := schedulerLock.Lock(nil) // Acquire Consul K/V Lock
	if err != nil {
		log.Error(fmt.Sprintf("scheduler.Lock() Error Aquiring Scheduler Key lock %s ! %s", key, err))
		return
	}

	if lockCh == nil {
		err = fmt.Errorf(`Scheduler Lock not aquired.`)
	}

	return
}

func Unlock() (err error) {
	if schedulerLock != nil {
		err = schedulerLock.Unlock()
	}
	return
}
