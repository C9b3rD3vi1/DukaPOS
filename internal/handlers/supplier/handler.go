package supplier

import (
	"strconv"

	"github.com/C9b3rD3vi1/DukaPOS/internal/models"
	"github.com/C9b3rD3vi1/DukaPOS/internal/repository"
	"github.com/gofiber/fiber/v2"
)

// Handler handles supplier and order HTTP requests
type Handler struct {
	supplierRepo *repository.SupplierRepository
	orderRepo    *repository.OrderRepository
	productRepo  *repository.ProductRepository
}

// getShopID returns shop_id from JWT token (uint) or URL params (string)
func getShopID(c *fiber.Ctx) (uint, error) {
	if sid, ok := c.Locals("shop_id").(uint); ok && sid > 0 {
		return sid, nil
	}
	if sid, ok := c.Locals("shop_id").(string); ok && sid != "" {
		id, err := strconv.ParseUint(sid, 10, 32)
		return uint(id), err
	}
	return 0, fiber.NewError(400, "invalid shop id")
}

// New creates a new supplier handler
func New(supplierRepo *repository.SupplierRepository, orderRepo *repository.OrderRepository, productRepo *repository.ProductRepository) *Handler {
	return &Handler{
		supplierRepo: supplierRepo,
		orderRepo:    orderRepo,
		productRepo:  productRepo,
	}
}

// ListSuppliers GET /suppliers - List all suppliers
func (h *Handler) ListSuppliers(c *fiber.Ctx) error {
	shopID, err := getShopID(c)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid shop id"})
	}

	suppliers, err := h.supplierRepo.GetByShopID(shopID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(suppliers)
}

// CreateSupplier POST /suppliers - Create a new supplier
func (h *Handler) CreateSupplier(c *fiber.Ctx) error {
	shopID, err := getShopID(c)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid shop id"})
	}

	var supplier models.Supplier
	if err := c.BodyParser(&supplier); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}

	supplier.ShopID = shopID
	if err := h.supplierRepo.Create(&supplier); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(supplier)
}

// GetSupplier GET /suppliers/:id - Get a supplier
func (h *Handler) GetSupplier(c *fiber.Ctx) error {
	shopID, err := getShopID(c)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid shop id"})
	}

	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid supplier id"})
	}

	supplier, err := h.supplierRepo.GetByID(uint(id))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "supplier not found"})
	}

	if supplier.ShopID != shopID {
		return c.Status(403).JSON(fiber.Map{"error": "not authorized"})
	}

	return c.JSON(supplier)
}

// UpdateSupplier PUT /suppliers/:id - Update a supplier
func (h *Handler) UpdateSupplier(c *fiber.Ctx) error {
	shopID, err := getShopID(c)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid shop id"})
	}

	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid supplier id"})
	}

	supplier, err := h.supplierRepo.GetByID(uint(id))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "supplier not found"})
	}

	if supplier.ShopID != shopID {
		return c.Status(403).JSON(fiber.Map{"error": "not authorized"})
	}

	var updates map[string]interface{}
	if err := c.BodyParser(&updates); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}

	if name, ok := updates["name"].(string); ok {
		supplier.Name = name
	}
	if phone, ok := updates["phone"].(string); ok {
		supplier.Phone = phone
	}
	if email, ok := updates["email"].(string); ok {
		supplier.Email = email
	}
	if address, ok := updates["address"].(string); ok {
		supplier.Address = address
	}

	if err := h.supplierRepo.Update(supplier); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(supplier)
}

// DeleteSupplier DELETE /suppliers/:id - Delete a supplier
func (h *Handler) DeleteSupplier(c *fiber.Ctx) error {
	shopID, err := getShopID(c)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid shop id"})
	}

	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid supplier id"})
	}

	supplier, err := h.supplierRepo.GetByID(uint(id))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "supplier not found"})
	}

	if supplier.ShopID != shopID {
		return c.Status(403).JSON(fiber.Map{"error": "not authorized"})
	}

	if err := h.supplierRepo.Delete(uint(id)); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.SendStatus(204)
}

// ListOrders GET /orders - List all orders
func (h *Handler) ListOrders(c *fiber.Ctx) error {
	shopID, err := getShopID(c)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid shop id"})
	}

	status := c.Query("status")
	var orders []models.Order
	if status != "" {
		orders, err = h.orderRepo.GetByStatus(shopID, status)
	} else {
		orders, err = h.orderRepo.GetByShopID(shopID)
	}
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(orders)
}

// CreateOrder POST /orders - Create a new order
func (h *Handler) CreateOrder(c *fiber.Ctx) error {
	shopID, err := getShopID(c)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid shop id"})
	}

	type OrderRequest struct {
		SupplierID uint               `json:"supplier_id"`
		Status     string             `json:"status"`
		Notes      string             `json:"notes"`
		Items      []models.OrderItem `json:"items"`
	}

	var req OrderRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}

	// Validate supplier
	supplier, err := h.supplierRepo.GetByID(req.SupplierID)
	if err != nil || supplier.ShopID != shopID {
		return c.Status(400).JSON(fiber.Map{"error": "invalid supplier"})
	}

	// Calculate total
	var total float64
	for _, item := range req.Items {
		total += item.TotalCost
	}

	order := &models.Order{
		ShopID:      shopID,
		SupplierID:  req.SupplierID,
		Status:      req.Status,
		TotalAmount: total,
		Notes:       req.Notes,
	}

	if err := h.orderRepo.Create(order); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Create order items
	for i := range req.Items {
		req.Items[i].OrderID = order.ID
		req.Items[i].TotalCost = float64(req.Items[i].Quantity) * req.Items[i].UnitCost
		if err := h.orderRepo.CreateItem(&req.Items[i]); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
	}

	// Reload order with items
	order, _ = h.orderRepo.GetByID(order.ID)
	return c.Status(201).JSON(order)
}

// GetOrder GET /orders/:id - Get an order
func (h *Handler) GetOrder(c *fiber.Ctx) error {
	shopID, err := getShopID(c)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid shop id", "details": err.Error()})
	}

	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid order id"})
	}

	order, err := h.orderRepo.GetByID(uint(id))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "order not found", "order_id": id, "shop_id": shopID, "details": err.Error()})
	}

	if order.ShopID != shopID {
		return c.Status(403).JSON(fiber.Map{"error": "not authorized", "order_shop": order.ShopID, "user_shop": shopID})
	}

	return c.JSON(order)
}

// UpdateOrderStatus PUT /orders/:id/status - Update order status
func (h *Handler) UpdateOrderStatus(c *fiber.Ctx) error {
	shopID, err := getShopID(c)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid shop id"})
	}

	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid order id"})
	}

	order, err := h.orderRepo.GetByID(uint(id))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "order not found"})
	}

	if order.ShopID != shopID {
		return c.Status(403).JSON(fiber.Map{"error": "not authorized"})
	}

	type StatusUpdate struct {
		Status string `json:"status"`
	}

	var update StatusUpdate
	if err := c.BodyParser(&update); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}

	validStatuses := map[string]bool{
		"pending":   true,
		"confirmed": true,
		"shipped":   true,
		"delivered": true,
		"cancelled": true,
	}

	if !validStatuses[update.Status] {
		return c.Status(400).JSON(fiber.Map{"error": "invalid status"})
	}

	order.Status = update.Status
	if err := h.orderRepo.Update(order); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(order)
}

// DeleteOrder DELETE /orders/:id - Delete an order
func (h *Handler) DeleteOrder(c *fiber.Ctx) error {
	shopID, err := getShopID(c)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid shop id"})
	}

	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid order id"})
	}

	order, err := h.orderRepo.GetByID(uint(id))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "order not found"})
	}

	if order.ShopID != shopID {
		return c.Status(403).JSON(fiber.Map{"error": "not authorized"})
	}

	if err := h.orderRepo.DeleteItems(uint(id)); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	if err := h.orderRepo.Delete(uint(id)); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.SendStatus(204)
}
