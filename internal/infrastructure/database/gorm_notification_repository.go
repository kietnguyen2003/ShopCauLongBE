package database

import (
	appNotification "kafka-order-demo/backend/internal/application/notification"
	domainNotification "kafka-order-demo/backend/internal/domain/notification"
	"time"

	"gorm.io/gorm"
)

type GormNotification struct {
	ID        uint   `gorm:"primaryKey"`
	UserID    uint   `gorm:"not null;index"`
	Type      string `gorm:"not null"`
	Title     string `gorm:"not null"`
	Message   string `gorm:"not null"`
	Data      string `gorm:"type:jsonb;default:'{}'"`
	IsRead    bool   `gorm:"default:false;index"`
	ReadAt    *int64
	CreatedAt int64 `gorm:"index"`
	UpdatedAt int64
}

func (GormNotification) TableName() string {
	return "notifications"
}

type GormNotificationRepository struct {
	db *gorm.DB
}

func NewGormNotificationRepository(db *gorm.DB) *GormNotificationRepository {
	return &GormNotificationRepository{db: db}
}

var _ appNotification.NotificationRepository = (*GormNotificationRepository)(nil)

func (r *GormNotificationRepository) Create(notification *domainNotification.Notification) error {
	gormNotification := r.toGormNotification(notification)
	if err := r.db.Create(gormNotification).Error; err != nil {
		return err
	}

	notification.ID = gormNotification.ID
	return nil
}

func (r *GormNotificationRepository) GetByUserID(userID uint, page, limit int) ([]*domainNotification.Notification, int64, error) {
	var gormNotifications []GormNotification
	var total int64

	query := r.db.Model(&GormNotification{}).Where("user_id = ?", userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := query.Order("created_at DESC, id DESC").Limit(limit).Offset(offset).Find(&gormNotifications).Error; err != nil {
		return nil, 0, err
	}

	notifications := make([]*domainNotification.Notification, len(gormNotifications))
	for i, gormNotification := range gormNotifications {
		notifications[i] = r.toDomainNotification(&gormNotification)
	}

	return notifications, total, nil
}

func (r *GormNotificationRepository) GetByIDAndUserID(id, userID uint) (*domainNotification.Notification, error) {
	var gormNotification GormNotification
	if err := r.db.Where("id = ? AND user_id = ?", id, userID).First(&gormNotification).Error; err != nil {
		return nil, err
	}

	return r.toDomainNotification(&gormNotification), nil
}

func (r *GormNotificationRepository) CountUnreadByUserID(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&GormNotification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Count(&count).Error

	return count, err
}

func (r *GormNotificationRepository) Update(notification *domainNotification.Notification) error {
	return r.db.Save(r.toGormNotification(notification)).Error
}

func (r *GormNotificationRepository) MarkAllAsRead(userID uint) error {
	now := time.Now().Unix()
	return r.db.Model(&GormNotification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Updates(map[string]interface{}{
			"is_read":    true,
			"read_at":    now,
			"updated_at": now,
		}).Error
}

func (r *GormNotificationRepository) toGormNotification(notification *domainNotification.Notification) *GormNotification {
	return &GormNotification{
		ID:        notification.ID,
		UserID:    notification.UserID,
		Type:      notification.Type,
		Title:     notification.Title,
		Message:   notification.Message,
		Data:      notification.Data,
		IsRead:    notification.IsRead,
		ReadAt:    unixPtr(notification.ReadAt),
		CreatedAt: notification.CreatedAt.Unix(),
		UpdatedAt: notification.UpdatedAt.Unix(),
	}
}

func (r *GormNotificationRepository) toDomainNotification(gormNotification *GormNotification) *domainNotification.Notification {
	return &domainNotification.Notification{
		ID:        gormNotification.ID,
		UserID:    gormNotification.UserID,
		Type:      gormNotification.Type,
		Title:     gormNotification.Title,
		Message:   gormNotification.Message,
		Data:      gormNotification.Data,
		IsRead:    gormNotification.IsRead,
		ReadAt:    timePtrFromUnix(gormNotification.ReadAt),
		CreatedAt: timeFromUnix(gormNotification.CreatedAt),
		UpdatedAt: timeFromUnix(gormNotification.UpdatedAt),
	}
}
