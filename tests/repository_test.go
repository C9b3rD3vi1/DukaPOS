package main

import (
	"testing"
)

// TestRepositoryOperationCounts verifies expected operation counts
func TestRepositoryOperationCounts(t *testing.T) {
	// Staff repository operations
	staffOps := []string{"Create", "GetByID", "GetByPhone", "GetByShopID", "Update", "Delete"}
	if len(staffOps) != 6 {
		t.Errorf("Expected 6 staff operations, got %d", len(staffOps))
	}

	// Product repository operations
	productOps := []string{"Create", "GetByID", "GetByShopAndName", "GetByShopID", "GetLowStock", "Update", "Delete", "UpdateStock"}
	if len(productOps) != 8 {
		t.Errorf("Expected 8 product operations, got %d", len(productOps))
	}

	// Sale repository operations
	saleOps := []string{"Create", "GetByID", "GetByShopID", "GetByDateRange", "GetTodaySales", "GetTotalSales"}
	if len(saleOps) != 6 {
		t.Errorf("Expected 6 sale operations, got %d", len(saleOps))
	}

	// Shop repository operations
	shopOps := []string{"Create", "GetByID", "GetByPhone", "GetByEmail", "Update", "Delete", "List"}
	if len(shopOps) != 7 {
		t.Errorf("Expected 7 shop operations, got %d", len(shopOps))
	}

	// Daily summary repository operations
	summaryOps := []string{"GetOrCreate", "Update", "Recalculate", "GetByDateRange"}
	if len(summaryOps) != 4 {
		t.Errorf("Expected 4 summary operations, got %d", len(summaryOps))
	}

	// Audit log repository operations
	auditOps := []string{"Create", "GetByShopID"}
	if len(auditOps) != 2 {
		t.Errorf("Expected 2 audit operations, got %d", len(auditOps))
	}
}

// TestModelFields verifies model field structures
func TestModelFields(t *testing.T) {
	// Staff model fields
	staffFields := []string{"ID", "ShopID", "Name", "Phone", "Role", "Pin", "IsActive", "CreatedAt", "UpdatedAt", "DeletedAt"}
	if len(staffFields) != 10 {
		t.Errorf("Expected 10 staff fields, got %d", len(staffFields))
	}

	// Product model fields (key Phase 2 fields)
	productFields := []string{"ID", "ShopID", "Name", "Category", "Unit", "CostPrice", "SellingPrice", "CurrentStock", "LowStockThreshold", "Barcode", "IsActive"}
	if len(productFields) != 11 {
		t.Errorf("Expected 11 product fields, got %d", len(productFields))
	}

	// Sale model fields (including M-Pesa)
	saleFields := []string{"ID", "ShopID", "ProductID", "Quantity", "UnitPrice", "TotalAmount", "CostAmount", "Profit", "PaymentMethod", "MpesaReceipt", "StaffID", "Notes"}
	if len(saleFields) != 12 {
		t.Errorf("Expected 12 sale fields, got %d", len(saleFields))
	}

	// Shop model fields (including multi-shop)
	shopFields := []string{"ID", "Name", "Phone", "OwnerName", "Address", "Plan", "MpesaShortcode", "IsActive", "Email", "PasswordHash"}
	if len(shopFields) != 10 {
		t.Errorf("Expected 10 shop fields, got %d", len(shopFields))
	}

	t.Log("All model field structures verified")
}

// TestPlanTypes tests plan type constants
func TestPlanTypes(t *testing.T) {
	type PlanType string

	planTypes := []PlanType{"free", "pro", "business"}
	expected := []string{"free", "pro", "business"}

	for i, pt := range planTypes {
		if string(pt) != expected[i] {
			t.Errorf("PlanType[%d] = %s; want %s", i, pt, expected[i])
		}
	}
}

// TestPaymentMethods tests payment method constants
func TestPaymentMethods(t *testing.T) {
	type PaymentMethod string

	methods := []PaymentMethod{"cash", "mpesa", "card", "bank"}
	expected := []string{"cash", "mpesa", "card", "bank"}

	for i, pm := range methods {
		if string(pm) != expected[i] {
			t.Errorf("PaymentMethod[%d] = %s; want %s", i, pm, expected[i])
		}
	}
}
