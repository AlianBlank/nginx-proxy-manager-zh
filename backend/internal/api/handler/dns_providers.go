package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	c "npm/internal/api/context"
	h "npm/internal/api/http"
	"npm/internal/api/middleware"
	"npm/internal/dnsproviders"
	"npm/internal/entity/dnsprovider"
	"npm/internal/errors"

	"gorm.io/gorm"
)

// GetDNSProviders will return a list of DNS Providers
// Route: GET /dns-providers
func GetDNSProviders() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		pageInfo, err := getPageInfoFromRequest(r)
		if err != nil {
			h.ResultErrorJSON(w, r, http.StatusBadRequest, err.Error(), nil)
			return
		}

		items, err := dnsprovider.List(pageInfo, middleware.GetFiltersFromContext(r))
		if err != nil {
			h.ResultErrorJSON(w, r, http.StatusBadRequest, err.Error(), nil)
		} else {
			h.ResultResponseJSON(w, r, http.StatusOK, items)
		}
	}
}

// GetDNSProvider will return a single DNS Provider
// Route: GET /dns-providers/{providerID}
func GetDNSProvider() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		var providerID uint
		if providerID, err = getURLParamInt(r, "providerID"); err != nil {
			h.ResultErrorJSON(w, r, http.StatusBadRequest, err.Error(), nil)
			return
		}

		item, err := dnsprovider.GetByID(providerID)
		switch err {
		case gorm.ErrRecordNotFound:
			h.NotFound(w, r)
		case nil:
			h.ResultResponseJSON(w, r, http.StatusOK, item)
		default:
			h.ResultErrorJSON(w, r, http.StatusBadRequest, err.Error(), nil)
		}
	}
}

// CreateDNSProvider will create a DNS Provider
// Route: POST /dns-providers
func CreateDNSProvider() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, _ := r.Context().Value(c.BodyCtxKey).([]byte)

		var newItem dnsprovider.Model
		err := json.Unmarshal(bodyBytes, &newItem)
		if err != nil {
			h.ResultErrorJSON(w, r, http.StatusBadRequest, h.ErrInvalidPayload.Error(), nil)
			return
		}

		// Get userID from token
		userID, _ := r.Context().Value(c.UserIDCtxKey).(uint)
		newItem.UserID = userID

		if err = newItem.Save(); err != nil {
			h.ResultErrorJSON(w, r, http.StatusBadRequest, fmt.Sprintf("Unable to save DNS Provider: %s", err.Error()), nil)
			return
		}

		h.ResultResponseJSON(w, r, http.StatusOK, newItem)
	}
}

// UpdateDNSProvider updates a provider
// Route: PUT /dns-providers/{providerID}
func UpdateDNSProvider() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		var providerID uint
		if providerID, err = getURLParamInt(r, "providerID"); err != nil {
			h.ResultErrorJSON(w, r, http.StatusBadRequest, err.Error(), nil)
			return
		}

		item, err := dnsprovider.GetByID(providerID)
		switch err {
		case gorm.ErrRecordNotFound:
			h.NotFound(w, r)
		case nil:
			bodyBytes, _ := r.Context().Value(c.BodyCtxKey).([]byte)
			err := json.Unmarshal(bodyBytes, &item)
			if err != nil {
				h.ResultErrorJSON(w, r, http.StatusBadRequest, h.ErrInvalidPayload.Error(), nil)
				return
			}

			if err = item.Save(); err != nil {
				h.ResultErrorJSON(w, r, http.StatusBadRequest, err.Error(), nil)
				return
			}

			h.ResultResponseJSON(w, r, http.StatusOK, item)
		default:
			h.ResultErrorJSON(w, r, http.StatusBadRequest, err.Error(), nil)
		}
	}
}

// DeleteDNSProvider removes a provider
// Route: DELETE /dns-providers/{providerID}
func DeleteDNSProvider() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		var providerID uint
		if providerID, err = getURLParamInt(r, "providerID"); err != nil {
			h.ResultErrorJSON(w, r, http.StatusBadRequest, err.Error(), nil)
			return
		}

		item, err := dnsprovider.GetByID(providerID)
		switch err {
		case gorm.ErrRecordNotFound:
			h.NotFound(w, r)
		case nil:
			h.ResultResponseJSON(w, r, http.StatusOK, item.Delete())
		default:
			h.ResultErrorJSON(w, r, http.StatusBadRequest, err.Error(), nil)
		}
	}
}

// GetAcmeshProviders will return a list of acme.sh providers
// Route: GET /dns-providers/acmesh
func GetAcmeshProviders() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		h.ResultResponseJSON(w, r, http.StatusOK, dnsproviders.List())
	}
}

// GetAcmeshProvider will return a single acme.sh provider
// Route: GET /dns-providers/acmesh/{acmeshID}
func GetAcmeshProvider() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var acmeshID string
		var err error
		if acmeshID, err = getURLParamString(r, "acmeshID"); err != nil {
			h.ResultErrorJSON(w, r, http.StatusBadRequest, err.Error(), nil)
			return
		}

		item, getErr := dnsproviders.Get(acmeshID)
		switch getErr {
		case errors.ErrProviderNotFound:
			h.NotFound(w, r)
		case nil:
			h.ResultResponseJSON(w, r, http.StatusOK, item)
		default:
			h.ResultErrorJSON(w, r, http.StatusBadRequest, getErr.Error(), nil)
		}
	}
}
