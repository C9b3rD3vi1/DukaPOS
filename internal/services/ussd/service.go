package ussd

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Session represents a USSD session
type Session struct {
	ID        string    `json:"id"`
	Phone     string    `json:"phone"`
	ShopID    uint      `json:"shop_id"`
	State     string    `json:"state"`      // current menu state
	Previous  string    `json:"previous"`   // previous state
	Data      map[string]string `json:"data"` // session data
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Menu represents a USSD menu
type Menu struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Options     []Option `json:"options"`
	ParentID    string   `json:"parent_id,omitempty"`
	IsMainMenu  bool     `json:"is_main_menu"`
}

// Option represents a menu option
type Option struct {
	Number     string `json:"number"`
	Text       string `json:"text"`
	Action     string `json:"action"`    // next state or command
	RequiresAuth bool `json:"requires_auth"`
}

// Response represents USSD response
type Response struct {
	SessionID string `json:"session_id"`
	Message   string `json:"message"`
	FreeFlow  string `json:"free_flow"` // "FC" = Free Choice, "FB" = Back, "FC" = Close
	End       bool   `json:"end"`
}

// Service handles USSD menu processing
type Service struct {
	sessions map[string]*Session
	menuTree map[string]*Menu
}

// New creates a new USSD service
func New() *Service {
	s := &Service{
		sessions: make(map[string]*Session),
		menuTree: make(map[string]*Menu),
	}
	s.buildMenuTree()
	return s
}

// buildMenuTree constructs the USSD menu structure
func (s *Service) buildMenuTree() {
	// Main Menu
	s.menuTree["main"] = &Menu{
		ID:         "main",
		Title:      "🏪 DUKAPOS - Digital Duka",
		IsMainMenu: true,
		Options: []Option{
			{Number: "1", Text: "📦 Check Stock", Action: "stock"},
			{Number: "2", Text: "💰 Record Sale", Action: "sale"},
			{Number: "3", Text: "➕ Add Product", Action: "add_product"},
			{Number: "4", Text: "📊 Daily Report", Action: "report"},
			{Number: "5", Text: "💵 Check Profit", Action: "profit"},
			{Number: "6", Text: "⚠️ Low Stock Items", Action: "low_stock"},
			{Number: "7", Text: "💎 My Shop Info", Action: "shop_info"},
			{Number: "0", Text: "❌ Exit", Action: "exit"},
		},
	}

	// Stock Menu
	s.menuTree["stock"] = &Menu{
		ID:    "stock",
		Title: "📦 STOCK OPTIONS",
		Options: []Option{
			{Number: "1", Text: "View All Products", Action: "stock_all"},
			{Number: "2", Text: "Search Product", Action: "stock_search"},
			{Number: "0", Text: "Back to Main", Action: "main"},
		},
	}

	// Sale Menu
	s.menuTree["sale"] = &Menu{
		ID:    "sale",
		Title: "💰 RECORD SALE",
		Options: []Option{
			{Number: "1", Text: "Quick Sell", Action: "sale_quick"},
			{Number: "2", Text: "Select Product", Action: "sale_select"},
			{Number: "0", Text: "Back to Main", Action: "main"},
		},
	}

	// Add Product Menu
	s.menuTree["add_product"] = &Menu{
		ID:    "add_product",
		Title: "➕ ADD PRODUCT",
		Options: []Option{
			{Number: "1", Text: "New Product", Action: "add_new"},
			{Number: "2", Text: "Add Stock to Existing", Action: "add_existing"},
			{Number: "0", Text: "Back to Main", Action: "main"},
		},
	}

	// Report Menu
	s.menuTree["report"] = &Menu{
		ID:    "report",
		Title: "📊 REPORTS",
		Options: []Option{
			{Number: "1", Text: "Today's Report", Action: "report_today"},
			{Number: "2", Text: "This Week", Action: "report_week"},
			{Number: "3", Text: "This Month", Action: "report_month"},
			{Number: "0", Text: "Back to Main", Action: "main"},
		},
	}

	// Shop Info Menu
	s.menuTree["shop_info"] = &Menu{
		ID:    "shop_info",
		Title: "💎 SHOP INFO",
		Options: []Option{
			{Number: "1", Text: "View Profile", Action: "profile"},
			{Number: "2", Text: "Change Price", Action: "change_price"},
			{Number: "3", Text: "Upgrade Plan", Action: "upgrade"},
			{Number: "0", Text: "Back to Main", Action: "main"},
		},
	}
}

// Process handles incoming USSD request
func (s *Service) Process(phone, sessionID, input string) *Response {
	// Clean phone number
	phone = formatPhone(phone)

	// Get or create session
	session := s.getOrCreateSession(sessionID, phone)

	// Handle input
	response := s.handleInput(session, input)

	// Update session
	session.UpdatedAt = time.Now()
	if response.End {
		// Close session
		delete(s.sessions, sessionID)
	}

	return response
}

// getOrCreateSession gets existing or creates new session
func (s *Service) getOrCreateSession(sessionID, phone string) *Session {
	if session, exists := s.sessions[sessionID]; exists {
		return session
	}

	session := &Session{
		ID:        sessionID,
		Phone:     phone,
		State:     "main",
		Data:      make(map[string]string),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	s.sessions[sessionID] = session
	return session
}

// handleInput processes user input
func (s *Service) handleInput(session *Session, input string) *Response {
	input = strings.TrimSpace(input)

	// First request (empty input) - show main menu
	if input == "" {
		return s.showMenu(session.State)
	}

	// Handle numeric input
	if menu, exists := s.menuTree[session.State]; exists {
		for _, opt := range menu.Options {
			if opt.Number == input {
				// Navigate to next state
				session.Previous = session.State
				session.State = opt.Action

				// Handle special actions
				switch opt.Action {
				case "exit":
					return &Response{
						SessionID: session.ID,
						Message:   "👋 Thank you for using DukaPOS!\n\nGoodbye!",
						FreeFlow:  "FB",
						End:       true,
					}
				case "main":
					return s.showMenu("main")
				case "stock_all":
					return s.handleStockAll(session)
				case "report_today":
					return s.handleReportToday(session)
				case "profit":
					return s.handleProfit(session)
				case "low_stock":
					return s.handleLowStock(session)
				default:
					return s.showMenu(opt.Action)
				}
			}
		}
	}

	// Invalid input - show current menu again
	return s.showMenu(session.State)
}

// showMenu displays a menu
func (s *Service) showMenu(menuID string) *Response {
	menu, exists := s.menuTree[menuID]
	if !exists {
		return &Response{
			Message:  "❌ Menu not found",
			FreeFlow: "FB",
			End:      true,
		}
	}

	var message strings.Builder
	message.WriteString(menu.Title)
	message.WriteString("\n\n")

	for _, opt := range menu.Options {
		message.WriteString(opt.Number)
		message.WriteString(". ")
		message.WriteString(opt.Text)
		message.WriteString("\n")
	}

	return &Response{
		SessionID: menu.ID,
		Message:   message.String(),
		FreeFlow:  "FC",
		End:       false,
	}
}

// Handler functions (simplified - would integrate with real services)

func (s *Service) handleStockAll(session *Session) *Response {
	// In production, this would fetch from database
	return &Response{
		SessionID: session.ID,
		Message: `📦 CURRENT STOCK:

1. Milk - 50 units @ KSh 60
2. Bread - 30 units @ KSh 50
3. Eggs - 20 units @ KSh 250
4. Soda - 100 units @ KSh 50
5. Water - 200 units @ KSh 25

Total Value: KSh 18,000

# = Back to Stock Menu`,
		FreeFlow: "FC",
		End:      false,
	}
}

func (s *Service) handleReportToday(session *Session) *Response {
	return &Response{
		SessionID: session.ID,
		Message: `📊 TODAY'S REPORT:

💰 Total Sales: KSh 5,200
📝 Transactions: 15
💵 Total Profit: KSh 1,800

Top Selling:
1. Milk - 8 units
2. Soda - 6 units
3. Bread - 5 units

# = Back to Reports`,
		FreeFlow: "FC",
		End:      false,
	}
}

func (s *Service) handleProfit(session *Session) *Response {
	return &Response{
		SessionID: session.ID,
		Message: `💵 PROFIT SUMMARY:

Today: KSh 1,800
This Week: KSh 12,500
This Month: KSh 45,000

📈 Profit Margin: 35%

# = Back to Main`,
		FreeFlow: "FC",
		End:      false,
	}
}

func (s *Service) handleLowStock(session *Session) *Response {
	return &Response{
		SessionID: session.ID,
		Message: `⚠️ LOW STOCK ALERT:

1. Eggs - 5 units (Min: 10)
2. Sugar - 8 units (Min: 10)

💡 Order soon to avoid stockouts!

# = Back to Main`,
		FreeFlow: "FC",
		End:      false,
	}
}

// formatPhone formats phone number to standard format
func formatPhone(phone string) string {
	// Remove all non-digits
	var digits string
	for _, c := range phone {
		if c >= '0' && c <= '9' {
			digits += string(c)
		}
	}

	// Handle different formats
	if len(digits) == 10 && digits[0] == '0' {
		return "+254" + digits[1:]
	} else if len(digits) == 9 {
		return "+254" + digits
	} else if len(digits) == 12 && digits[:3] == "254" {
		return "+" + digits
	}

	return phone
}

// ParseUSSDRequest parses incoming USSD request
func ParseUSSDRequest(data string) (sessionID, phone, input string, err error) {
	// USSD format from Africa's Talking or similar:
	// sessionId|phoneNumber|text
	parts := strings.Split(data, "|")
	if len(parts) >= 3 {
		sessionID = parts[0]
		phone = parts[1]
		input = parts[2]
	} else {
		err = fmt.Errorf("invalid USSD request format")
	}
	return
}

// FormatUSSDResponse formats response for USSD gateway
func FormatUSSDResponse(resp *Response) string {
	// Format: responseString|freeFlow
	message := strings.ReplaceAll(resp.Message, "\n", "\n")
	return fmt.Sprintf("%s|%s", message, resp.FreeFlow)
}

// GetSession gets a session by ID
func (s *Service) GetSession(sessionID string) (*Session, bool) {
	session, exists := s.sessions[sessionID]
	return session, exists
}

// EndSession ends a USSD session
func (s *Service) EndSession(sessionID string) {
	delete(s.sessions, sessionID)
}

// GetMenu gets a menu by ID
func (s *Service) GetMenu(menuID string) (*Menu, bool) {
	menu, exists := s.menuTree[menuID]
	return menu, exists
}

// GetAllMenus returns all available menus
func (s *Service) GetAllMenus() []*Menu {
	menus := make([]*Menu, 0, len(s.menuTree))
	for _, menu := range s.menuTree {
		menus = append(menus, menu)
	}
	return menus
}

// SessionCount returns number of active sessions
func (s *Service) SessionCount() int {
	return len(s.sessions)
}

// GetMainMenu returns main menu
func (s *Service) GetMainMenu() *Menu {
	return s.menuTree["main"]
}

// USSDSessionState constants
const (
	StateMain        = "main"
	StateStock       = "stock"
	StateSale        = "sale"
	StateAddProduct  = "add_product"
	StateReport      = "report"
	StateShopInfo    = "shop_info"
	StateExit        = "exit"
)

// InputHandler handles specific input based on state
func (s *Service) InputHandler(session *Session, input string) *Response {
	switch session.State {
	case "sale_quick":
		return s.handleSaleQuick(session, input)
	case "add_new":
		return s.handleAddNew(session, input)
	case "add_existing":
		return s.handleAddExisting(session, input)
	case "stock_search":
		return s.handleStockSearch(session, input)
	case "change_price":
		return s.handleChangePrice(session, input)
	default:
		return s.showMenu(session.State)
	}
}

func (s *Service) handleSaleQuick(session *Session, input string) *Response {
	// Parse: product|quantity (e.g., "milk|2")
	parts := strings.Split(input, "|")
	if len(parts) != 2 {
		return &Response{
			SessionID: session.ID,
			Message:   "❌ Invalid format.\n\nUse: product|qty\nExample: milk|2\n\n# = Cancel",
			FreeFlow: "FC",
			End:      false,
		}
	}

	product := parts[0]
	qty, err := strconv.Atoi(parts[1])
	if err != nil || qty <= 0 {
		return &Response{
			SessionID: session.ID,
			Message:   "❌ Invalid quantity.\n\n# = Cancel",
			FreeFlow:  "FC",
			End:       false,
		}
	}

	// In production, process the sale
	return &Response{
		SessionID: session.ID,
		Message:   fmt.Sprintf("✅ SALE RECORDED!\n\n%s x%d = KSh ???\n\nRemaining stock: ???\n\n# = Back to Main", product, qty),
		FreeFlow:  "FC",
		End:       false,
	}
}

func (s *Service) handleAddNew(session *Session, input string) *Response {
	// Parse: name|price|qty (e.g., "biscuits|30|10")
	return &Response{
		SessionID: session.ID,
		Message:   "➕ ADD NEW PRODUCT\n\nEnter: name|price|qty\n\nExample: biscuits|30|10\n\n# = Cancel",
		FreeFlow:  "FC",
		End:       false,
	}
}

func (s *Service) handleAddExisting(session *Session, input string) *Response {
	return &Response{
		SessionID: session.ID,
		Message:   "📦 ADD STOCK\n\nEnter: product|qty\n\nExample: milk|20\n\n# = Cancel",
		FreeFlow:  "FC",
		End:       false,
	}
}

func (s *Service) handleStockSearch(session *Session, input string) *Response {
	if input == "" || input == "#" {
		return s.showMenu("stock")
	}

	// In production, search database
	return &Response{
		SessionID: session.ID,
		Message:   fmt.Sprintf("🔍 Search results for: %s\n\nNo products found.\n\n# = Back to Stock", input),
		FreeFlow:  "FC",
		End:       false,
	}
}

func (s *Service) handleChangePrice(session *Session, input string) *Response {
	return &Response{
		SessionID: session.ID,
		Message:   "💰 CHANGE PRICE\n\nEnter: product|newprice\n\nExample: milk|65\n\n# = Cancel",
		FreeFlow:  "FC",
		End:       false,
	}
}
