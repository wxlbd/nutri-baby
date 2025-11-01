package persistence

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/wxlbd/nutri-baby-server/internal/domain/entity"
	"github.com/wxlbd/nutri-baby-server/internal/domain/repository"
	"github.com/wxlbd/nutri-baby-server/pkg/errors"
)

// babyInvitationRepositoryImpl 宝宝邀请仓储实现
type babyInvitationRepositoryImpl struct {
	db *gorm.DB
}

// NewBabyInvitationRepository 创建宝宝邀请仓储
func NewBabyInvitationRepository(db *gorm.DB) repository.BabyInvitationRepository {
	return &babyInvitationRepositoryImpl{db: db}
}

// Create 创建邀请
func (r *babyInvitationRepositoryImpl) Create(ctx context.Context, invitation *entity.BabyInvitation) error {
	if err := r.db.WithContext(ctx).Create(invitation).Error; err != nil {
		return errors.Wrap(errors.DatabaseError, "failed to create baby invitation", err)
	}
	return nil
}

// FindByToken 根据token查找邀请
func (r *babyInvitationRepositoryImpl) FindByToken(ctx context.Context, token string) (*entity.BabyInvitation, error) {
	var invitation entity.BabyInvitation
	err := r.db.WithContext(ctx).
		Where("token = ? AND deleted_at IS NULL", token).
		First(&invitation).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New(errors.NotFound, "邀请不存在或已失效")
	}
	if err != nil {
		return nil, errors.Wrap(errors.DatabaseError, "failed to find invitation by token", err)
	}

	return &invitation, nil
}

// FindByShortCode 根据短码查找邀请
func (r *babyInvitationRepositoryImpl) FindByShortCode(ctx context.Context, shortCode string) (*entity.BabyInvitation, error) {
	var invitation entity.BabyInvitation
	err := r.db.WithContext(ctx).
		Where("short_code = ? AND deleted_at IS NULL", shortCode).
		First(&invitation).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New(errors.NotFound, "邀请不存在或已失效")
	}
	if err != nil {
		return nil, errors.Wrap(errors.DatabaseError, "failed to find invitation by short code", err)
	}

	return &invitation, nil
}

// FindByBabyID 查找宝宝的所有邀请记录
func (r *babyInvitationRepositoryImpl) FindByBabyID(ctx context.Context, babyID string) ([]*entity.BabyInvitation, error) {
	var invitations []*entity.BabyInvitation

	err := r.db.WithContext(ctx).
		Where("baby_id = ? AND deleted_at IS NULL", babyID).
		Order("create_time DESC").
		Find(&invitations).Error

	if err != nil {
		return nil, errors.Wrap(errors.DatabaseError, "failed to find invitations by baby id", err)
	}

	return invitations, nil
}

// MarkAsUsed 标记邀请已使用
func (r *babyInvitationRepositoryImpl) MarkAsUsed(ctx context.Context, invitationID, usedBy string, usedAt int64) error {
	err := r.db.WithContext(ctx).
		Model(&entity.BabyInvitation{}).
		Where("invitation_id = ? AND deleted_at IS NULL", invitationID).
		Updates(map[string]interface{}{
			"used_by": usedBy,
			"used_at": usedAt,
		}).Error

	if err != nil {
		return errors.Wrap(errors.DatabaseError, "failed to mark invitation as used", err)
	}

	return nil
}

// CleanExpired 清理过期的邀请
func (r *babyInvitationRepositoryImpl) CleanExpired(ctx context.Context) error {
	now := time.Now().UnixMilli()

	err := r.db.WithContext(ctx).
		Model(&entity.BabyInvitation{}).
		Where("valid_until < ? AND deleted_at IS NULL", now).
		Update("deleted_at", now).Error

	if err != nil {
		return errors.Wrap(errors.DatabaseError, "failed to clean expired invitations", err)
	}

	return nil
}

// Delete 删除邀请(软删除)
func (r *babyInvitationRepositoryImpl) Delete(ctx context.Context, invitationID string) error {
	err := r.db.WithContext(ctx).
		Model(&entity.BabyInvitation{}).
		Where("invitation_id = ?", invitationID).
		Update("deleted_at", time.Now().UnixMilli()).Error

	if err != nil {
		return errors.Wrap(errors.DatabaseError, "failed to delete invitation", err)
	}

	return nil
}
