package service

import (
	"Auth/internal/db"
	pb "Auth/internal/generated"
	"Auth/internal/hash"
	"Auth/internal/jwt"
	"context"
	"fmt"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
)

type UserService struct {
	pb.UnimplementedAuthenticationServer
	mu sync.RWMutex // Когда-то точно пригодиться
}

func convertRoleToPB(roleStr string) pb.UserType {
	roleStr = strings.ToUpper(strings.TrimSpace(roleStr))
	value, ok := pb.UserType_value[roleStr]
	if !ok {
		log.Errorf("Неизвестная роль %s была заменена на CLIENT", roleStr)
		return pb.UserType_CLIENT
	} else {
		return pb.UserType(value)
	}
}
func createTokens(id, version int64, role string) (string, string, error) {

	// Создание acces-токена
	accesToken, err := jwt.CreateAccessToken(id, role)
	if err != nil {
		return "", "", fmt.Errorf("ошибка создания acces-токена: %w", err)
	}
	// Создание refresh-токена
	refreshToken, err := jwt.CreateRefreshToken(id, version)
	if err != nil {
		return "", "", fmt.Errorf("ошибка создания refresh-токена: %w", err)
	}
	return accesToken, refreshToken, nil
}

func createProfileWithTokens(user db.User, version int64, role string) (*pb.UserProfileWithTokens, error) {
	// Создание acces-токена
	accesToken, refreshToken, err := createTokens(user.ID, version, role)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	return &pb.UserProfileWithTokens{
		Id:           user.ID,
		Login:        user.Login,
		Role:         convertRoleToPB(user.Role),
		Mail:         user.Mail,
		Phone:        user.Phone,
		AccessToken:  accesToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *UserService) AddUser(ctx context.Context, req *pb.UserRequest) (*pb.UserProfileWithTokens, error) {
	// Создаем базовую роль для регистрации через сайт
	roleStr := "CLIENT"
	// Хешируем пароль
	hashPassword, err := hash.GetHashPassword(req.Password)
	if err != nil {
		return nil, err
	}
	// Добавляем пользователя в БД
	user := db.User{Login: req.Login, Password: hashPassword, Role: roleStr}
	user, err = db.AddUser(ctx, user)
	if err != nil {
		return nil, err
	}
	// Создание и вывод профиля
	return createProfileWithTokens(user, 0, roleStr)
}

func (s *UserService) Auth(ctx context.Context, req *pb.UserRequest) (*pb.UserProfileWithTokens, error) {
	// Поиск пользователя по логину
	user, err := db.GetProfileUserDataByLogin(ctx, req.Login)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения данных для %s: %w", req.Login, err)
	}
	passwordServer, err := db.GetPasswordByLogin(ctx, req.Login)
	if err != nil {
		return nil, err
	}
	// Проверка пароля пользователя
	if !hash.CompareHashPassword(req.Password, passwordServer) {
		return nil, fmt.Errorf("неверный пароль")
	}
	// Обновляем версию RefreshToken
	newVerisonToken, err := db.UpdateTokenById(ctx, user.ID)
	// Создание и вывод профиля
	return createProfileWithTokens(user, newVerisonToken, user.Role)
}

func (s *UserService) GetMainData(ctx context.Context, req *pb.UserAccessRequest) (*pb.UserMainDataResponse, error) {
	// Проверяем валдиность токена
	claims, err := jwt.ParseAccesToken(req.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("недействительный access_token: %w", err)
	}
	// Получаем данные с токена
	userIdFloat, ok := claims["id"].(float64)
	if !ok {
		return nil, fmt.Errorf("не удалось получить id из токена")
	}
	userId := int64(userIdFloat)

	roleStr, ok := claims["role"].(string)
	if !ok {
		return nil, fmt.Errorf("не удалось получить role из токена")
	}
	role := convertRoleToPB(roleStr)

	return &pb.UserMainDataResponse{
		Id:   userId,
		Role: role,
	}, nil

}
func (s *UserService) GetProfile(ctx context.Context, req *pb.UserProfileRequest) (*pb.UserProfile, error) {
	user, err := db.GetProfileUserDataById(ctx, req.Id)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения данных профиля: %w", err)
	}
	role := convertRoleToPB(user.Role)
	return &pb.UserProfile{
		Id:    user.ID,
		Login: user.Login,
		Role:  role,
		Mail:  user.Mail,
		Phone: user.Phone,
	}, nil
}

func (s *UserService) ChangeProfile(ctx context.Context, req *pb.UserProfileWithTokens) (*pb.UserProfileWithTokens, error) {
	// Проверяем валдиность токена
	claims, err := jwt.ParseAccesToken(req.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("недействительный access_token: %w", err)
	}
	// Получаем данные с токена
	userIdFloat, ok := claims["id"].(float64)
	if !ok {
		return nil, fmt.Errorf("не удалось получить id из токена")
	}
	userId := int64(userIdFloat)
	if req.Id == userId {
		err := db.UpdateProfileUserDataById(ctx, req.Id, req.Login, req.Mail, req.Phone)
		if err != nil {
			return nil, err
		}
		// Обновляем версию RefreshToken
		newVerisonToken, err := db.UpdateTokenById(ctx, userId)
		// Получаем токены
		role, err := db.GetRoleById(ctx, userId)
		if err != nil {
			return nil, err
		}
		accessToken, refreshToken, err := createTokens(userId, newVerisonToken, role)
		if err != nil {
			return nil, err
		}
		return &pb.UserProfileWithTokens{
			Id:           req.Id,
			Login:        req.Login,
			Role:         req.Role,
			Mail:         req.Mail,
			Phone:        req.Phone,
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		}, nil
	}
	return nil, fmt.Errorf("не совпадает айди из access-токена и айи пользвователя: %d и %d", userId, req.Id)

}

func (s *UserService) ChangePassword(ctx context.Context, req *pb.PasswordChange) (*pb.TokenResponse, error) {
	// Поиск пользователя по логину
	user, err := db.GetUserDataById(ctx, req.Id)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения данных для пользователя с ID = %d: %w", req.Id, err)
	}
	// Проверка пароля пользователя
	if !hash.CompareHashPassword(req.OldPassword, user.Password) {
		return nil, fmt.Errorf("неверный пароль")
	}
	// Хешируем пароль
	hashPassword, err := hash.GetHashPassword(req.NewPassword)
	if err != nil {
		return nil, fmt.Errorf("ошибка хеширования пароля: %w", err)
	}
	// Изменение пароля пользователя
	err = db.UpdatePasswordById(ctx, req.Id, hashPassword)
	// Обновляем версию RefreshToken
	newVerisonToken, err := db.UpdateTokenById(ctx, user.ID)
	// Получаем токены
	accessToken, refreshToken, err := createTokens(user.ID, newVerisonToken, user.Role)
	return &pb.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *UserService) RefreshAccessToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.AccessTokenResponse, error) {
	// Проверяем валдиность токена
	claims, err := jwt.ParseRefreshToken(req.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("недействительный refresh_token: %w", err)
	}
	// Получаем данные с токена
	userIdFloat, ok := claims["id"].(float64)
	if !ok {
		return nil, fmt.Errorf("не удалось получить id из токена")
	}
	userId := int64(userIdFloat)
	// Проверяем версию токена
	verisonFloat, ok := claims["version"].(float64)
	if !ok {
		return nil, fmt.Errorf("не удалось получить version из токена")
	}
	version := int64(verisonFloat)
	verisonServer, err := db.GetTokenVersion(ctx, userId)
	if err != nil {
		return nil, err
	}
	if version != verisonServer {
		return nil, fmt.Errorf("неактуальная версия refresh-токена")
	}

	// Генерируем новый access-токен
	role, err := db.GetRoleById(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения роли пользователя: %w", err)
	}
	accesToken, err := jwt.CreateAccessToken(userId, role)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания access-токена: %w", err)
	}

	return &pb.AccessTokenResponse{
		AccessToken: accesToken,
	}, nil
}
