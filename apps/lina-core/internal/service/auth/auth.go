package auth

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/golang-jwt/jwt/v5"
	"github.com/mssola/useragent"
	"golang.org/x/crypto/bcrypt"

	"lina-core/internal/consts"
	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/loginlog"
)

// Service provides authentication operations.
type Service struct {
	loginLogSvc *loginlog.Service
}

// New creates and returns a new Service instance.
func New() *Service {
	return &Service{
		loginLogSvc: loginlog.New(),
	}
}

// Claims defines JWT token claims.
type Claims struct {
	UserId   int    `json:"userId"`
	Username string `json:"username"`
	Status   int    `json:"status"`
	jwt.RegisteredClaims
}

// LoginInput defines input for Login function.
type LoginInput struct {
	Username string
	Password string
}

// LoginOutput defines output for Login function.
type LoginOutput struct {
	AccessToken string
}

// Login verifies credentials and issues JWT token.
func (s *Service) Login(ctx context.Context, in LoginInput) (*LoginOutput, error) {
	// Extract client info for login log
	var ip, browser, osName string
	if r := g.RequestFromCtx(ctx); r != nil {
		ip = r.GetClientIp()
		ua := useragent.New(r.GetHeader("User-Agent"))
		browserName, browserVersion := ua.Browser()
		browser = browserName + " " + browserVersion
		osName = ua.OS()
	}

	recordLoginLog := func(username string, status int, msg string) {
		_ = s.loginLogSvc.Create(ctx, loginlog.CreateInput{
			UserName: username,
			Status:   status,
			Ip:       ip,
			Browser:  browser,
			Os:       osName,
			Msg:      msg,
		})
	}

	// Query user by username (exclude soft-deleted)
	var user *entity.SysUser
	cols := dao.SysUser.Columns()
	err := dao.SysUser.Ctx(ctx).
		Where(do.SysUser{Username: in.Username}).
		WhereNull(cols.DeletedAt).
		Scan(&user)
	if err != nil {
		return nil, err
	}
	if user == nil {
		recordLoginLog(in.Username, 1, "用户名或密码错误")
		return nil, gerror.New("用户名或密码错误")
	}

	// Verify password
	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(in.Password)); err != nil {
		recordLoginLog(in.Username, 1, "用户名或密码错误")
		return nil, gerror.New("用户名或密码错误")
	}

	// Check status
	if user.Status == consts.UserStatusDisabled {
		recordLoginLog(in.Username, 1, "用户已停用")
		return nil, gerror.New("用户已停用")
	}

	// Generate JWT token
	token, err := s.generateToken(ctx, user)
	if err != nil {
		return nil, err
	}

	// Record login time
	_, _ = dao.SysUser.Ctx(ctx).
		Where(do.SysUser{Id: user.Id}).
		Data(do.SysUser{LoginDate: gtime.Now()}).
		Update()

	recordLoginLog(in.Username, 0, "登录成功")
	return &LoginOutput{AccessToken: token}, nil
}

// ParseToken parses and validates JWT token, returns claims.
func (s *Service) ParseToken(ctx context.Context, tokenString string) (*Claims, error) {
	secret := g.Cfg().MustGet(ctx, "jwt.secret").String()
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, gerror.New("无效的Token")
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, gerror.New("无效的Token")
}

// HashPassword hashes password using bcrypt.
func (s *Service) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", gerror.Wrap(err, "密码加密失败")
	}
	return string(hash), nil
}

// Logout records logout login log.
func (s *Service) Logout(ctx context.Context, username string) {
	var ip, browser, osName string
	if r := g.RequestFromCtx(ctx); r != nil {
		ip = r.GetClientIp()
		ua := useragent.New(r.GetHeader("User-Agent"))
		browserName, browserVersion := ua.Browser()
		browser = browserName + " " + browserVersion
		osName = ua.OS()
	}
	_ = s.loginLogSvc.Create(ctx, loginlog.CreateInput{
		UserName: username,
		Status:   0,
		Ip:       ip,
		Browser:  browser,
		Os:       osName,
		Msg:      "登出成功",
	})
}

// generateToken generates JWT token for given user.
func (s *Service) generateToken(ctx context.Context, user *entity.SysUser) (string, error) {
	var (
		secret     = g.Cfg().MustGet(ctx, "jwt.secret").String()
		expireHour = g.Cfg().MustGet(ctx, "jwt.expireHour").Int()
	)
	claims := Claims{
		UserId:   user.Id,
		Username: user.Username,
		Status:   user.Status,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expireHour) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
