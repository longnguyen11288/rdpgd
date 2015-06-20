package workers

import (
	"fmt"
	"time"

	"github.com/wayneeseguin/rdpg-agent/log"
	"github.com/wayneeseguin/rdpg-agent/rdpg"
	"github.com/wayneeseguin/rdpg-agent/tasks"
)

func Work() {
	for {
		r := rdpg.New()
		err := r.OpenDB("rdpg")
		if err != nil {
			log.Error(fmt.Sprintf(`tasks.Dequeue() Opening rdpg database ! %s`, err))
		}

		// TODO: only work for my role type: write vs read
		task := tasks.Task{}
		sq := `SELECT task_id,func,data,ttl FROM work.tasks WHERE processed_at IS NULL AND locked_by IS NULL ORDER BY created_at DESC LIMIT 1;`
		err = r.DB.Select(&task, sq)
		if err != nil {
			log.Error(fmt.Sprintf(`worker.Run() Selecting Tasks ! %s`, err))
		}

		task.Lock()
		task.Dequeue()

		switch task.Action {
		case "RemoveDatabase":
			// Role: all
			err = RemoveDatabase(task.Data)
		case "BackupDatabase":
			// Role: read
			err = BackupDatabase(task.Data)
		case "BackupAllDatabases":
			// Role: read
			err = BackupAllDatabases(task.Data)
		default:
			err = fmt.Errorf(`worker.Work() Unknown Task Func %s`, task.Action)
		}

		if err != nil {
			log.Error(fmt.Sprintf(`worker.Work() Task %+v ! %s`, task, err))
		} else {
			sq := fmt.Sprintf(`UPDATE work.tasks SET processed_at=CURRENT_TIMESTAMP WHERE task_id='%s';`, task.TaskId)
			_, err = r.DB.Exec(sq)
			if err != nil {
				log.Error(fmt.Sprintf(`tasks.Work() Error setting processed_at for task %s ! %s`, t.TaskId, err))
			}
		}

		task.Unlock()

		time.Sleep(1 * time.Second)
	}
}
