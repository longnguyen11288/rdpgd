package tasks

func BackupDatabase(data string) (err error) {
	//r := rdpg.NewRDPG()
	//key := fmt.Sprintf("rdpg/%s/work/database/%s/backup", r.Datacenter,data)
	//client, _ := api.NewClient(api.DefaultConfig())
	//lock, err := client.LockKey()
	//if err != nil {
	//	log.Error(fmt.Sprintf("worker.BackupDatabase() Error aquiring lock ! %s", err))
	//	return
	//}
	//leaderCh, err := lock.Lock(nil)
	//if err != nil {
	//	log.Error(fmt.Sprintf("worker.BackupDatabase() Error aquiring lock ! %s", err))
	//	return
	//}
	//if leaderCh == nil {
	//	log.Trace(fmt.Sprintf("worker.BackupDatabase() > Not Leader."))
	//	return
	//}
	//log.Trace(fmt.Sprintf("worker.BackupDatabase() > Leader."))

	// Be sure to keep audit history in the rdpg database backups & audit schema.
	// func BackupDatabase(dbname, backup_location) {
	// 	fileName := backup_location + "/" + dbname + epochcalculation
	// 	exec "pg_dump -Fc ... connection info ... "
	// }
	//func BackupDatabase(dbname, backup_location) {
	//	start_at := now
	//	fileName := backup_location + "/" + dbname + epochcalculation
	//	host := somehow get the host or ip of the worker running this task
	//	exec "pg_dump -Fc ... connection info ... "
	//
	//	//Log Backup History
	//	sql := `INSERT INTO history.backup_restores (host, action, started_at, finished_at, file_location, dbname)
	//	        VALUES (` + host + ",'backup'," + start_at + "," + now() + ",'" + fileName + "','" + dbname + "'"
	//}
	//

	return
}

func BackupAllDatabases(data string) (err error) {
	//r := rdpg.NewRDPG()
	//key := fmt.Sprintf("rdpg/%s/work/databases/backup", r.Datacenter,data)
	//client, _ := api.NewClient(api.DefaultConfig())
	//lock, err := client.LockKey()
	//if err != nil {
	//	log.Error(fmt.Sprintf("worker.BackupAllDatabases() Error aquiring lock ! %s", err))
	//	return
	//}
	//leaderCh, err := lock.Lock(nil)
	//if err != nil {
	//	log.Error(fmt.Sprintf("worker.BackupAllDatabases() Error aquiring lock ! %s", err))
	//	return
	//}
	//if leaderCh == nil {
	//	log.Trace(fmt.Sprintf("worker.BackupAllDatabases() > Not Leader."))
	//	return
	//}
	//log.Trace(fmt.Sprintf("worker.BackupAllDatabases() > Leader."))

	// Be sure to keep audit history in the rdpg database backups & audit schema.
	//
	// start_at := now
	// 	//Get list of databases
	// r := rdpg.OpenDB("postgres")
	// 	dbList := rdpg.GetDBList()
	//
	// 	//Get backup location
	// 	backup_location := GetBackupLocation()
	//
	// 	//Perform backup for each DB
	// 	for each dbname in dbList
	// 		BackupDatabase(dbname, backup_location)
	// 	next
	// 	//Perform pg_dumpall for logins/users
	// 	BackupUsers(backup_location)

	return
}

// func BackupUsers(backup_location) {
// 	fileName := backup_location + "/" + users + epochcalculation
// 	exec "pg_dumpall --globals-only ... connection info ... "
// }
