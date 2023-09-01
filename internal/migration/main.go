package main

import (
	"github.com/redhatinsights/payload-tracker-go/internal/config"
	"github.com/redhatinsights/payload-tracker-go/internal/db"
	"github.com/redhatinsights/payload-tracker-go/internal/logging"
	models "github.com/redhatinsights/payload-tracker-go/internal/models/db"
)

func main() {
	logging.InitLogger()

	cfg := config.Get()

	db.DbConnect(cfg)

	db.DB.AutoMigrate(
		&models.Payload{},
	)

//	db.DB.Exec("ALTER SEQUENCE payloads_id_seq AS bigint")
	db.DB.Exec("ALTER TABLE payloads alter column id add generated always as identity")

	logging.Log.Info("DB Migration Complete")
}
