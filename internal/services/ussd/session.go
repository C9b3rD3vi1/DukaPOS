package ussd

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// SessionRepository handles USSD session persistence
type SessionRepository struct {
	db *gorm.DB
}

// NewSessionRepository creates a new USSD session repository
func NewSessionRepository(db *gorm.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

// USSDSession represents a persisted USSD session
type USSDSession struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	SessionID      string    `gorm:"uniqueIndex;size:100" json:"session_id"`
	Phone          string    `gorm:"index;size:20" json:"phone"`
	ShopID         uint      `gorm:"index" json:"shop_id"`
	State          string    `gorm:"size:50" json:"state"`
	PreviousState  string    `gorm:"size:50" json:"previous_state"`
	Data           string    `gorm:"type:text" json:"data"`
	MenuLevel      int       `gorm:"default:0" json:"menu_level"`
	IsActive       bool      `gorm:"default:true" json:"is_active"`
	ExpiresAt      time.Time `json:"expires_at"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// Create creates a new USSD session
func (r *SessionRepository) Create(session *USSDSession) error {
	session.CreatedAt = time.Now()
	session.UpdatedAt = time.Now()
	session.ExpiresAt = time.Now().Add(10 * time.Minute)
	return r.db.Create(session).Error
}

// GetBySessionID retrieves session by session ID
func (r *SessionRepository) GetBySessionID(sessionID string) (*USSDSession, error) {
	var session USSDSession
	err := r.db.Where("session_id = ? AND is_active = ?", sessionID, true).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// GetByPhone retrieves active session for phone
func (r *SessionRepository) GetByPhone(phone string) (*USSDSession, error) {
	var session USSDSession
	err := r.db.Where("phone = ? AND is_active = ?", phone, true).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// Update updates session state and data
func (r *SessionRepository) Update(session *USSDSession) error {
	session.UpdatedAt = time.Now()
	session.ExpiresAt = time.Now().Add(10 * time.Minute)
	return r.db.Save(session).Error
}

// UpdateState updates session state
func (r *SessionRepository) UpdateState(sessionID, state string) error {
	return r.db.Model(&USSDSession{}).
		Where("session_id = ?", sessionID).
		Updates(map[string]interface{}{
			"state":       state,
			"updated_at":  time.Now(),
			"expires_at":  time.Now().Add(10 * time.Minute),
		}).Error
}

// UpdateData updates session data
func (r *SessionRepository) UpdateData(sessionID string, data map[string]string) error {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return r.db.Model(&USSDSession{}).
		Where("session_id = ?", sessionID).
		Updates(map[string]interface{}{
			"data":        string(dataJSON),
			"updated_at": time.Now(),
			"expires_at": time.Now().Add(10 * time.Minute),
		}).Error
}

// GetData retrieves session data as map
func (r *SessionRepository) GetData(sessionID string) (map[string]string, error) {
	session, err := r.GetBySessionID(sessionID)
	if err != nil {
		return nil, err
	}

	var data map[string]string
	if session.Data != "" {
		err := json.Unmarshal([]byte(session.Data), &data)
		if err != nil {
			return nil, err
		}
	}

	return data, nil
}

// SetData sets a value in session data
func (r *SessionRepository) SetData(sessionID, key, value string) error {
	data, err := r.GetData(sessionID)
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	if data == nil {
		data = make(map[string]string)
	}
	data[key] = value

	return r.UpdateData(sessionID, data)
}

// Delete removes/inactivates a session
func (r *SessionRepository) Delete(sessionID string) error {
	return r.db.Model(&USSDSession{}).
		Where("session_id = ?", sessionID).
		Updates(map[string]interface{}{
			"is_active": false,
			"updated_at": time.Now(),
		}).Error
}

// CleanExpired removes expired sessions
func (r *SessionRepository) CleanExpired() (int64, error) {
	result := r.db.Where("expires_at < ? AND is_active = ?", time.Now(), true).Delete(&USSDSession{})
	return result.RowsAffected, result.Error
}

// GetByShop retrieves all active sessions for a shop
func (r *SessionRepository) GetByShop(shopID uint) ([]USSDSession, error) {
	var sessions []USSDSession
	err := r.db.Where("shop_id = ? AND is_active = ?", shopID, true).Find(&sessions).Error
	return sessions, err
}

// CountActive returns count of active sessions
func (r *SessionRepository) CountActive() (int64, error) {
	var count int64
	err := r.db.Model(&USSDSession{}).
		Where("is_active = ? AND expires_at > ?", true, time.Now()).
		Count(&count).Error
	return count, err
}

// Migrate creates the sessions table
func (r *SessionRepository) Migrate() error {
	return r.db.AutoMigrate(&USSDSession{})
}

// SessionManager provides USSD session management
type SessionManager struct {
	repo *SessionRepository
}

// NewSessionManager creates a new session manager
func NewSessionManager(repo *SessionRepository) *SessionManager {
	return &SessionManager{repo: repo}
}

// StartSession creates a new session
func (m *SessionManager) StartSession(phone, sessionID string, shopID uint) (*USSDSession, error) {
	session := &USSDSession{
		SessionID:    sessionID,
		Phone:        phone,
		ShopID:       shopID,
		State:        "main",
		PreviousState: "",
		Data:         "{}",
		MenuLevel:   0,
		IsActive:    true,
	}

	if err := m.repo.Create(session); err != nil {
		return nil, err
	}

	return session, nil
}

// GetOrCreateSession gets existing or creates new session
func (m *SessionManager) GetOrCreateSession(phone, sessionID string, shopID uint) (*USSDSession, error) {
	// Try by phone first
	session, err := m.repo.GetByPhone(phone)
	if err == nil {
		return session, nil
	}

	// Try by session ID
	session, err = m.repo.GetBySessionID(sessionID)
	if err == nil {
		return session, nil
	}

	// Create new session
	return m.StartSession(phone, sessionID, shopID)
}

// GetState gets current session state
func (m *SessionManager) GetState(sessionID string) (string, error) {
	session, err := m.repo.GetBySessionID(sessionID)
	if err != nil {
		return "", err
	}
	return session.State, nil
}

// SetState updates session state
func (m *SessionManager) SetState(sessionID, state string) error {
	session, err := m.repo.GetBySessionID(sessionID)
	if err != nil {
		return err
	}

	return m.repo.db.Model(&USSDSession{}).
		Where("session_id = ?", sessionID).
		Updates(map[string]interface{}{
			"state":          state,
			"previous_state": session.State,
			"updated_at":     time.Now(),
		}).Error
}

// GetData gets a value from session data
func (m *SessionManager) GetData(sessionID, key string) (string, error) {
	data, err := m.repo.GetData(sessionID)
	if err != nil {
		return "", err
	}
	return data[key], nil
}

// SetData sets a value in session data
func (m *SessionManager) SetData(sessionID, key, value string) error {
	return m.repo.SetData(sessionID, key, value)
}

// EndSession ends a session
func (m *SessionManager) EndSession(sessionID string) error {
	return m.repo.Delete(sessionID)
}
