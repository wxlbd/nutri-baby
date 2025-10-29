package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/wxlbd/nutri-baby-server/internal/application/dto"
	"github.com/wxlbd/nutri-baby-server/internal/domain/entity"
	"github.com/wxlbd/nutri-baby-server/internal/domain/repository"
	"github.com/wxlbd/nutri-baby-server/pkg/errors"
)

// BabyService 宝宝服务 (去家庭化架构)
type BabyService struct {
	babyRepo         repository.BabyRepository
	collaboratorRepo repository.BabyCollaboratorRepository
	invitationRepo   repository.BabyInvitationRepository
	userRepo         repository.UserRepository
	vaccineService   *VaccineService
}

// NewBabyService 创建宝宝服务
func NewBabyService(
	babyRepo repository.BabyRepository,
	collaboratorRepo repository.BabyCollaboratorRepository,
	invitationRepo repository.BabyInvitationRepository,
	userRepo repository.UserRepository,
	vaccineService *VaccineService,
) *BabyService {
	return &BabyService{
		babyRepo:         babyRepo,
		collaboratorRepo: collaboratorRepo,
		invitationRepo:   invitationRepo,
		userRepo:         userRepo,
		vaccineService:   vaccineService,
	}
}

// CreateBaby 创建宝宝
func (s *BabyService) CreateBaby(ctx context.Context, openID string, req *dto.CreateBabyRequest) (*dto.BabyDTO, error) {
	// 验证日期格式
	if _, err := time.Parse(time.DateOnly, req.BirthDate); err != nil {
		return nil, errors.New(errors.ParamError, "出生日期格式错误，应为YYYY-MM-DD")
	}

	babyID := uuid.New().String()
	now := time.Now().UnixMilli()

	// 创建宝宝实体
	baby := &entity.Baby{
		BabyID:      babyID,
		Name:        req.Name,
		Nickname:    req.Nickname,
		Gender:      req.Gender,
		BirthDate:   req.BirthDate,
		AvatarURL:   req.AvatarURL,
		CreatorID:   openID,
		FamilyGroup: req.FamilyGroup, // 可选的家庭分组
		CreateTime:  now,
		UpdateTime:  now,
	}

	// 创建宝宝
	if err := s.babyRepo.Create(ctx, baby); err != nil {
		return nil, err
	}

	// 获取用户信息(用于协作者记录)
	//user, err := s.userRepo.FindByOpenID(ctx, openID)
	//if err != nil {
	//	return nil, err
	//}

	// 创建者自动成为管理员
	creator := &entity.BabyCollaborator{
		BabyID:     babyID,
		OpenID:     openID,
		Role:       "admin",
		AccessType: "permanent",
		JoinTime:   now,
		UpdateTime: now,
	}

	if err := s.collaboratorRepo.Create(ctx, creator); err != nil {
		return nil, err
	}

	// 如果指定了复制协作者,则批量复制
	if req.CopyCollaboratorsFrom != "" {
		if err := s.copyCollaborators(ctx, req.CopyCollaboratorsFrom, babyID, openID); err != nil {
			// 记录错误但不影响创建宝宝
			// logger.Warn("Failed to copy collaborators", zap.Error(err))
		}
	}

	// 初始化疫苗提醒
	if err := s.vaccineService.InitializeVaccineReminders(ctx, babyID); err != nil {
		// 记录错误但不影响创建宝宝
		// logger.Error("Failed to initialize vaccine reminders", zap.Error(err))
	}

	return &dto.BabyDTO{
		BabyID:      baby.BabyID,
		Name:        baby.Name,
		Nickname:    baby.Nickname,
		Gender:      baby.Gender,
		BirthDate:   baby.BirthDate,
		AvatarURL:   baby.AvatarURL,
		CreatorID:   baby.CreatorID,
		FamilyGroup: baby.FamilyGroup,
		CreateTime:  baby.CreateTime,
		UpdateTime:  baby.UpdateTime,
	}, nil
}

// GetUserBabies 获取用户可访问的宝宝列表
func (s *BabyService) GetUserBabies(ctx context.Context, openID string) ([]dto.BabyDTO, error) {
	babies, err := s.babyRepo.FindByUserID(ctx, openID)
	if err != nil {
		return nil, err
	}

	result := make([]dto.BabyDTO, 0, len(babies))
	for _, baby := range babies {
		result = append(result, dto.BabyDTO{
			BabyID:      baby.BabyID,
			Name:        baby.Name,
			Nickname:    baby.Nickname,
			Gender:      baby.Gender,
			BirthDate:   baby.BirthDate,
			AvatarURL:   baby.AvatarURL,
			CreatorID:   baby.CreatorID,
			FamilyGroup: baby.FamilyGroup,
			CreateTime:  baby.CreateTime,
			UpdateTime:  baby.UpdateTime,
		})
	}

	return result, nil
}

// GetBabyDetail 获取宝宝详情
func (s *BabyService) GetBabyDetail(ctx context.Context, babyID, openID string) (*dto.BabyDTO, error) {
	// 检查权限
	if err := s.checkPermission(ctx, babyID, openID); err != nil {
		return nil, err
	}

	baby, err := s.babyRepo.FindByID(ctx, babyID)
	if err != nil {
		return nil, err
	}

	return &dto.BabyDTO{
		BabyID:      baby.BabyID,
		Name:        baby.Name,
		Nickname:    baby.Nickname,
		Gender:      baby.Gender,
		BirthDate:   baby.BirthDate,
		AvatarURL:   baby.AvatarURL,
		CreatorID:   baby.CreatorID,
		FamilyGroup: baby.FamilyGroup,
		CreateTime:  baby.CreateTime,
		UpdateTime:  baby.UpdateTime,
	}, nil
}

// UpdateBaby 更新宝宝信息
func (s *BabyService) UpdateBaby(ctx context.Context, babyID, openID string, req *dto.UpdateBabyRequest) error {
	// 检查编辑权限
	canEdit, err := s.collaboratorRepo.CanEdit(ctx, babyID, openID)
	if err != nil {
		return err
	}
	if !canEdit {
		return errors.New(errors.PermissionDenied, "您没有权限编辑该宝宝信息")
	}

	baby, err := s.babyRepo.FindByID(ctx, babyID)
	if err != nil {
		return err
	}

	// 更新字段
	if req.Name != "" {
		baby.Name = req.Name
	}
	if req.Nickname != "" {
		baby.Nickname = req.Nickname
	}
	if req.Gender != "" {
		baby.Gender = req.Gender
	}
	if req.BirthDate != "" {
		if _, err := time.Parse("2006-01-02", req.BirthDate); err != nil {
			return errors.New(errors.ParamError, "出生日期格式错误，应为YYYY-MM-DD")
		}
		baby.BirthDate = req.BirthDate
	}
	if req.AvatarURL != "" {
		baby.AvatarURL = req.AvatarURL
	}
	if req.FamilyGroup != "" {
		baby.FamilyGroup = req.FamilyGroup
	}

	baby.UpdateTime = time.Now().UnixMilli()

	return s.babyRepo.Update(ctx, baby)
}

// DeleteBaby 删除宝宝
func (s *BabyService) DeleteBaby(ctx context.Context, babyID, openID string) error {
	// 只有管理员可以删除宝宝
	isAdmin, err := s.collaboratorRepo.IsAdmin(ctx, babyID, openID)
	if err != nil {
		return err
	}
	if !isAdmin {
		return errors.New(errors.PermissionDenied, "只有管理员可以删除宝宝")
	}

	return s.babyRepo.Delete(ctx, babyID)
}

// GetCollaborators 获取宝宝的协作者列表
func (s *BabyService) GetCollaborators(ctx context.Context, babyID, openID string) ([]dto.CollaboratorDTO, error) {
	// 检查权限
	if err := s.checkPermission(ctx, babyID, openID); err != nil {
		return nil, err
	}

	collaborators, err := s.collaboratorRepo.FindByBabyID(ctx, babyID)
	if err != nil {
		return nil, err
	}

	result := make([]dto.CollaboratorDTO, 0, len(collaborators))
	for _, collab := range collaborators {
		// 获取用户信息
		user, err := s.userRepo.FindByOpenID(ctx, collab.OpenID)
		if err != nil {
			continue // 跳过无法找到的用户
		}

		result = append(result, dto.CollaboratorDTO{
			OpenID:     collab.OpenID,
			NickName:   user.NickName,
			AvatarURL:  user.AvatarURL,
			Role:       collab.Role,
			AccessType: collab.AccessType,
			ExpiresAt:  collab.ExpiresAt,
			JoinTime:   collab.JoinTime,
		})
	}

	return result, nil
}

// InviteCollaborator 邀请协作者 (微信分享/二维码)
func (s *BabyService) InviteCollaborator(ctx context.Context, babyID, openID string, req *dto.InviteCollaboratorRequest) (*dto.BabyInvitationDTO, error) {
	// 只有管理员和编辑者可以邀请
	canEdit, err := s.collaboratorRepo.CanEdit(ctx, babyID, openID)
	if err != nil {
		return nil, err
	}
	if !canEdit {
		return nil, errors.New(errors.PermissionDenied, "您没有权限邀请协作者")
	}

	// 获取宝宝信息
	baby, err := s.babyRepo.FindByID(ctx, babyID)
	if err != nil {
		return nil, err
	}

	// 获取邀请人信息
	inviter, err := s.userRepo.FindByOpenID(ctx, openID)
	if err != nil {
		return nil, err
	}

	// 生成临时token
	token := s.generateInvitationToken()
	now := time.Now().UnixMilli()
	validUntil := now + (7 * 24 * 60 * 60 * 1000) // 7天有效期

	// 创建邀请记录
	invitation := &entity.BabyInvitation{
		InvitationID: uuid.New().String(),
		BabyID:       babyID,
		InviterID:    openID,
		Token:        token,
		InviteType:   req.InviteType,
		Role:         req.Role,
		AccessType:   req.AccessType,
		ExpiresAt:    req.ExpiresAt,
		ValidUntil:   validUntil,
		CreateTime:   now,
	}

	if err := s.invitationRepo.Create(ctx, invitation); err != nil {
		return nil, err
	}

	// 构建返回信息
	result := &dto.BabyInvitationDTO{
		BabyID:      babyID,
		Name:        baby.Name,
		InviterName: inviter.NickName,
		Role:        req.Role,
		ExpiresAt:   req.ExpiresAt,
		ValidUntil:  validUntil,
	}

	// 根据邀请类型构建不同的参数
	if req.InviteType == "share" {
		// 微信分享参数
		result.ShareParams = &dto.ShareParams{
			Title:    fmt.Sprintf("邀请你一起记录%s的成长", baby.Name),
			Path:     fmt.Sprintf("pages/baby/join?babyId=%s&token=%s", babyID, token),
			ImageURL: baby.AvatarURL, // 使用宝宝头像,如果没有则前端使用默认图
		}
	} else if req.InviteType == "qrcode" {
		// 二维码参数
		scene := fmt.Sprintf("b=%s&t=%s", babyID, token)
		result.QRCodeParams = &dto.QRCodeParams{
			Scene: scene,
			Page:  "pages/baby/join",
			// QRCodeURL 由前端调用微信 API 生成或后端生成后返回
		}
	}

	return result, nil
}

// JoinBaby 加入宝宝协作 (通过微信分享或二维码)
func (s *BabyService) JoinBaby(ctx context.Context, openID string, req *dto.JoinBabyRequest) (*dto.BabyDTO, error) {
	// 查找邀请记录
	invitation, err := s.invitationRepo.FindByToken(ctx, req.Token)
	if err != nil {
		return nil, err
	}

	// 验证邀请是否有效
	if invitation.IsExpired() {
		return nil, errors.New(errors.ParamError, "邀请已过期")
	}

	if invitation.IsUsed() {
		return nil, errors.New(errors.ParamError, "邀请已被使用")
	}

	if invitation.BabyID != req.BabyID {
		return nil, errors.New(errors.ParamError, "邀请参数不匹配")
	}

	// 检查用户是否已经是协作者
	existing, err := s.collaboratorRepo.FindByBabyAndUser(ctx, req.BabyID, openID)
	if err != nil && !errors.Is(err, errors.ErrNotFound) {
		return nil, err
	}

	if existing != nil {
		return nil, errors.New(errors.ParamError, "您已经是该宝宝的协作者")
	}

	now := time.Now().UnixMilli()

	// 创建协作者记录
	collaborator := &entity.BabyCollaborator{
		BabyID:     invitation.BabyID,
		OpenID:     openID,
		Role:       invitation.Role,
		AccessType: invitation.AccessType,
		ExpiresAt:  invitation.ExpiresAt,
		JoinTime:   now,
		UpdateTime: now,
	}

	if err := s.collaboratorRepo.Create(ctx, collaborator); err != nil {
		return nil, err
	}

	// 标记邀请已使用
	if err := s.invitationRepo.MarkAsUsed(ctx, invitation.InvitationID, openID, now); err != nil {
		// 记录错误但不影响加入
		// logger.Warn("Failed to mark invitation as used", zap.Error(err))
	}

	// 返回宝宝信息
	return s.GetBabyDetail(ctx, req.BabyID, openID)
}

// RemoveCollaborator 移除协作者
func (s *BabyService) RemoveCollaborator(ctx context.Context, babyID, openID, targetOpenID string) error {
	// 只有管理员可以移除协作者
	isAdmin, err := s.collaboratorRepo.IsAdmin(ctx, babyID, openID)
	if err != nil {
		return err
	}
	if !isAdmin {
		return errors.New(errors.PermissionDenied, "只有管理员可以移除协作者")
	}

	// 不能移除创建者
	baby, err := s.babyRepo.FindByID(ctx, babyID)
	if err != nil {
		return err
	}
	if baby.CreatorID == targetOpenID {
		return errors.New(errors.ParamError, "不能移除创建者")
	}

	return s.collaboratorRepo.Delete(ctx, babyID, targetOpenID)
}

// UpdateCollaboratorRole 更新协作者角色
func (s *BabyService) UpdateCollaboratorRole(ctx context.Context, babyID, openID, targetOpenID, newRole string) error {
	// 只有管理员可以更新角色
	isAdmin, err := s.collaboratorRepo.IsAdmin(ctx, babyID, openID)
	if err != nil {
		return err
	}
	if !isAdmin {
		return errors.New(errors.PermissionDenied, "只有管理员可以更新协作者角色")
	}

	// 不能修改创建者角色
	baby, err := s.babyRepo.FindByID(ctx, babyID)
	if err != nil {
		return err
	}
	if baby.CreatorID == targetOpenID {
		return errors.New(errors.ParamError, "不能修改创建者的角色")
	}

	collaborator, err := s.collaboratorRepo.FindByBabyAndUser(ctx, babyID, targetOpenID)
	if err != nil {
		return err
	}
	if collaborator == nil {
		return errors.New(errors.NotFound, "协作者不存在")
	}

	collaborator.Role = newRole
	collaborator.UpdateTime = time.Now().UnixMilli()

	return s.collaboratorRepo.Update(ctx, collaborator)
}

// checkPermission 检查用户是否有权限访问宝宝
func (s *BabyService) checkPermission(ctx context.Context, babyID, openID string) error {
	isCollaborator, err := s.collaboratorRepo.IsCollaborator(ctx, babyID, openID)
	if err != nil {
		return err
	}
	if !isCollaborator {
		return errors.New(errors.PermissionDenied, "您没有权限访问该宝宝")
	}
	return nil
}

// copyCollaborators 复制协作者列表到新宝宝
func (s *BabyService) copyCollaborators(ctx context.Context, sourceBabyID, targetBabyID, openID string) error {
	// 检查源宝宝的权限
	if err := s.checkPermission(ctx, sourceBabyID, openID); err != nil {
		return err
	}

	// 获取源宝宝的协作者
	sourceCollaborators, err := s.collaboratorRepo.FindByBabyID(ctx, sourceBabyID)
	if err != nil {
		return err
	}

	// 创建新的协作者列表(排除创建者,因为已经添加)
	newCollaborators := make([]*entity.BabyCollaborator, 0)
	now := time.Now().UnixMilli()

	for _, collab := range sourceCollaborators {
		if collab.OpenID == openID {
			continue // 跳过创建者
		}

		newCollaborators = append(newCollaborators, &entity.BabyCollaborator{
			BabyID:     targetBabyID,
			OpenID:     collab.OpenID,
			Role:       collab.Role,
			AccessType: collab.AccessType,
			ExpiresAt:  collab.ExpiresAt,
			JoinTime:   now,
			UpdateTime: now,
		})
	}

	return s.collaboratorRepo.BatchCreate(ctx, newCollaborators)
}

// generateInvitationCode 生成邀请码(已废弃,保留兼容)
func (s *BabyService) generateInvitationCode() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// generateInvitationToken 生成邀请token(用于微信分享和二维码)
func (s *BabyService) generateInvitationToken() string {
	bytes := make([]byte, 32) // 64位十六进制字符串
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
