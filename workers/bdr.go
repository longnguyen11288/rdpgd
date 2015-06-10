package workers

func (w *Worker) RemoveDatabases() {
	/*
		r := rdpg.New()
			ids := []string{}
			sq := fmt.Sprintf(`SELECT id from cfsb.instances WHERE ineffective_at IS NOT NULL AND ineffective_at < CURRENT_TIMESTAMP AND decommissioned_at IS NULL`)
			_,err := r.DB.Query(&ids,sq)
			for id in range ids {
				i := FindInstance()

				err = r.DisableDatabase(i.Database)
				if err != nil {
					log.Error(fmt.Sprintf("Instance#Remove(%s) DisableDatabase(%s) ! %s", i.InstanceId, i.Database, err))
					return err
				}

				err := r.BackupDatabase(i.Database)
				// Question, do we need to "stop" the replication group before dropping the database?
				err = r.DropDatabase(i.Database)
				if err != nil {
					log.Error(fmt.Sprintf("Instance#Remove(%s) BackupDatabase(%s) ! %s", i.InstanceId, i.Database, err))
					return err
				}

				err = r.DropUser(i.User)
				if err != nil {
					log.Error(fmt.Sprintf("Instance#Remove(%s) DropUser(%s) ! %s", i.InstanceId, i.User, err))
					return err
				}

				err = r.DropDatabase(i.Database)
				if err != nil {
					log.Error(fmt.Sprintf("Instance#Remove(%s) DropDatabase(%s) ! %s", i.InstanceId, i.Database, err))
					return err
				}
			}
		r.DB.Close()
	*/
}
