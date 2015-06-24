package tasks

import (
	"fmt"
	"os"
	"syscall"
	"time"

	"github.com/wayneeseguin/rdpgd/log"
	"github.com/wayneeseguin/rdpgd/rdpg"
)

func Work(role string) {
	r := rdpg.NewRDPG()
	err := r.OpenDB("rdpg")
	if err != nil {
		log.Error(fmt.Sprintf("tasks.Work() Failed connecting to %s err: %s", r.URI, err))
		proc, _ := os.FindProcess(os.Getpid())
		proc.Signal(syscall.SIGTERM)
	}
	defer r.DB.Close()

	for {
		// TODO: only work for my role type: write vs read
		// eg. WHERE role = 'read'
		tasks := []Task{}
		sq := fmt.Sprintf(`SELECT task_id,func,data,ttl FROM tasks.tasks WHERE processed_at IS NULL AND locked_by IS NULL AND role = '%s' ORDER BY created_at DESC LIMIT 10;`, role)
		err = r.DB.Select(&tasks, sq)
		if err != nil {
			log.Error(fmt.Sprintf(`tasks.Work() Selecting Task ! %s`, err))
		}

		for _, task := range tasks {
			err = task.Lock()
			if err != nil {
				continue
			}

			task.Dequeue()

			// TODO: Add in TTL Logic with error logging.

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
			case "PrecreateCreateDatabase":
				err = PrecreateDatabase(task.Data)
			default:
				err = fmt.Errorf(`tasks.Work() Unknown Task Action %s`, task.Action)
			}
			if err != nil {
				log.Error(fmt.Sprintf(`tasks.Work() Task %+v ! %s`, task, err))
			} else {
				//sq := fmt.Sprintf(`UPDATE tasks.tasks SET processed_at=CURRENT_TIMESTAMP WHERE task_id='%s';`, task.TaskId)
				sq := fmt.Sprintf(`DELETE FROM tasks.tasks WHERE task_id='%s';`, task.TaskId)
				// TODO: VACUUM FULL the tasks.tasks EVERY MINUTE
				_, err = r.DB.Exec(sq)
				if err != nil {
					log.Error(fmt.Sprintf(`tasks.Work() Error setting processed_at for task %s ! %s`, task.TaskId, err))
				}
			}
			task.Unlock()
		}

		time.Sleep(1 * time.Second)

	}
}
