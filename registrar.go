package concept

import (
	"github.com/dityaaa/concept/database"
	"github.com/dityaaa/concept/database/mysql"
	"github.com/dityaaa/concept/source"
	"github.com/dityaaa/concept/source/file"
)

func init() {
	database.Register(&mysql.MySQL{}, mysql.Open)

	source.Register(&file.File{}, file.Open)
}
