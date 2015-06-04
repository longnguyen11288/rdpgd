package cfsb

type Instance struct {
	Id           string
	DatabaseName string
}

func (i *Instance) Provision() {
	// rdpg.CreateUser()
	// rdpg.CreatDatabase()
	// rdpg.CreateReplicationGroup()
}

func (i *Instance) Deprovision() {
	// rdpg.DisableDatabase()
	// rdpg.BackupDatabase()
	// rdpg.DeleteDatabase()
	// rdpg.DeleteUser()
}
