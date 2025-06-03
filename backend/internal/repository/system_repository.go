package repository

import (
	"database/sql"
	"fmt"
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
		Strategy:            "shortest_completion_time",
		FastChargingPileNum: 2,
		SlowChargingPileNum: 3,
		WaitingAreaSize:     6,
		ChargingQueueLen:    2,
		FastChargingPower:   30.0,
		SlowChargingPower:   7.0,
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

	return &schedulingConfig, nil
}

// UpdateSchedulingConfig 更新调度配置
func (r *SystemRepository) UpdateSchedulingConfig(config *model.SchedulingConfig) error {
	// 开始事务
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	// 准备更新语句
	stmt, err := tx.Prepare(`
		UPDATE system_config
		SET config_value = $1, updated_at = $2
		WHERE config_key = $3
	`)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	now := time.Now().UTC()

	// 更新各项配置
	_, err = stmt.Exec(config.Strategy, now, "scheduling_strategy")
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = stmt.Exec(strconv.Itoa(config.FastChargingPileNum), now, "fast_charging_pile_num")
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = stmt.Exec(strconv.Itoa(config.SlowChargingPileNum), now, "slow_charging_pile_num")
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = stmt.Exec(strconv.Itoa(config.WaitingAreaSize), now, "waiting_area_size")
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = stmt.Exec(strconv.Itoa(config.ChargingQueueLen), now, "charging_queue_len")
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = stmt.Exec(strconv.FormatFloat(config.FastChargingPower, 'f', 2, 64), now, "fast_charging_power")
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = stmt.Exec(strconv.FormatFloat(config.SlowChargingPower, 'f', 2, 64), now, "slow_charging_power")
	if err != nil {
		tx.Rollback()
		return err
	}

	// 提交事务
	return tx.Commit()
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

// GenerateStatisticsReport 生成统计报表
func (r *SystemRepository) GenerateStatisticsReport(period string, date time.Time) (*model.StatisticsReport, error) {
	var startDate, endDate time.Time
	var periodStr string

	// 根据周期确定起止时间
	switch period {
	case "day":
		startDate = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
		endDate = time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 999999999, date.Location())
		periodStr = startDate.Format("2006-01-02")
	case "week":
		// 找到这周的星期一
		weekday := date.Weekday()
		if weekday == 0 { // 周日
			weekday = 7
		}
		startDate = time.Date(date.Year(), date.Month(), date.Day()-int(weekday-1), 0, 0, 0, 0, date.Location())
		endDate = time.Date(date.Year(), date.Month(), date.Day()+(7-int(weekday)), 23, 59, 59, 999999999, date.Location())
		yearNum, weekNum := startDate.ISOWeek()
		periodStr = fmt.Sprintf("%d-W%d", yearNum, weekNum)
	case "month":
		startDate = time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
		endDate = time.Date(date.Year(), date.Month()+1, 0, 23, 59, 59, 999999999, date.Location())
		periodStr = startDate.Format("2006-01")
	default:
		return nil, fmt.Errorf("无效的时间周期: %s", period)
	}

	// 创建报表对象
	report := model.StatisticsReport{
		Period: period,
		Date:   periodStr,
	}

	// 获取所有充电桩的统计信息
	pileQuery := `
		SELECT cp.id, cp.pile_type,
			COUNT(DISTINCT cs.id) as total_sessions,
			COALESCE(SUM(cs.charging_duration), 0) as total_duration,
			COALESCE(SUM(cs.charged_capacity), 0) as total_energy
		FROM charging_piles cp
		LEFT JOIN charging_sessions cs ON cp.id = cs.pile_id
			AND cs.start_time >= $1 AND (cs.end_time <= $2 OR cs.end_time IS NULL)
			AND cs.status IN ('completed', 'interrupted')
		GROUP BY cp.id, cp.pile_type
		ORDER BY cp.id
	`

	rows, err := r.db.Query(pileQuery, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var pileReport model.PileReport
		var pileType string
		err := rows.Scan(
			&pileReport.PileID,
			&pileType,
			&pileReport.TotalSessions,
			&pileReport.TotalDuration,
			&pileReport.TotalEnergy,
		)
		if err != nil {
			return nil, err
		}

		// 计算充电费用和服务费
		billingQuery := `
			SELECT 
				COALESCE(SUM(charging_fee), 0) as charging_fee,
				COALESCE(SUM(service_fee), 0) as service_fee,
				COALESCE(SUM(total_fee), 0) as total_fee
			FROM billing_details
			WHERE pile_id = $1 AND start_time >= $2 AND stop_time <= $3
		`

		err = r.db.QueryRow(billingQuery, pileReport.PileID, startDate, endDate).Scan(
			&pileReport.ChargingFee,
			&pileReport.ServiceFee,
			&pileReport.TotalFee,
		)
		if err != nil {
			return nil, err
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// 获取系统总体统计信息
	systemQuery := `
		SELECT 
			COUNT(*) as total_requests,
			COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed_requests,
			COUNT(CASE WHEN status = 'cancelled' THEN 1 END) as cancelled_requests
		FROM charging_requests
		WHERE created_at >= $1 AND created_at <= $2
	`

	err = r.db.QueryRow(systemQuery, startDate, endDate).Scan(
		&report.SystemReport.TotalRequests,
		&report.SystemReport.CompletedRequests,
		&report.SystemReport.CancelledRequests,
	)
	if err != nil {
		return nil, err
	}

	// 计算平均等待时间（以分钟为单位）
	waitingTimeQuery := `
		SELECT 
			COALESCE(AVG(EXTRACT(EPOCH FROM (cs.start_time - cr.created_at)) / 60), 0) as avg_waiting_time
		FROM charging_sessions cs
		JOIN charging_requests cr ON cs.request_id = cr.id
		WHERE cs.start_time >= $1 AND (cs.end_time <= $2 OR cs.end_time IS NULL)
	`

	err = r.db.QueryRow(waitingTimeQuery, startDate, endDate).Scan(&report.SystemReport.AvgWaitingTime)
	if err != nil {
		return nil, err
	}

	// 计算总收入
	revenueQuery := `
		SELECT COALESCE(SUM(total_fee), 0) as total_revenue
		FROM billing_details
		WHERE start_time >= $1 AND stop_time <= $2
	`

	err = r.db.QueryRow(revenueQuery, startDate, endDate).Scan(&report.SystemReport.TotalRevenue)
	if err != nil {
		return nil, err
	}

	// 首先获取峰时时段
	peakTimeQuery := `
		SELECT 
			EXTRACT(HOUR FROM start_time) as start_hour,
			EXTRACT(HOUR FROM end_time) as end_hour
		FROM pricing_config
		WHERE price_type = 'peak' AND effective_date <= $1
		ORDER BY effective_date DESC
		LIMIT 1
	`

	rows, err = r.db.Query(peakTimeQuery, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type peakTime struct {
		StartHour int
		EndHour   int
	}

	var peakTimes []peakTime
	for rows.Next() {
		var pt peakTime
		err := rows.Scan(&pt.StartHour, &pt.EndHour)
		if err != nil {
			return nil, err
		}
		peakTimes = append(peakTimes, pt)
	}

	// 计算每天的峰时小时数
	var dailyPeakHours int
	for _, pt := range peakTimes {
		if pt.EndHour > pt.StartHour {
			dailyPeakHours += pt.EndHour - pt.StartHour
		} else {
			dailyPeakHours += pt.EndHour + (24 - pt.StartHour)
		}
	}

	// 根据周期计算总峰时时间
	var totalDays int
	switch period {
	case "day":
		totalDays = 1
	case "week":
		totalDays = 7
	case "month":
		// 计算月份的总天数
		totalDays = endDate.Day()
	}

	totalPeakHours := dailyPeakHours * totalDays

	// 计算峰时充电时长
	peakUsageQuery := `
		SELECT COALESCE(SUM(
			CASE 
				WHEN EXTRACT(HOUR FROM start_time) BETWEEN $1 AND $2
					OR EXTRACT(HOUR FROM start_time) BETWEEN $3 AND $4
				THEN charging_duration
				ELSE 0
			END
		), 0) as peak_usage
		FROM charging_sessions
		WHERE start_time >= $5 AND (end_time <= $6 OR end_time IS NULL)
	`

	var peakUsage float64
	if len(peakTimes) >= 2 {
		err = r.db.QueryRow(peakUsageQuery,
			peakTimes[0].StartHour, peakTimes[0].EndHour,
			peakTimes[1].StartHour, peakTimes[1].EndHour,
			startDate, endDate).Scan(&peakUsage)
	} else if len(peakTimes) == 1 {
		err = r.db.QueryRow(peakUsageQuery,
			peakTimes[0].StartHour, peakTimes[0].EndHour,
			-1, -1, // 无效的时间范围
			startDate, endDate).Scan(&peakUsage)
	} else {
		peakUsage = 0
	}

	if err != nil {
		return nil, err
	}

	// 峰时使用率 = 峰时使用时间 / (充电桩数量 * 总峰时时间)
	var totalPiles int
	err = r.db.QueryRow("SELECT COUNT(*) FROM charging_piles").Scan(&totalPiles)
	if err != nil {
		return nil, err
	}

	if totalPiles > 0 && totalPeakHours > 0 {
		report.SystemReport.PeakTimeUsage = (peakUsage / float64(totalPiles*totalPeakHours)) * 100
	} else {
		report.SystemReport.PeakTimeUsage = 0
	}

	return &report, nil
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
