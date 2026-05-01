package http

import (
	"net/http"
	"strconv"

	appAddress "kafka-order-demo/backend/internal/application/address"

	"github.com/gin-gonic/gin"
)

type AddressHandler struct {
	addressService *appAddress.Service
}

func NewAddressHandler(addressService *appAddress.Service) *AddressHandler {
	return &AddressHandler{addressService: addressService}
}

func (h *AddressHandler) GetAddresses(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		errorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	addresses, err := h.addressService.GetAddresses(userID.(uint))
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	successResponse(c, http.StatusOK, "Get addresses successfully", toAddressHTTPResponses(addresses))
}

func (h *AddressHandler) CreateAddress(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		errorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req addressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	address, err := h.addressService.CreateAddress(toAddressInput(req, userID.(uint)))
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	successResponse(c, http.StatusCreated, "Create address successfully", toAddressHTTPResponse(*address))
}

func (h *AddressHandler) UpdateAddress(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		errorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	id, err := parseAddressID(c)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "Invalid address ID")
		return
	}

	var req addressUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	address, err := h.addressService.UpdateAddress(uint(id), userID.(uint), toAddressUpdateInput(req))
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	successResponse(c, http.StatusOK, "Update address successfully", toAddressHTTPResponse(*address))
}

func (h *AddressHandler) DeleteAddress(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		errorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	id, err := parseAddressID(c)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "Invalid address ID")
		return
	}

	if err := h.addressService.DeleteAddress(uint(id), userID.(uint)); err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	successResponse(c, http.StatusOK, "Delete address successfully", nil)
}

func (h *AddressHandler) SetDefaultAddress(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		errorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	id, err := parseAddressID(c)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "Invalid address ID")
		return
	}

	if err := h.addressService.SetDefaultAddress(uint(id), userID.(uint)); err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	successResponse(c, http.StatusOK, "Set default address successfully", nil)
}

func parseAddressID(c *gin.Context) (uint64, error) {
	return strconv.ParseUint(c.Param("id"), 10, 32)
}
