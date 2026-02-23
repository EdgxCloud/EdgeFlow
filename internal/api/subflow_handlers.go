package api

import (
	"github.com/EdgxCloud/EdgeFlow/internal/subflow"
	"github.com/gofiber/fiber/v2"
)

// SetupSubflowRoutes registers subflow API routes
func (h *Handler) SetupSubflowRoutes(app *fiber.App) {
	// Initialize subflow components
	registry := subflow.GlobalRegistry()
	library := subflow.NewLibrary("./data/subflows", registry)
	executor := subflow.NewExecutor(registry, nil) // NodeExecutor will be set later

	// Create subflow handler
	sfh := &subflowHandler{
		registry: registry,
		library:  library,
		executor: executor,
	}

	// Subflow routes
	api := app.Group("/api/subflows")

	// Stats (must be before /:id routes)
	api.Get("/stats", sfh.getStats)

	// Library (must be before /:id routes)
	api.Get("/library", sfh.listLibrary)
	api.Get("/library/categories", sfh.listCategories)
	api.Post("/import", sfh.importSubflow)
	api.Post("/package/export", sfh.exportPackage)
	api.Post("/package/import", sfh.importPackage)

	// Instances (must be before /:id routes)
	api.Get("/instances/:instanceId", sfh.getInstance)
	api.Delete("/instances/:instanceId", sfh.deleteInstance)
	api.Get("/instances/:instanceId/state", sfh.getInstanceState)
	api.Post("/instances/:instanceId/stop", sfh.stopInstance)

	// Definitions
	api.Get("/", sfh.listSubflows)
	api.Post("/", sfh.createSubflow)
	api.Get("/:id", sfh.getSubflow)
	api.Put("/:id", sfh.updateSubflow)
	api.Delete("/:id", sfh.deleteSubflow)
	api.Post("/:id/clone", sfh.cloneSubflow)
	api.Get("/:id/instances", sfh.listInstances)
	api.Get("/:id/export", sfh.exportSubflow)
}

type subflowHandler struct {
	registry *subflow.Registry
	library  *subflow.Library
	executor *subflow.Executor
}

func (h *subflowHandler) listSubflows(c *fiber.Ctx) error {
	definitions := h.registry.ListDefinitions()
	return c.JSON(fiber.Map{
		"subflows": definitions,
		"count":    len(definitions),
	})
}

func (h *subflowHandler) getSubflow(c *fiber.Ctx) error {
	id := c.Params("id")
	def, err := h.registry.GetDefinition(id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(def)
}

func (h *subflowHandler) createSubflow(c *fiber.Ctx) error {
	var def subflow.SubflowDefinition
	if err := c.BodyParser(&def); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if err := h.registry.RegisterDefinition(&def); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(def)
}

func (h *subflowHandler) updateSubflow(c *fiber.Ctx) error {
	id := c.Params("id")

	var def subflow.SubflowDefinition
	if err := c.BodyParser(&def); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if def.ID != id {
		return c.Status(400).JSON(fiber.Map{"error": "ID mismatch"})
	}

	if err := h.registry.UpdateDefinition(&def); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(def)
}

func (h *subflowHandler) deleteSubflow(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.library.Delete(id); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(204)
}

func (h *subflowHandler) cloneSubflow(c *fiber.Ctx) error {
	sourceID := c.Params("id")

	var req struct {
		NewID   string `json:"newId"`
		NewName string `json:"newName"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	clone, err := h.library.Clone(sourceID, req.NewID, req.NewName)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(clone)
}

func (h *subflowHandler) listInstances(c *fiber.Ctx) error {
	id := c.Params("id")
	instances := h.registry.GetInstancesBySubflow(id)
	return c.JSON(fiber.Map{
		"instances": instances,
		"count":     len(instances),
	})
}

func (h *subflowHandler) getInstance(c *fiber.Ctx) error {
	instanceID := c.Params("instanceId")
	instance, err := h.registry.GetInstance(instanceID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(instance)
}

func (h *subflowHandler) deleteInstance(c *fiber.Ctx) error {
	instanceID := c.Params("instanceId")
	h.executor.StopInstance(instanceID)
	h.registry.UnregisterInstance(instanceID)
	return c.SendStatus(204)
}

func (h *subflowHandler) getInstanceState(c *fiber.Ctx) error {
	instanceID := c.Params("instanceId")
	state, err := h.executor.GetInstanceState(instanceID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(state)
}

func (h *subflowHandler) stopInstance(c *fiber.Ctx) error {
	instanceID := c.Params("instanceId")
	if err := h.executor.StopInstance(instanceID); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{
		"status":  "stopped",
		"message": "Instance stopped successfully",
	})
}

func (h *subflowHandler) listLibrary(c *fiber.Ctx) error {
	entries, err := h.library.ListLibrary()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{
		"entries": entries,
		"count":   len(entries),
	})
}

func (h *subflowHandler) listCategories(c *fiber.Ctx) error {
	entries, err := h.library.ListLibrary()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	categories := make(map[string]int)
	for _, entry := range entries {
		categories[entry.Category]++
	}

	return c.JSON(fiber.Map{
		"categories": categories,
	})
}

func (h *subflowHandler) exportSubflow(c *fiber.Ctx) error {
	id := c.Params("id")

	def, err := h.registry.GetDefinition(id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": err.Error()})
	}

	c.Set("Content-Type", "application/json")
	c.Set("Content-Disposition", "attachment; filename="+def.Name+".json")

	data, err := def.ToJSON()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Send(data)
}

func (h *subflowHandler) importSubflow(c *fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Failed to read file"})
	}

	fileData, err := file.Open()
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Failed to open file"})
	}
	defer fileData.Close()

	def, err := h.library.ImportFromReader(fileData)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(def)
}

func (h *subflowHandler) exportPackage(c *fiber.Ctx) error {
	var req struct {
		SubflowIDs []string `json:"subflowIds"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	pkg := subflow.SubflowPackage{
		Version:  "1.0",
		Subflows: make([]*subflow.SubflowDefinition, 0, len(req.SubflowIDs)),
	}

	for _, id := range req.SubflowIDs {
		def, err := h.registry.GetDefinition(id)
		if err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "Subflow not found: " + id})
		}
		pkg.Subflows = append(pkg.Subflows, def)
	}

	c.Set("Content-Type", "application/json")
	c.Set("Content-Disposition", "attachment; filename=subflows-package.json")
	return c.JSON(pkg)
}

func (h *subflowHandler) importPackage(c *fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Failed to read file"})
	}

	fileData, err := file.Open()
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Failed to open file"})
	}
	defer fileData.Close()

	// Read file content
	data := make([]byte, file.Size)
	if _, err := fileData.Read(data); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Failed to read file data"})
	}

	var pkg subflow.SubflowPackage
	if err := c.App().Config().JSONDecoder(data, &pkg); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Failed to parse package"})
	}

	var imported []string
	for _, def := range pkg.Subflows {
		if err := h.registry.RegisterDefinition(def); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Failed to register subflow " + def.ID})
		}
		imported = append(imported, def.ID)
	}

	return c.Status(201).JSON(fiber.Map{
		"imported": imported,
		"count":    len(imported),
	})
}

func (h *subflowHandler) getStats(c *fiber.Ctx) error {
	stats := h.registry.Stats()
	return c.JSON(stats)
}
