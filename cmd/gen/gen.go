package main

import (
	"StudentService/database"

	"gorm.io/gen"
)

func main() {
	db, err := database.CreateGorm()
	if err != nil {
		panic(err)
	}

	g := gen.NewGenerator(gen.Config{
		OutPath:      "./test/dal/query",
		ModelPkgPath: "./test/dal/model",
		WithUnitTest: true,
		Mode:         gen.WithDefaultQuery | gen.WithQueryInterface,
	})

	g.UseDB(db)
	g.ApplyBasic(
		g.GenerateModel("students"),
	)
	g.Execute()
}
