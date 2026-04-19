package database

import (
	appAuth "kafka-order-demo/backend/internal/application/auth"
	"kafka-order-demo/backend/internal/domain/auth"

	"gorm.io/gorm"
)

type GormUser struct {
	ID        uint   `gorm:"primaryKey"`
	Username  string `gorm:"unique;not null"`
	Email     string `gorm:"unique;not null"`
	Password  string `gorm:"not null"`
	IsAdmin   bool   `gorm:"default:false"`
	CreatedAt int64
	UpdatedAt int64
}

func (GormUser) TableName() string {
	return "users"
}

type GormUserRepository struct {
	db *gorm.DB
}

func NewGormUserRepository(db *gorm.DB) *GormUserRepository {
	return &GormUserRepository{db: db}
}

var _ appAuth.UserRepository = (*GormUserRepository)(nil)

func (r *GormUserRepository) Create(user *auth.User) error {
	gormUser := &GormUser{
		Username:  user.Username,
		Email:     user.Email,
		Password:  user.Password,
		IsAdmin:   user.IsAdmin,
		CreatedAt: user.CreatedAt.Unix(),
		UpdatedAt: user.UpdatedAt.Unix(),
	}

	err := r.db.Create(gormUser).Error
	if err != nil {
		return err
	}

	user.ID = gormUser.ID
	return nil
}

func (r *GormUserRepository) GetByID(id uint) (*auth.User, error) {
	var gormUser GormUser
	err := r.db.First(&gormUser, id).Error
	if err != nil {
		return nil, err
	}

	return r.toDomainUser(&gormUser), nil
}

func (r *GormUserRepository) GetByUsername(username string) (*auth.User, error) {
	var gormUser GormUser
	err := r.db.Where("username = ?", username).First(&gormUser).Error
	if err != nil {
		return nil, err
	}

	return r.toDomainUser(&gormUser), nil
}

func (r *GormUserRepository) GetByEmail(email string) (*auth.User, error) {
	var gormUser GormUser
	err := r.db.Where("email = ?", email).First(&gormUser).Error
	if err != nil {
		return nil, err
	}

	return r.toDomainUser(&gormUser), nil
}

func (r *GormUserRepository) Update(user *auth.User) error {
	gormUser := &GormUser{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Password:  user.Password,
		IsAdmin:   user.IsAdmin,
		CreatedAt: user.CreatedAt.Unix(),
		UpdatedAt: user.UpdatedAt.Unix(),
	}

	return r.db.Save(gormUser).Error
}

func (r *GormUserRepository) Delete(id uint) error {
	return r.db.Delete(&GormUser{}, id).Error
}

func (r *GormUserRepository) toDomainUser(gormUser *GormUser) *auth.User {
	return &auth.User{
		ID:        gormUser.ID,
		Username:  gormUser.Username,
		Email:     gormUser.Email,
		Password:  gormUser.Password,
		IsAdmin:   gormUser.IsAdmin,
		CreatedAt: timeFromUnix(gormUser.CreatedAt),
		UpdatedAt: timeFromUnix(gormUser.UpdatedAt),
	}
}
