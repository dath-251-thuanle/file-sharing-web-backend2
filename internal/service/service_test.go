package service

import (
	"context"
	"io"
	"mime/multipart"
	"testing"
	"time"

	"github.com/dath-251-thuanle/file-sharing-web-backend2/config"
	"github.com/dath-251-thuanle/file-sharing-web-backend2/internal/api/dto"
	"github.com/dath-251-thuanle/file-sharing-web-backend2/internal/domain"
	jwt "github.com/dath-251-thuanle/file-sharing-web-backend2/internal/infrastructure/jwt"
	"github.com/dath-251-thuanle/file-sharing-web-backend2/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

func strPtr(s string) *string { return &s }

type MockUserRepo struct{ 
	mock.Mock 
}
func (m *MockUserRepo) FindByEmail(email string, user *domain.User) *utils.ReturnStatus {
	args := m.Called(email, user)
	if args.Get(0) == nil {
		if u, ok := args.Get(1).(domain.User); ok {
			*user = u
		}
		return nil
	}
	return args.Get(0).(*utils.ReturnStatus)
}
func (m *MockUserRepo) AddTimestamp(id, cid string) *utils.ReturnStatus {
	return m.Called(id, cid).Get(0).(*utils.ReturnStatus)
}
func (m *MockUserRepo) FindById(id string, user *domain.User) *utils.ReturnStatus {
	args := m.Called(id, user)
	if args.Get(0) == nil {
		if u, ok := args.Get(1).(domain.User); ok { *user = u }
		return nil
	}
	return args.Get(0).(*utils.ReturnStatus)
}
func (m *MockUserRepo) FindByCId(cid string, u *domain.UsersLoginSession) *utils.ReturnStatus { 
	return nil 
}
func (m *MockUserRepo) DeleteTimestamp(id string) *utils.ReturnStatus { 
	return nil 
}

type MockFileRepo struct{ 
	mock.Mock 
}
func (m *MockFileRepo) CreateFile(ctx context.Context, file *domain.File) (*domain.File, *utils.ReturnStatus) {
	args := m.Called(ctx, file)
	if fn, ok := args.Get(0).(func(context.Context, *domain.File) *domain.File); ok {
		return fn(ctx, file), args.Get(1).(*utils.ReturnStatus)
	}
	return args.Get(0).(*domain.File), args.Get(1).(*utils.ReturnStatus)
}
func (m *MockFileRepo) GetMyFiles(ctx context.Context, uid string, p domain.ListFileParams) ([]domain.File, *utils.ReturnStatus) {
	args := m.Called(ctx, uid, p)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*utils.ReturnStatus)
	}
	return args.Get(0).([]domain.File), args.Get(1).(*utils.ReturnStatus)
}
func (m *MockFileRepo) GetFileByID(ctx context.Context, id string) (*domain.File, *utils.ReturnStatus) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil { return nil, args.Get(1).(*utils.ReturnStatus) }
	return args.Get(0).(*domain.File), args.Get(1).(*utils.ReturnStatus)
}
func (m *MockFileRepo) GetFileByToken(ctx context.Context, t string) (*domain.File, *utils.ReturnStatus) {
	args := m.Called(ctx, t)
	if args.Get(0) == nil { return nil, args.Get(1).(*utils.ReturnStatus) }
	return args.Get(0).(*domain.File), args.Get(1).(*utils.ReturnStatus)
}
func (m *MockFileRepo) RegisterDownload(ctx context.Context, fid, uid string) *utils.ReturnStatus {
	return m.Called(ctx, fid, uid).Get(0).(*utils.ReturnStatus)
}
func (m *MockFileRepo) DeleteFile(ctx context.Context, id string) *utils.ReturnStatus {
	return m.Called(ctx, id).Get(0).(*utils.ReturnStatus)
}
func (m *MockFileRepo) GetFileSummary(ctx context.Context, uid string) (*domain.FileSummary, *utils.ReturnStatus) { 
	return nil, nil 
}
func (m *MockFileRepo) GetTotalUserFiles(ctx context.Context, uid string) (int, *utils.ReturnStatus) { 
	return 0, nil 
}
func (m *MockFileRepo) FindAll(ctx context.Context) ([]domain.File, *utils.ReturnStatus) { 
	return nil, nil 
}
func (m *MockFileRepo) GetFileDownloadHistory(ctx context.Context, fid string) (*domain.FileDownloadHistory, *utils.ReturnStatus) { 
	return nil, nil 
}
func (m *MockFileRepo) GetFileStats(ctx context.Context, fid string) (*domain.FileStat, *utils.ReturnStatus) {
	return nil, nil 
}
func (m *MockFileRepo) GetAccessibleFiles(ctx context.Context, uid string) ([]domain.File, *utils.ReturnStatus) { 
	return nil, nil
}

type MockStorage struct{ 
	mock.Mock 
}
func (m *MockStorage) SaveFile(f *multipart.FileHeader, name string) (string, *utils.ReturnStatus) {
	args := m.Called(f, name)
	return args.String(0), args.Get(1).(*utils.ReturnStatus)
}
func (m *MockStorage) DeleteFile(name string) *utils.ReturnStatus {
	return m.Called(name).Get(0).(*utils.ReturnStatus)
}
func (m *MockStorage) GetFile(name string) (io.Reader, *utils.ReturnStatus) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*utils.ReturnStatus)
	}
	return args.Get(0).(io.Reader), args.Get(1).(*utils.ReturnStatus)
}

type MockTokenService struct{ 
	mock.Mock 
}
func (m *MockTokenService) GenerateAccessToken(u domain.User) (string, error) {
	args := m.Called(u)
	return args.String(0), args.Error(1)
}

func (m *MockTokenService) ParseToken(s string) (*jwt.Claims, error) {
	args := m.Called(s)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*jwt.Claims), args.Error(1)
}

type MockSharedRepo struct{ 
	mock.Mock 
}
func (m *MockSharedRepo) ShareFileWithUsers(ctx context.Context, fid string, emails []string) *utils.ReturnStatus {
	return nil
}
func (m *MockSharedRepo) GetUsersSharedWith(ctx context.Context, fid string) (*domain.Shared, *utils.ReturnStatus) {
	args := m.Called(ctx, fid)
	if args.Get(0) == nil { return nil, args.Get(1).(*utils.ReturnStatus) }
	return args.Get(0).(*domain.Shared), args.Get(1).(*utils.ReturnStatus)
}

type MockAuthRepo struct{ 
	mock.Mock 
}
func (m *MockAuthRepo) Create(u *domain.User) (*domain.User, *utils.ReturnStatus) { 
	return nil, nil 
}
func (m *MockAuthRepo) BlacklistToken(t string, e time.Time) *utils.ReturnStatus {
	return nil 
}
func (m *MockAuthRepo) IsTokenBlacklisted(t string) (bool, *utils.ReturnStatus) {
	return false, nil 
}
func (m *MockAuthRepo) SaveSecret(u, s string) *utils.ReturnStatus { 
	return nil 
}
func (m *MockAuthRepo) GetSecret(u string) (string, *utils.ReturnStatus) {
	return "", nil 
}
func (m *MockAuthRepo) EnableTOTP(u string) *utils.ReturnStatus { 
	return nil
}

func TestAuthService_Login(t *testing.T) {
	mockUserRepo := new(MockUserRepo)
	mockTokenSvc := new(MockTokenService)
	svc := NewAuthService(mockUserRepo, new(MockAuthRepo), mockTokenSvc)
	email := "test@gmail.com"
	pass := "password@abc123"
	hashed, _ := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)

	t.Run("Success_NoTOTP", func(t *testing.T) {
		user := domain.User{Id: "u1", Email: email, Password: string(hashed), EnableTOTP: false}
		mockUserRepo.On("FindByEmail", email, mock.Anything).Return(nil, user).Once()
		mockTokenSvc.On("GenerateAccessToken", user).Return("mock_token", nil).Once()
		resUser, token, err := svc.Login(email, pass)
		assert.Nil(t, err)
		assert.Equal(t, "mock_token", token)
		assert.Equal(t, "u1", resUser.Id)
	})

	t.Run("Success_WithTOTP", func(t *testing.T) {
		userTOTP := domain.User{Id: "u2", Email: email, Password: string(hashed), EnableTOTP: true}
		mockUserRepo.On("FindByEmail", email, mock.Anything).Return(nil, userTOTP).Once()
		mockUserRepo.On("AddTimestamp", "u2", mock.AnythingOfType("string")).Return(nil).Once()
		resUser, cid, err := svc.Login(email, pass)
		assert.Nil(t, err)
		assert.NotEmpty(t, cid)
		assert.Equal(t, "u2", resUser.Id)
	})
}

func TestFileService_UploadFile(t *testing.T) {
	mockFileRepo := new(MockFileRepo)
	mockStorage := new(MockStorage)
	cfg := &config.Config{Policy: &config.SystemPolicy{
		MaxFileSizeMB: 10, DefaultValidityDays: 7, MinValidityHours: 1, MaxValidityDays: 30,
	}}
	svc := NewFileService(cfg, mockFileRepo, new(MockSharedRepo), new(MockUserRepo), mockStorage)
	ctx := context.Background()
	header := &multipart.FileHeader{Filename: "test.pdf", Size: 1024}
	req := &dto.UploadRequest{IsPublic: true}
	ownerID := "user1"
	mockStorage.On("SaveFile", header, mock.AnythingOfType("string")).Return("storage_uuid", nil)
	mockFileRepo.On("CreateFile", ctx, mock.AnythingOfType("*domain.File")).Return(
		func(_ context.Context, f *domain.File) *domain.File {
			f.Id = "gen-id"
			return f
		}, nil,
	)

	file, err := svc.UploadFile(ctx, header, req, strPtr(ownerID))
	assert.Nil(t, err)
	assert.Equal(t, "gen-id", file.Id)
	assert.Equal(t, "test.pdf", file.FileName)
	assert.True(t, file.IsPublic)
	assert.Equal(t, 7, file.ValidityDays)
	mockStorage.AssertExpectations(t)
	mockFileRepo.AssertExpectations(t)
}