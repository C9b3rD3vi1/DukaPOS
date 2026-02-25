package whitelabel

import (
	"errors"

	"github.com/C9b3rD3vi1/DukaPOS/internal/models"
	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

type BrandingConfig struct {
	BrandName           string `json:"brand_name"`
	BrandLogo           string `json:"brand_logo"`
	BrandPrimaryColor   string `json:"brand_primary_color"`
	BrandSecondaryColor string `json:"brand_secondary_color"`
	BrandAccentColor    string `json:"brand_accent_color"`
	BrandFont           string `json:"brand_font"`
	CustomDomain        string `json:"custom_domain"`
	CustomSubdomain     string `json:"custom_subdomain"`
	InvoiceFooter       string `json:"invoice_footer"`
	ReceiptHeader       string `json:"receipt_header"`
	ReceiptFooter       string `json:"receipt_footer"`
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

func (s *Service) GetBranding(shopID uint) (*BrandingConfig, error) {
	var shop models.Shop
	if err := s.db.First(&shop, shopID).Error; err != nil {
		return nil, err
	}

	return &BrandingConfig{
		BrandName:           shop.BrandName,
		BrandLogo:           shop.BrandLogo,
		BrandPrimaryColor:   shop.BrandPrimaryColor,
		BrandSecondaryColor: shop.BrandSecondaryColor,
		BrandAccentColor:    shop.BrandAccentColor,
		BrandFont:           shop.BrandFont,
		CustomDomain:        shop.CustomDomain,
		CustomSubdomain:     shop.CustomSubdomain,
		InvoiceFooter:       shop.InvoiceFooter,
		ReceiptHeader:       shop.ReceiptHeader,
		ReceiptFooter:       shop.ReceiptFooter,
	}, nil
}

func (s *Service) UpdateBranding(shopID uint, config *BrandingConfig) error {
	updates := map[string]interface{}{}

	if config.BrandName != "" {
		updates["brand_name"] = config.BrandName
	}
	if config.BrandLogo != "" {
		updates["brand_logo"] = config.BrandLogo
	}
	if config.BrandPrimaryColor != "" {
		if !isValidHexColor(config.BrandPrimaryColor) {
			return errors.New("invalid primary color format")
		}
		updates["brand_primary_color"] = config.BrandPrimaryColor
	}
	if config.BrandSecondaryColor != "" {
		if !isValidHexColor(config.BrandSecondaryColor) {
			return errors.New("invalid secondary color format")
		}
		updates["brand_secondary_color"] = config.BrandSecondaryColor
	}
	if config.BrandAccentColor != "" {
		if !isValidHexColor(config.BrandAccentColor) {
			return errors.New("invalid accent color format")
		}
		updates["brand_accent_color"] = config.BrandAccentColor
	}
	if config.BrandFont != "" {
		updates["brand_font"] = config.BrandFont
	}
	if config.CustomDomain != "" {
		if !isValidDomain(config.CustomDomain) {
			return errors.New("invalid domain format")
		}
		updates["custom_domain"] = config.CustomDomain
	}
	if config.CustomSubdomain != "" {
		updates["custom_subdomain"] = config.CustomSubdomain
	}
	if config.InvoiceFooter != "" {
		updates["invoice_footer"] = config.InvoiceFooter
	}
	if config.ReceiptHeader != "" {
		updates["receipt_header"] = config.ReceiptHeader
	}
	if config.ReceiptFooter != "" {
		updates["receipt_footer"] = config.ReceiptFooter
	}

	if len(updates) == 0 {
		return errors.New("no valid fields to update")
	}

	return s.db.Model(&models.Shop{}).Where("id = ?", shopID).Updates(updates).Error
}

func (s *Service) ResetBranding(shopID uint) error {
	updates := map[string]interface{}{
		"brand_name":            nil,
		"brand_logo":            nil,
		"brand_primary_color":   nil,
		"brand_secondary_color": nil,
		"brand_accent_color":    nil,
		"brand_font":            nil,
		"custom_domain":         nil,
		"custom_subdomain":      nil,
		"invoice_footer":        nil,
		"receipt_header":        nil,
		"receipt_footer":        nil,
	}

	return s.db.Model(&models.Shop{}).Where("id = ?", shopID).Updates(updates).Error
}

func (s *Service) GenerateCSSVariables(shopID uint) (map[string]string, error) {
	branding, err := s.GetBranding(shopID)
	if err != nil {
		return nil, err
	}

	cssVars := map[string]string{
		"--brand-primary":   "#2563eb",
		"--brand-secondary": "#64748b",
		"--brand-accent":    "#f59e0b",
		"--brand-font":      "system-ui, sans-serif",
	}

	if branding.BrandPrimaryColor != "" {
		cssVars["--brand-primary"] = branding.BrandPrimaryColor
	}
	if branding.BrandSecondaryColor != "" {
		cssVars["--brand-secondary"] = branding.BrandSecondaryColor
	}
	if branding.BrandAccentColor != "" {
		cssVars["--brand-accent"] = branding.BrandAccentColor
	}
	if branding.BrandFont != "" {
		cssVars["--brand-font"] = branding.BrandFont
	}

	return cssVars, nil
}

func (s *Service) GenerateInvoiceHTML(shopID uint, invoiceData interface{}) (string, error) {
	branding, err := s.GetBranding(shopID)
	if err != nil {
		return "", err
	}

	header := branding.ReceiptHeader
	if header == "" {
		header = "INVOICE"
	}

	footer := branding.InvoiceFooter
	if footer == "" {
		footer = "Thank you for your business!"
	}

	return buildInvoiceHTML(header, footer, branding.BrandPrimaryColor), nil
}

func isValidHexColor(color string) bool {
	if len(color) != 7 {
		return false
	}
	if color[0] != '#' {
		return false
	}
	for i := 1; i < 7; i++ {
		c := color[i]
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

func isValidDomain(domain string) bool {
	if len(domain) < 4 || len(domain) > 253 {
		return false
	}
	return true
}

func buildInvoiceHTML(header, footer, primaryColor string) string {
	return `
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: system-ui, sans-serif; padding: 20px; }
        .header { color: ` + primaryColor + `; font-size: 24px; font-weight: bold; margin-bottom: 20px; }
        .footer { margin-top: 30px; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="header">` + header + `</div>
    <div class="content">{{.Content}}</div>
    <div class="footer">` + footer + `</div>
</body>
</html>`
}
