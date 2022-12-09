package concept

type Hooks struct {
	PreMigrate  func(m *Migration)
	PostMigrate func(m *Migration)
	MigrateErr  func(m *Migration, err error)

	PreRollback  func(m *Migration)
	PostRollback func(m *Migration)
	RollbackErr  func(m *Migration, err error)
}
