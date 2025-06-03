package repository

import (
	"database/sql"
	"errors"
	"time"

	"backend/internal/model"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// UserRepository 用户仓库
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository 创建用户仓库
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

// Create 创建用户
func (r *UserRepository) Create(user *model.CreateUserRequest) (*model.User, error) {
	// 检查用户名是否已存在
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM users WHERE username = $1", user.Username).Scan(&count)
	if err != nil {
		return nil, err
	}

	if count > 0 {
		return nil, errors.New("用户名已存在")
	}

	// 生成密码哈希
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// 生成新的UUID
	userID := uuid.New()
	now := time.Now().UTC()

	// 插入新用户
	query := `
		INSERT INTO users (id, username, password_hash, user_type, license_plate, battery_capacity, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, username, user_type, license_plate, battery_capacity, created_at, updated_at
	`

	var newUser model.User
	err = r.db.QueryRow(
		query,
		userID,
		user.Username,
		string(hashedPassword),
		model.UserTypeUser,
		user.LicensePlate,
		user.BatteryCapacity,
		now,
		now,
	).Scan(
		&newUser.ID,
		&newUser.Username,
		&newUser.UserType,
		&newUser.LicensePlate,
		&newUser.BatteryCapacity,
		&newUser.CreatedAt,
		&newUser.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &newUser, nil
}

// GetByID 根据ID获取用户
func (r *UserRepository) GetByID(id uuid.UUID) (*model.User, error) {
	query := `
		SELECT id, username, password_hash, user_type, license_plate, battery_capacity, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var user model.User
	err := r.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.UserType,
		&user.LicensePlate,
		&user.BatteryCapacity,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}

	return &user, nil
}

// GetByUsername 根据用户名获取用户
func (r *UserRepository) GetByUsername(username string) (*model.User, error) {
	query := `
		SELECT id, username, password_hash, user_type, license_plate, battery_capacity, created_at, updated_at 
		FROM users 
		WHERE username = $1
	`

	var user model.User
	err := r.db.QueryRow(query, username).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.UserType,
		&user.LicensePlate,
		&user.BatteryCapacity,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}

	return &user, nil
}

// List 获取用户列表
func (r *UserRepository) List(page, pageSize int) ([]*model.User, int, error) {
	// 获取总数
	var total int
	err := r.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	query := `
		SELECT id, username, user_type, license_plate, battery_capacity, created_at, updated_at 
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(query, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []*model.User
	for rows.Next() {
		var user model.User
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.UserType,
			&user.LicensePlate,
			&user.BatteryCapacity,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// Update 更新用户信息
func (r *UserRepository) Update(user *model.User) error {
	query := `
		UPDATE users 
		SET license_plate = $1, battery_capacity = $2, updated_at = $3 
		WHERE id = $4
	`

	_, err := r.db.Exec(
		query,
		user.LicensePlate,
		user.BatteryCapacity,
		time.Now().UTC(),
		user.ID,
	)

	return err
}

// VerifyPassword 验证密码
func (r *UserRepository) VerifyPassword(username, password string) (*model.User, error) {
	user, err := r.GetByUsername(username)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, errors.New("密码错误")
	}

	return user, nil
}

// SaveSession 保存用户会话
func (r *UserRepository) SaveSession(userID uuid.UUID, token string, expiresAt time.Time) error {
	query := `
		INSERT INTO user_sessions (id, user_id, token, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.db.Exec(
		query,
		uuid.New(),
		userID,
		token,
		expiresAt,
		time.Now().UTC(),
	)

	return err
}

// InvalidateSession 使会话无效
func (r *UserRepository) InvalidateSession(token string) error {
	query := `DELETE FROM user_sessions WHERE token = $1`
	_, err := r.db.Exec(query, token)
	return err
}

// GetSessionByToken 通过Token获取会话
func (r *UserRepository) GetSessionByToken(token string) (*uuid.UUID, error) {
	query := `
		SELECT user_id 
		FROM user_sessions 
		WHERE token = $1 AND expires_at > $2
	`

	var userID uuid.UUID
	err := r.db.QueryRow(query, token, time.Now().UTC()).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("会话不存在或已过期")
		}
		return nil, err
	}

	return &userID, nil
}
