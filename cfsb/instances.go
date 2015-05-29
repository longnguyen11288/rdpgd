package instances

import (
	_ "github.com/lib/pq"
)

type Instance struct {
	Id string
	DatabaseName string
}

