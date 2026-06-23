package db

func Init() {
	InitMysql()
	//InitRedis()
	//InitMemDb()
}

func Stop() {
	CloseMysql()
}
