package repository

import (
	"database/sql"
	"strconv"
	"time"

	"backend/internal/model"

	"github.com/google/uuid"
)

// SystemRepository 系统配置仓库
type SystemRepository struct {
	db *sql.DB
}

// NewSystemRepository 创建系统配置仓库
func NewSystemRepository(db *sql.DB) *SystemRepository {
	return &SystemRepository{
		db: db,
	}
}

// GetConfig 获取系统配置
func (r *SystemRepository) GetConfig(key string) (*model.SystemConfig, error) {
	query := `
		SELECT id, config_key, config_value, config_type, description, updated_at
		FROM system_config
		WHERE config_key = $1
	`

	var config model.SystemConfig
	err := r.db.QueryRow(query, key).Scan(
		&config.ID,
		&config.ConfigKey,
		&config.ConfigValue,
		&config.ConfigType,
		&config.Description,
		&config.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, err
	}

	return &config, nil
}

// GetAllConfigs 获取所有系统配置
func (r *SystemRepository) GetAllConfigs() ([]*model.SystemConfig, error) {
	query := `
		SELECT id, config_key, config_value, config_type, description, updated_at
		FROM system_config
		ORDER BY id
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []*model.SystemConfig
	for rows.Next() {
		var config model.SystemConfig
		err := rows.Scan(
			&config.ID,
			&config.ConfigKey,
			&config.ConfigValue,
			&config.ConfigType,
			&config.Description,
			&config.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		configs = append(configs, &config)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return configs, nil
}

// UpdateConfig 更新系统配置
func (r *SystemRepository) UpdateConfig(key string, value string) error {
	query := `
		UPDATE system_config
		SET config_value = $1, updated_at = $2
		WHERE config_key = $3
	`

	_, err := r.db.Exec(query, value, time.Now().UTC(), key)
	return err
}

// GetSchedulingConfig 获取调度配置
func (r *SystemRepository) GetSchedulingConfig() (*model.SchedulingConfig, error) {
	// 查询所有需要的配置项
	configs, err := r.GetAllConfigs()
	if err != nil {
		return nil, err
	}

	// 将配置映射到map便于访问
	configMap := make(map[string]string)
	for _, config := range configs {
		configMap[config.ConfigKey] = config.ConfigValue
	}
	// 创建默认配置
	schedulingConfig := model.SchedulingConfig{
		Strategy:               "shortest_completion_time",
		FastChargingPileNum:    2,
		SlowChargingPileNum:    3,
		WaitingAreaSize:        6,
		ChargingQueueLen:       2,
		FastChargingPower:      30.0,
		SlowChargingPower:      7.0,
		ExtendedSchedulingMode: model.ExtendedModeDisabled,
	}

	// 应用配置，如果存在的话
	if val, ok := configMap["scheduling_strategy"]; ok {
		schedulingConfig.Strategy = val
	}

	if val, ok := configMap["fast_charging_pile_num"]; ok {
		if num, err := strconv.Atoi(val); err == nil {
			schedulingConfig.FastChargingPileNum = num
		}
	}

	if val, ok := configMap["slow_charging_pile_num"]; ok {
		if num, err := strconv.Atoi(val); err == nil {
			schedulingConfig.SlowChargingPileNum = num
		}
	}

	if val, ok := configMap["waiting_area_size"]; ok {
		if num, err := strconv.Atoi(val); err == nil {
			schedulingConfig.WaitingAreaSize = num
		}
	}

	if val, ok := configMap["charging_queue_len"]; ok {
		if num, err := strconv.Atoi(val); err == nil {
			schedulingConfig.ChargingQueueLen = num
		}
	}

	if val, ok := configMap["fast_charging_power"]; ok {
		if num, err := strconv.ParseFloat(val, 64); err == nil {
			schedulingConfig.FastChargingPower = num
		}
	}
	if val, ok := configMap["slow_charging_power"]; ok {
		if num, err := strconv.ParseFloat(val, 64); err == nil {
			schedulingConfig.SlowChargingPower = num
		}
	}

	if val, ok := configMap["extended_scheduling_mode"]; ok {
		schedulingConfig.ExtendedSchedulingMode = model.ExtendedSchedulingMode(val)
	}

	return &schedulingConfig, nil
}

// CreateFaultRecord 创建故障记录
func (r *SystemRepository) CreateFaultRecord(record *model.FaultRecord) error {
	query := `
		INSERT INTO fault_records 
		(id, pile_id, fault_type, description, occurred_at, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	now := time.Now().UTC()
	_, err := r.db.Exec(
		query,
		uuid.New(),
		record.PileID,
		record.FaultType,
		record.Description,
		record.OccurredAt,
		record.Status,
		now,
	)

	return err
}

// UpdateFaultRecord 更新故障记录
func (r *SystemRepository) UpdateFaultRecord(id uuid.UUID, recoveredAt time.Time, affectedSessions int) error {
	query := `
		UPDATE fault_records
		SET recovered_at = $1, affected_sessions = $2, status = 'resolved'
		WHERE id = $3
	`

	_, err := r.db.Exec(query, recoveredAt, affectedSessions, id)
	return err
}

// GetActiveFaultByPileID 通过充电桩ID获取活跃故障
func (r *SystemRepository) GetActiveFaultByPileID(pileID string) (*model.FaultRecord, error) {
	query := `
		SELECT id, pile_id, fault_type, description, occurred_at, recovered_at, affected_sessions, status, created_at
		FROM fault_records
		WHERE pile_id = $1 AND status = 'active'
		ORDER BY occurred_at DESC
		LIMIT 1
	`

	var record model.FaultRecord
	var recoveredAt sql.NullTime
	err := r.db.QueryRow(query, pileID).Scan(
		&record.ID,
		&record.PileID,
		&record.FaultType,
		&record.Description,
		&record.OccurredAt,
		&recoveredAt,
		&record.AffectedSessions,
		&record.Status,
		&record.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // 没有活跃故障
		}
		return nil, err
	}

	if recoveredAt.Valid {
		record.RecoveredAt = &recoveredAt.Time
	}

	return &record, nil
}

// GetFaultRecords 获取故障记录
func (r *SystemRepository) GetFaultRecords(startTime, endTime time.Time, pileID *string, page, pageSize int) ([]*model.FaultRecord, int, error) {
	// 构建基本查询和参数
	var countQuery, dataQuery string
	var countArgs, dataArgs []any
	countArgs = append(countArgs, startTime, endTime)
	dataArgs = append(dataArgs, startTime, endTime)

	// 根据是否有充电桩ID构建不同的查询
	if pileID != nil {
		countQuery = `
			SELECT COUNT(*)
			FROM fault_records
			WHERE occurred_at >= $1 AND occurred_at <= $2 AND pile_id = $3
		`
		countArgs = append(countArgs, *pileID)

		dataQuery = `
			SELECT id, pile_id, fault_type, description, occurred_at, recovered_at, affected_sessions, status, created_at
			FROM fault_records
			WHERE occurred_at >= $1 AND occurred_at <= $2 AND pile_id = $3
			ORDER BY occurred_at DESC
			LIMIT $4 OFFSET $5
		`
		dataArgs = append(dataArgs, *pileID, pageSize, (page-1)*pageSize)
	} else {
		countQuery = `
			SELECT COUNT(*)
			FROM fault_records
			WHERE occurred_at >= $1 AND occurred_at <= $2
		`

		dataQuery = `
			SELECT id, pile_id, fault_type, description, occurred_at, recovered_at, affected_sessions, status, created_at
			FROM fault_records
			WHERE occurred_at >= $1 AND occurred_at <= $2
			ORDER BY occurred_at DESC
			LIMIT $3 OFFSET $4
		`
		dataArgs = append(dataArgs, pageSize, (page-1)*pageSize)
	}

	// 获取总记录数
	var total int
	err := r.db.QueryRow(countQuery, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	rows, err := r.db.Query(dataQuery, dataArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var records []*model.FaultRecord
	for rows.Next() {
		var record model.FaultRecord
		var recoveredAt sql.NullTime
		err := rows.Scan(
			&record.ID,
			&record.PileID,
			&record.FaultType,
			&record.Description,
			&record.OccurredAt,
			&recoveredAt,
			&record.AffectedSessions,
			&record.Status,
			&record.CreatedAt,
		)

		if err != nil {
			return nil, 0, err
		}

		if recoveredAt.Valid {
			record.RecoveredAt = &recoveredAt.Time
		}

		records = append(records, &record)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return records, total, nil
}

// CreateConfig 创建新的系统配置项
func (r *SystemRepository) CreateConfig(key string, value string, configType string, description string) error {
	query := `
		INSERT INTO system_config (config_key, config_value, config_type, description, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (config_key) DO NOTHING
	`

	_, err := r.db.Exec(query, key, value, configType, description, time.Now().UTC())
	return err
}

// HasPricingForDate 检查指定日期是否已有电价配置
func (r *SystemRepository) HasPricingForDate(date string) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1 FROM pricing_config 
			WHERE effective_date = $1
		)
	`

	var exists bool
	err := r.db.QueryRow(query, date).Scan(&exists)
	return exists, err
}

// CreatePricingConfig 创建电价配置
func (r *SystemRepository) CreatePricingConfig(priceType string, unitPrice float64, startTime string, endTime string, serviceFeeRate float64) error {
	query := `
		INSERT INTO pricing_config (price_type, unit_price, start_time, end_time, service_fee_rate, effective_date)
		VALUES ($1, $2, $3, $4, $5, CURRENT_DATE)
		ON CONFLICT DO NOTHING
	`

	_, err := r.db.Exec(query, priceType, unitPrice, startTime, endTime, serviceFeeRate)
	return err
}
