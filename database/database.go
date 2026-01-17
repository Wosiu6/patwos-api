package database

import (
	"fmt"

	"github.com/Wosiu6/patwos-api/config"
	"github.com/Wosiu6/patwos-api/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

func Migrate(db *gorm.DB) error {
	db.Exec("DELETE FROM article_votes v WHERE NOT EXISTS (SELECT 1 FROM articles a WHERE a.id = v.article_id)")
	db.Exec("DELETE FROM article_votes v WHERE NOT EXISTS (SELECT 1 FROM users u WHERE u.id = v.user_id)")

	return db.AutoMigrate(
		&models.User{},
		&models.Article{},
		&models.Comment{},
		&models.ArticleVote{},
		&models.RevokedToken{},
	)
}
