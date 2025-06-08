package repository

import (
	"database/sql"
	"fmt"
	"time"

	"backend/internal/model"

	"github.com/google/uuid"
)

// BillingRepository 计费仓库
type BillingRepository struct {
	db *sql.DB
}

// NewBillingRepository 创建计费仓库
func NewBillingRepository(db *sql.DB) *BillingRepository {
	return &BillingRepository{
		db: db,
	}
}

// CreateBillingDetail 创建充电详单
func (r *BillingRepository) CreateBillingDetail(bill *model.BillingDetail) (*model.BillingDetail, error) {
	query := `
		INSERT INTO billing_details 
		(id, session_id, user_id, pile_id, charging_capacity, charging_duration,
		 start_time, stop_time, unit_price, price_type, charging_fee, service_fee, total_fee, generated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id, session_id, user_id, pile_id, charging_capacity, charging_duration,
				  start_time, stop_time, unit_price, price_type, charging_fee, service_fee, total_fee, generated_at
	`

	now := time.Now().UTC()

	var newBill model.BillingDetail
	err := r.db.QueryRow(
		query,
		bill.ID,
		bill.SessionID,
		bill.UserID,
		bill.PileID,
		bill.ChargingCapacity,
		bill.ChargingDuration,
		bill.StartTime,
		bill.EndTime,
		bill.UnitPrice,
		bill.PriceType,
		bill.ChargingFee,
		bill.ServiceFee,
		bill.TotalFee,
		now,
	).Scan(
		&newBill.ID,
		&newBill.SessionID,
		&newBill.UserID,
		&newBill.PileID,
		&newBill.ChargingCapacity,
		&newBill.ChargingDuration,
		&newBill.StartTime,
		&newBill.EndTime,
		&newBill.UnitPrice,
		&newBill.PriceType,
		&newBill.ChargingFee,
		&newBill.ServiceFee,
		&newBill.TotalFee,
		&newBill.GeneratedAt,
	)

	if err != nil {
		return nil, err
	}

	return &newBill, nil
}

// GetByID 通过ID获取充电详单
func (r *BillingRepository) GetByID(id uuid.UUID) (*model.BillingDetail, error) {
	query := `
		SELECT id, session_id, user_id, pile_id, charging_capacity, charging_duration,
			   start_time, stop_time, unit_price, price_type, charging_fee, service_fee, total_fee, generated_at
		FROM billing_details
		WHERE id = $1
	`

	var bill model.BillingDetail
	err := r.db.QueryRow(query, id).Scan(
		&bill.ID,
		&bill.SessionID,
		&bill.UserID,
		&bill.PileID,
		&bill.ChargingCapacity,
		&bill.ChargingDuration,
		&bill.StartTime,
		&bill.EndTime,
		&bill.UnitPrice,
		&bill.PriceType,
		&bill.ChargingFee,
		&bill.ServiceFee,
		&bill.TotalFee,
		&bill.GeneratedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, err
	}

	return &bill, nil
}

// GetBySessionID 通过会话ID获取充电详单
func (r *BillingRepository) GetBySessionID(sessionID uuid.UUID) (*model.BillingDetail, error) {
	query := `
		SELECT id, session_id, user_id, pile_id, charging_capacity, charging_duration,
			   start_time, stop_time, unit_price, price_type, charging_fee, service_fee, total_fee, generated_at
		FROM billing_details
		WHERE session_id = $1
	`

	var bill model.BillingDetail
	err := r.db.QueryRow(query, sessionID).Scan(
		&bill.ID,
		&bill.SessionID,
		&bill.UserID,
		&bill.PileID,
		&bill.ChargingCapacity,
		&bill.ChargingDuration,
		&bill.StartTime,
		&bill.EndTime,
		&bill.UnitPrice,
		&bill.PriceType,
		&bill.ChargingFee,
		&bill.ServiceFee,
		&bill.TotalFee,
		&bill.GeneratedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, err
	}

	return &bill, nil
}

// GetUserBillingDetails 获取用户的充电详单
func (r *BillingRepository) GetUserBillingDetails(userID uuid.UUID, startDate, endDate *time.Time, page, pageSize int) ([]*model.BillingDetail, int, error) {
	// 构建基本查询条件
	whereClause := "WHERE user_id = $1"
	countArgs := []interface{}{userID}
	queryArgs := []interface{}{userID}
	paramCount := 1

	// 添加日期过滤条件
	if startDate != nil {
		paramCount++
		whereClause += fmt.Sprintf(" AND start_time >= $%d", paramCount)
		countArgs = append(countArgs, *startDate)
		queryArgs = append(queryArgs, *startDate)
	}

	if endDate != nil {
		paramCount++
		whereClause += fmt.Sprintf(" AND start_time <= $%d", paramCount)
		countArgs = append(countArgs, *endDate)
		queryArgs = append(queryArgs, *endDate)
	}

	// 获取总数
	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM billing_details %s", whereClause)
	err := r.db.QueryRow(countQuery, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	query := fmt.Sprintf(`
		SELECT id, session_id, user_id, pile_id, charging_capacity, charging_duration,
			   start_time, stop_time, unit_price, price_type, charging_fee, service_fee, total_fee, generated_at
		FROM billing_details
		%s
		ORDER BY generated_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, paramCount+1, paramCount+2)

	queryArgs = append(queryArgs, pageSize, offset)
	rows, err := r.db.Query(query, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var bills []*model.BillingDetail
	for rows.Next() {
		var bill model.BillingDetail
		err := rows.Scan(
			&bill.ID,
			&bill.SessionID,
			&bill.UserID,
			&bill.PileID,
			&bill.ChargingCapacity,
			&bill.ChargingDuration,
			&bill.StartTime,
			&bill.EndTime,
			&bill.UnitPrice,
			&bill.PriceType,
			&bill.ChargingFee,
			&bill.ServiceFee,
			&bill.TotalFee,
			&bill.GeneratedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		bills = append(bills, &bill)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return bills, total, nil
}

// GetCurrentPricing 获取当前时间对应的电价配置
func (r *BillingRepository) GetCurrentPricing(t time.Time) (*model.PriceRate, error) {
	hour, min, sec := t.Clock()
	currentTime := time.Date(0, 1, 1, hour, min, sec, 0, time.UTC)

	query := `
		SELECT price_type, unit_price, service_fee_rate
		FROM pricing_config
		WHERE (start_time <= $1 AND end_time > $1)
			OR (start_time > end_time AND (start_time <= $1 OR end_time > $1))
		ORDER BY effective_date DESC
		LIMIT 1
	`

	var priceRate model.PriceRate
	err := r.db.QueryRow(query, currentTime).Scan(
		&priceRate.Period,
		&priceRate.ElectricFee,
		&priceRate.ServiceFee,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			// 如果没有找到匹配的价格，返回一个默认价格
			priceRate.Period = "normal"
			priceRate.ElectricFee = 0.7
			priceRate.ServiceFee = 0.8
			return &priceRate, nil
		}
		return nil, err
	}

	return &priceRate, nil
}

// GetAllPricingConfig 获取所有电价配置
func (r *BillingRepository) GetAllPricingConfig() ([]*model.PricePeriod, error) {
	query := `
		SELECT price_type, start_time, end_time
		FROM pricing_config
		WHERE effective_date = (SELECT MAX(effective_date) FROM pricing_config)
		ORDER BY start_time
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pricePeriods []*model.PricePeriod
	for rows.Next() {
		var period model.PricePeriod
		var startTime, endTime time.Time

		err := rows.Scan(
			&period.Period,
			&startTime,
			&endTime,
		)

		if err != nil {
			return nil, err
		}

		period.TimeRange.StartHour = startTime.Hour()
		period.TimeRange.EndHour = endTime.Hour()
		pricePeriods = append(pricePeriods, &period)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return pricePeriods, nil
}

// GetBillingStatistics 获取计费统计
func (r *BillingRepository) GetBillingStatistics(startTime, endTime time.Time, pileID *string) (*model.BillingStatistics, error) {
	var query string
	var args []any

	if pileID != nil {
		query = `
			SELECT 
				COUNT(*) as count,
				COALESCE(SUM(charging_duration), 0) as total_duration,
				COALESCE(SUM(charging_capacity), 0) as total_capacity,
				COALESCE(SUM(charging_fee), 0) as total_charging_fee,
				COALESCE(SUM(service_fee), 0) as total_service_fee,
				COALESCE(SUM(total_fee), 0) as total_fee
			FROM billing_details
			WHERE start_time >= $1 AND stop_time <= $2 AND pile_id = $3
		`
		args = []any{startTime, endTime, *pileID}
	} else {
		query = `
			SELECT 
				COUNT(*) as count,
				COALESCE(SUM(charging_duration), 0) as total_duration,
				COALESCE(SUM(charging_capacity), 0) as total_capacity,
				COALESCE(SUM(charging_fee), 0) as total_charging_fee,
				COALESCE(SUM(service_fee), 0) as total_service_fee,
				COALESCE(SUM(total_fee), 0) as total_fee
			FROM billing_details
			WHERE start_time >= $1 AND stop_time <= $2
		`
		args = []any{startTime, endTime}
	}

	var stats model.BillingStatistics
	stats.StartTime = startTime
	stats.EndTime = endTime
	if pileID != nil {
		stats.PileID = *pileID
	}

	err := r.db.QueryRow(query, args...).Scan(
		&stats.Count,
		&stats.TotalDuration,
		&stats.TotalCapacity,
		&stats.TotalChargingFee,
		&stats.TotalServiceFee,
		&stats.TotalFee,
	)

	if err != nil {
		return nil, err
	}

	return &stats, nil
}

// CreateFaultRecord 创建故障记录
func (r *BillingRepository) CreateFaultRecord(record *model.FaultRecord) error {
	query := `
		INSERT INTO fault_records 
		(id, pile_id, fault_type, description, occurred_at, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	now := time.Now().UTC()
	_, err := r.db.Exec(
		query,
		record.ID,
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
func (r *BillingRepository) UpdateFaultRecord(id uuid.UUID, recoveredAt time.Time, affectedSessions int) error {
	query := `
		UPDATE fault_records
		SET recovered_at = $1, affected_sessions = $2, status = 'resolved'
		WHERE id = $3
	`

	_, err := r.db.Exec(query, recoveredAt, affectedSessions, id)
	return err
}

// GetActiveFaultByPileID 通过充电桩ID获取活跃故障
func (r *BillingRepository) GetActiveFaultByPileID(pileID string) (*model.FaultRecord, error) {
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
func (r *BillingRepository) GetFaultRecords(startTime, endTime time.Time, pileID *string, page, pageSize int) ([]*model.FaultRecord, int, error) {
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

// GetSystemConfig 获取系统配置
func (r *BillingRepository) GetSystemConfig(key string) (string, error) {
	query := `
		SELECT config_value
		FROM system_config
		WHERE config_key = $1
	`

	var value string
	err := r.db.QueryRow(query, key).Scan(&value)
	if err != nil {
		return "", err
	}

	return value, nil
}

// UpdateSystemConfig 更新系统配置
func (r *BillingRepository) UpdateSystemConfig(key, value string) error {
	query := `
		UPDATE system_config
		SET config_value = $1, updated_at = $2
		WHERE config_key = $3
	`

	_, err := r.db.Exec(query, value, time.Now().UTC(), key)
	return err
}
