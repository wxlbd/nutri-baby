package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/wxlbd/nutri-baby-server/internal/application/dto"
	"github.com/wxlbd/nutri-baby-server/internal/domain/entity"
	"github.com/wxlbd/nutri-baby-server/internal/domain/repository"
	"github.com/wxlbd/nutri-baby-server/pkg/errors"
)

// FamilyService 家庭服务
type FamilyService struct {
	familyRepo       repository.FamilyRepository
	familyMemberRepo repository.FamilyMemberRepository
	invitationRepo   repository.InvitationRepository
	userRepo         repository.UserRepository
}

// NewFamilyService 创建家庭服务
func NewFamilyService(
	familyRepo repository.FamilyRepository,
	familyMemberRepo repository.FamilyMemberRepository,
	invitationRepo repository.InvitationRepository,
	userRepo repository.UserRepository,
) *FamilyService {
	return &FamilyService{
		familyRepo:       familyRepo,
		familyMemberRepo: familyMemberRepo,
		invitationRepo:   invitationRepo,
		userRepo:         userRepo,
	}
}

// CreateFamily 创建家庭
func (s *FamilyService) CreateFamily(ctx context.Context, openID string, req *dto.CreateFamilyRequest) (*dto.FamilyDTO, error) {
	familyID := uuid.New().String()
	now := time.Now().UnixMilli()

	// 创建家庭
	family := &entity.Family{
		FamilyID:   familyID,
		FamilyName: req.FamilyName,
		CreatorID:  openID,
		CreateTime: now,
		UpdateTime: now,
	}

	if err := s.familyRepo.Create(ctx, family); err != nil {
		return nil, err
	}

	// 添加创建者为家庭成员
	member := &entity.FamilyMember{
		FamilyID:   familyID,
		OpenID:     openID,
		Role:       "owner",
		JoinTime:   now,
		UpdateTime: now,
	}

	if err := s.familyMemberRepo.Add(ctx, member); err != nil {
		return nil, err
	}

	return &dto.FamilyDTO{
		FamilyID:   family.FamilyID,
		FamilyName: family.FamilyName,
	}, nil
}

// GetFamilyList 获取用户的家庭列表
func (s *FamilyService) GetFamilyList(ctx context.Context, openID string) ([]dto.FamilyDTO, error) {
	// 直接通过openID查找用户所在的家庭列表
	families, err := s.familyRepo.FindByMember(ctx, openID)
	if err != nil {
		return nil, err
	}

	result := make([]dto.FamilyDTO, 0, len(families))
	for _, family := range families {
		result = append(result, dto.FamilyDTO{
			FamilyID:   family.FamilyID,
			FamilyName: family.FamilyName,
		})
	}

	return result, nil
}

// GetFamilyDetail 获取家庭详情
func (s *FamilyService) GetFamilyDetail(ctx context.Context, familyID, openID string) (*dto.FamilyDetailDTO, error) {
	// 验证用户是否为家庭成员
	if err := s.checkFamilyMembership(ctx, familyID, openID); err != nil {
		return nil, err
	}

	// 获取家庭信息
	family, err := s.familyRepo.FindByID(ctx, familyID)
	if err != nil {
		return nil, err
	}

	// 获取家庭成员
	members, err := s.familyMemberRepo.FindByFamilyID(ctx, familyID)
	if err != nil {
		return nil, err
	}

	// 获取成员用户信息
	memberDTOs := make([]dto.FamilyMemberDTO, 0, len(members))
	for _, member := range members {
		user, err := s.userRepo.FindByOpenID(ctx, member.OpenID)
		if err != nil {
			continue
		}

		memberDTOs = append(memberDTOs, dto.FamilyMemberDTO{
			MemberID: "",  // FamilyMember不再有MemberID
			FamilyID: member.FamilyID,
			UserID:   member.OpenID,
			UserInfo: dto.UserInfoDTO{
				OpenID:    user.OpenID,
				NickName:  user.NickName,
				AvatarURL: user.AvatarURL,
			},
			Role:     member.Role,
			JoinTime: member.JoinTime,
		})
	}

	return &dto.FamilyDetailDTO{
		FamilyID:   family.FamilyID,
		FamilyName: family.FamilyName,
		CreateBy:   family.CreatorID,
		CreateTime: family.CreateTime,
		Members:    memberDTOs,
	}, nil
}

// UpdateFamily 更新家庭信息
func (s *FamilyService) UpdateFamily(ctx context.Context, familyID, openID string, req *dto.UpdateFamilyRequest) error {
	// 验证用户权限（必须是owner或admin）
	if err := s.checkFamilyPermission(ctx, familyID, openID, []string{"owner", "admin"}); err != nil {
		return err
	}

	family, err := s.familyRepo.FindByID(ctx, familyID)
	if err != nil {
		return err
	}

	family.FamilyName = req.FamilyName
	family.UpdateTime = time.Now().UnixMilli()

	return s.familyRepo.Update(ctx, family)
}

// DeleteFamily 删除家庭
func (s *FamilyService) DeleteFamily(ctx context.Context, familyID, openID string) error {
	// 验证用户权限（必须是owner）
	if err := s.checkFamilyPermission(ctx, familyID, openID, []string{"owner"}); err != nil {
		return err
	}

	return s.familyRepo.Delete(ctx, familyID)
}

// CreateInvitation 创建邀请
func (s *FamilyService) CreateInvitation(ctx context.Context, openID string, req *dto.CreateInvitationRequest) (*dto.InvitationDTO, error) {
	// 验证用户权限
	if err := s.checkFamilyPermission(ctx, req.FamilyID, openID, []string{"owner", "admin"}); err != nil {
		return nil, err
	}

	// 获取家庭信息
	family, err := s.familyRepo.FindByID(ctx, req.FamilyID)
	if err != nil {
		return nil, err
	}

	// 生成邀请码
	invitationCode := s.generateInvitationCode()
	now := time.Now().UnixMilli()
	expireTime := time.Now().Add(7 * 24 * time.Hour).UnixMilli() // 7天有效期

	invitation := &entity.Invitation{
		InvitationCode: invitationCode,
		FamilyID:       req.FamilyID,
		CreatorID:      openID,
		ExpiresAt:      expireTime,
		CreateTime:     now,
	}

	if err := s.invitationRepo.Create(ctx, invitation); err != nil {
		return nil, err
	}

	return &dto.InvitationDTO{
		InvitationCode: invitation.InvitationCode,
		FamilyID:       invitation.FamilyID,
		FamilyName:     family.FamilyName,
		Role:           req.Role,
		CreateBy:       invitation.CreatorID,
		CreateTime:     invitation.CreateTime,
		ExpireTime:     invitation.ExpiresAt,
	}, nil
}

// JoinFamily 加入家庭
func (s *FamilyService) JoinFamily(ctx context.Context, openID string, req *dto.JoinFamilyRequest) (*dto.FamilyDTO, error) {
	// 查找邀请
	invitation, err := s.invitationRepo.FindByCode(ctx, req.InvitationCode)
	if err != nil {
		if err == errors.ErrRecordNotFound {
			return nil, errors.New(errors.ParamError, "邀请码无效")
		}
		return nil, err
	}

	// 验证邀请是否过期
	if invitation.IsExpired() {
		return nil, errors.New(errors.ParamError, "邀请码已过期")
	}

	// 检查用户是否已经是成员
	members, err := s.familyMemberRepo.FindByFamilyID(ctx, invitation.FamilyID)
	if err != nil {
		return nil, err
	}

	for _, member := range members {
		if member.OpenID == openID {
			return nil, errors.New(errors.ParamError, "您已经是该家庭的成员")
		}
	}

	// 添加成员
	now := time.Now().UnixMilli()
	member := &entity.FamilyMember{
		FamilyID:   invitation.FamilyID,
		OpenID:     openID,
		Role:       "member", // 默认角色为member
		JoinTime:   now,
		UpdateTime: now,
	}

	if err := s.familyMemberRepo.Add(ctx, member); err != nil {
		return nil, err
	}

	// 删除邀请
	if err := s.invitationRepo.Delete(ctx, invitation.InvitationCode); err != nil {
		return nil, err
	}

	// 获取家庭信息
	family, err := s.familyRepo.FindByID(ctx, invitation.FamilyID)
	if err != nil {
		return nil, err
	}

	return &dto.FamilyDTO{
		FamilyID:   family.FamilyID,
		FamilyName: family.FamilyName,
	}, nil
}

// RemoveMember 移除成员
func (s *FamilyService) RemoveMember(ctx context.Context, familyID, memberID, openID string) error {
	// 验证用户权限
	if err := s.checkFamilyPermission(ctx, familyID, openID, []string{"owner", "admin"}); err != nil {
		return err
	}

	// memberID实际上是openID（因为FamilyMember没有单独的ID）
	// 获取要删除的成员信息
	member, err := s.familyMemberRepo.FindByFamilyAndUser(ctx, familyID, memberID)
	if err != nil {
		return err
	}

	// 不能删除owner
	if member.Role == "owner" {
		return errors.New(errors.PermissionDenied, "不能移除家庭所有者")
	}

	return s.familyMemberRepo.Remove(ctx, familyID, memberID)
}

// LeaveFamily 离开家庭
func (s *FamilyService) LeaveFamily(ctx context.Context, familyID, openID string) error {
	// 查找成员
	member, err := s.familyMemberRepo.FindByFamilyAndUser(ctx, familyID, openID)
	if err != nil {
		return errors.New(errors.ParamError, "您不是该家庭的成员")
	}

	// owner不能离开家庭
	if member.Role == "owner" {
		return errors.New(errors.PermissionDenied, "家庭所有者不能离开家庭，请先转让所有权或删除家庭")
	}

	return s.familyMemberRepo.Remove(ctx, familyID, openID)
}

// checkFamilyMembership 检查用户是否为家庭成员
func (s *FamilyService) checkFamilyMembership(ctx context.Context, familyID, openID string) error {
	members, err := s.familyMemberRepo.FindByFamilyID(ctx, familyID)
	if err != nil {
		return err
	}

	for _, member := range members {
		if member.OpenID == openID {
			return nil
		}
	}

	return errors.New(errors.PermissionDenied, "您不是该家庭的成员")
}

// checkFamilyPermission 检查用户权限
func (s *FamilyService) checkFamilyPermission(ctx context.Context, familyID, openID string, allowedRoles []string) error {
	members, err := s.familyMemberRepo.FindByFamilyID(ctx, familyID)
	if err != nil {
		return err
	}

	for _, member := range members {
		if member.OpenID == openID {
			for _, role := range allowedRoles {
				if member.Role == role {
					return nil
				}
			}
			return errors.New(errors.PermissionDenied, "权限不足")
		}
	}

	return errors.New(errors.PermissionDenied, "您不是该家庭的成员")
}

// generateInvitationCode 生成邀请码
func (s *FamilyService) generateInvitationCode() string {
	return fmt.Sprintf("INV%d", time.Now().UnixNano())
}
