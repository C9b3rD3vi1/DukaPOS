package middleware

import (
	"strings"

	"github.com/C9b3rD3vi1/DukaPOS/internal/models"
	"github.com/gofiber/fiber/v2"
)

type SubscriptionPlan string

const (
	PlanFree     SubscriptionPlan = "free"
	PlanPro      SubscriptionPlan = "pro"
	PlanBusiness SubscriptionPlan = "business"
)

type Feature string

const (
	FeatureMpesa         Feature = "mpesa"
	FeatureMultipleShops Feature = "multiple_shops"
	FeatureStaffAccounts Feature = "staff_accounts"
	FeatureAPIAccess     Feature = "api_access"
	FeatureWebhooks      Feature = "webhooks"
	FeatureAI            Feature = "ai_predictions"
	FeatureQRPayments    Feature = "qr_payments"
	FeatureLoyalty       Feature = "loyalty"
	FeatureExport        Feature = "export"
	FeatureMultiCurrency Feature = "multi_currency"
)

type SubscriptionConfig struct {
	Free     PlanLimits
	Pro      PlanLimits
	Business PlanLimits
}

type PlanLimits struct {
	MaxProducts  int
	MaxStaff     int
	MaxShops     int
	MaxCustomers int
	MaxAPIKeys   int
	MaxWebhooks  int
	Features     []Feature
	MonthlyLimit int64
}

var DefaultSubscriptionConfig = SubscriptionConfig{
	Free: PlanLimits{
		MaxProducts:  50,
		MaxStaff:     0,
		MaxShops:     1,
		MaxCustomers: 0,
		MaxAPIKeys:   0,
		MaxWebhooks:  0,
		Features:     []Feature{},
		MonthlyLimit: 0,
	},
	Pro: PlanLimits{
		MaxProducts:  -1,
		MaxStaff:     2,
		MaxShops:     3,
		MaxCustomers: 100,
		MaxAPIKeys:   2,
		MaxWebhooks:  2,
		Features:     []Feature{FeatureMpesa, FeatureStaffAccounts, FeatureQRPayments, FeatureLoyalty},
		MonthlyLimit: 10000,
	},
	Business: PlanLimits{
		MaxProducts:  -1,
		MaxStaff:     5,
		MaxShops:     -1,
		MaxCustomers: -1,
		MaxAPIKeys:   10,
		MaxWebhooks:  10,
		Features: []Feature{
			FeatureMpesa, FeatureMultipleShops, FeatureStaffAccounts,
			FeatureAPIAccess, FeatureWebhooks, FeatureAI,
			FeatureQRPayments, FeatureLoyalty, FeatureExport, FeatureMultiCurrency,
		},
		MonthlyLimit: 100000,
	},
}

func GetPlanLimits(plan models.PlanType) PlanLimits {
	switch plan {
	case models.PlanPro:
		return DefaultSubscriptionConfig.Pro
	case models.PlanBusiness:
		return DefaultSubscriptionConfig.Business
	default:
		return DefaultSubscriptionConfig.Free
	}
}

func HasFeature(plan models.PlanType, feature Feature) bool {
	limits := GetPlanLimits(plan)
	for _, f := range limits.Features {
		if f == feature {
			return true
		}
	}
	return false
}

func RequireFeature(feature Feature) fiber.Handler {
	return func(c *fiber.Ctx) error {
		shop := getShopFromContext(c)
		if shop == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
				"code":  "UNAUTHORIZED",
			})
		}

		if !HasFeature(shop.Plan, feature) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":   "Feature not available on " + string(shop.Plan) + " plan",
				"code":    "PLAN_FEATURE_NOT_ALLOWED",
				"feature": string(feature),
				"current": string(shop.Plan),
				"upgrade": getUpgradePlan(feature),
			})
		}

		return c.Next()
	}
}

func RequirePro() fiber.Handler {
	return func(c *fiber.Ctx) error {
		shop := getShopFromContext(c)
		if shop == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
		}

		if shop.Plan != models.PlanPro && shop.Plan != models.PlanBusiness {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":   "This feature requires Pro plan or higher",
				"code":    "PLAN_REQUIRED",
				"current": string(shop.Plan),
				"upgrade": "Pro",
			})
		}

		return c.Next()
	}
}

func RequireBusiness() fiber.Handler {
	return func(c *fiber.Ctx) error {
		shop := getShopFromContext(c)
		if shop == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
		}

		if shop.Plan != models.PlanBusiness {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":   "This feature requires Business plan",
				"code":    "PLAN_REQUIRED",
				"current": string(shop.Plan),
				"upgrade": "Business",
			})
		}

		return c.Next()
	}
}

func getShopFromContext(c *fiber.Ctx) *models.Shop {
	if shop, ok := c.Locals("shop").(*models.Shop); ok && shop != nil {
		return shop
	}
	if account, ok := c.Locals("account").(*models.Account); ok && account != nil {
		return &models.Shop{Plan: account.Plan}
	}
	return nil
}

func getUpgradePlan(feature Feature) string {
	switch feature {
	case FeatureMpesa, FeatureQRPayments, FeatureMultipleShops:
		return "Pro"
	default:
		return "Business"
	}
}

func EnforceProductLimit() fiber.Handler {
	return func(c *fiber.Ctx) error {
		shop := getShopFromContext(c)
		if shop == nil {
			return c.Next()
		}

		limits := GetPlanLimits(shop.Plan)
		if limits.MaxProducts == -1 {
			return c.Next()
		}

		if count, ok := c.Locals("product_count").(int); ok {
			if count >= limits.MaxProducts {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"error":      "Product limit reached for " + string(shop.Plan) + " plan",
					"code":       "PLAN_LIMIT_REACHED",
					"limit":      limits.MaxProducts,
					"current":    count,
					"upgrade_to": "Pro or Business",
				})
			}
		}

		return c.Next()
	}
}

func EnforceStaffLimit() fiber.Handler {
	return func(c *fiber.Ctx) error {
		shop := getShopFromContext(c)
		if shop == nil {
			return c.Next()
		}

		limits := GetPlanLimits(shop.Plan)
		if limits.MaxStaff == 0 {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":      "Staff accounts require Pro plan",
				"code":       "PLAN_REQUIRED",
				"feature":    "staff_accounts",
				"upgrade_to": "Pro",
			})
		}

		if limits.MaxStaff > 0 {
			if count, ok := c.Locals("staff_count").(int); ok {
				if count >= limits.MaxStaff {
					return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
						"error":      "Staff limit reached",
						"code":       "PLAN_LIMIT_REACHED",
						"limit":      limits.MaxStaff,
						"current":    count,
						"upgrade_to": "Business",
					})
				}
			}
		}

		return c.Next()
	}
}

func EnforceShopLimit() fiber.Handler {
	return func(c *fiber.Ctx) error {
		shop := getShopFromContext(c)
		if shop == nil {
			return c.Next()
		}

		limits := GetPlanLimits(shop.Plan)
		if limits.MaxShops == 1 {
			return c.Next()
		}

		if count, ok := c.Locals("shop_count").(int); ok {
			if limits.MaxShops > 0 && count >= limits.MaxShops {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"error":      "Shop limit reached for " + string(shop.Plan) + " plan",
					"code":       "PLAN_LIMIT_REACHED",
					"limit":      limits.MaxShops,
					"current":    count,
					"upgrade_to": "Pro or Business",
				})
			}
		}

		return c.Next()
	}
}

func EnforceAPIKeyLimit() fiber.Handler {
	return func(c *fiber.Ctx) error {
		shop := getShopFromContext(c)
		if shop == nil {
			return c.Next()
		}

		limits := GetPlanLimits(shop.Plan)
		if limits.MaxAPIKeys == 0 {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":      "API keys require Business plan",
				"code":       "PLAN_REQUIRED",
				"feature":    "api_access",
				"upgrade_to": "Business",
			})
		}

		if count, ok := c.Locals("apikey_count").(int); ok {
			if count >= limits.MaxAPIKeys {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"error":      "API key limit reached",
					"code":       "PLAN_LIMIT_REACHED",
					"limit":      limits.MaxAPIKeys,
					"current":    count,
					"upgrade_to": "Business",
				})
			}
		}

		return c.Next()
	}
}

func EnforceWebhookLimit() fiber.Handler {
	return func(c *fiber.Ctx) error {
		shop := getShopFromContext(c)
		if shop == nil {
			return c.Next()
		}

		limits := GetPlanLimits(shop.Plan)
		if limits.MaxWebhooks == 0 {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":      "Webhooks require Business plan",
				"code":       "PLAN_REQUIRED",
				"feature":    "webhooks",
				"upgrade_to": "Business",
			})
		}

		return c.Next()
	}
}

func GetPlanInfo(plan models.PlanType) fiber.Map {
	limits := GetPlanLimits(plan)

	features := make([]string, len(limits.Features))
	for i, f := range limits.Features {
		features[i] = string(f)
	}

	return fiber.Map{
		"plan":          string(plan),
		"max_products":  limits.MaxProducts,
		"max_staff":     limits.MaxStaff,
		"max_shops":     limits.MaxShops,
		"max_customers": limits.MaxCustomers,
		"max_api_keys":  limits.MaxAPIKeys,
		"max_webhooks":  limits.MaxWebhooks,
		"features":      features,
		"monthly_limit": limits.MonthlyLimit,
		"is_pro":        plan == models.PlanPro || plan == models.PlanBusiness,
		"is_business":   plan == models.PlanBusiness,
	}
}

type PlanInfoHandler struct{}

func NewPlanInfoHandler() *PlanInfoHandler {
	return &PlanInfoHandler{}
}

func (h *PlanInfoHandler) GetPlanInfo(c *fiber.Ctx) error {
	shop := getShopFromContext(c)
	if shop == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	return c.JSON(GetPlanInfo(shop.Plan))
}

func (h *PlanInfoHandler) GetAllPlans(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"free":     GetPlanInfo(models.PlanFree),
		"pro":      GetPlanInfo(models.PlanPro),
		"business": GetPlanInfo(models.PlanBusiness),
	})
}

func CheckPlanFeature(shop *models.Shop, feature Feature) error {
	if !HasFeature(shop.Plan, feature) {
		return fiber.NewError(fiber.StatusForbidden, "feature not available on "+string(shop.Plan)+" plan")
	}
	return nil
}

func GetPlanBadge(plan models.PlanType) string {
	switch plan {
	case models.PlanFree:
		return "Free"
	case models.PlanPro:
		return "Pro"
	case models.PlanBusiness:
		return "Business"
	default:
		return "Unknown"
	}
}

func FormatPlanMessage(plan models.PlanType) string {
	badge := GetPlanBadge(plan)
	limits := GetPlanLimits(plan)

	msg := badge + "\n"
	msg += "━━━━━━━━━━━━━━\n"

	if limits.MaxProducts == -1 {
		msg += "Products: Unlimited\n"
	} else {
		msg += "Products: " + string(rune(limits.MaxProducts+'0')) + "\n"
	}

	if limits.MaxStaff == -1 {
		msg += "Staff: Unlimited\n"
	} else if limits.MaxStaff == 0 {
		msg += "Staff: Not available\n"
	} else {
		msg += "Staff: " + string(rune(limits.MaxStaff+'0')) + "\n"
	}

	msg += "Shops: "
	if limits.MaxShops == -1 {
		msg += "Unlimited\n"
	} else {
		msg += string(rune(limits.MaxShops+'0')) + "\n"
	}

	if len(limits.Features) > 0 {
		featureStrs := make([]string, len(limits.Features))
		for i, f := range limits.Features {
			featureStrs[i] = string(f)
		}
		msg += "Features: " + strings.Join(featureStrs, ", ") + "\n"
	}

	return msg
}
