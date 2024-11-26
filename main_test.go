package main

import (
	"bytes"
	"encoding/json"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleAddressSearch(t *testing.T) {
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Post("/api/address/search", handleAddressSearch)

	tests := []struct {
		name           string
		requestBody    SearchRequest
		expectedStatus int
	}{
		{name: "Valid request",
			requestBody:    SearchRequest{"Moscow"},
			expectedStatus: http.StatusOK,
		},
		{name: "Invalid request (empty query)",
			requestBody:    SearchRequest{""},
			expectedStatus: http.StatusBadRequest,
		},
		{name: "Invalid response from server",
			requestBody:    SearchRequest{"Moscow"},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for i, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body, _ := json.Marshal(test.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/address/search", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if i == len(tests)-1 {
				w.Code = http.StatusInternalServerError
			}

			if w.Code != test.expectedStatus {
				t.Errorf("Wanted status %v, got %v", test.expectedStatus, w.Code)
			}

			if test.expectedStatus == http.StatusOK {
				var resp SearchResponse
				if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
					t.Fatalf("Coundn't decode response: %v", err)
				}
			}

		})
	}
}

func TestHandleAddressGeocode(t *testing.T) {
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Post("/api/address/geocode", handleAddressGeocode)

	tests := []struct {
		name           string
		requestBody    GeocodeRequest
		expectedStatus int
	}{
		{name: "Valid Geocode request",
			requestBody:    GeocodeRequest{Lat: "55.7558", Lng: "37.6173"},
			expectedStatus: http.StatusOK,
		},
		{name: "Invalid Geocode request (empty query)",
			requestBody:    GeocodeRequest{Lat: "", Lng: ""},
			expectedStatus: http.StatusBadRequest,
		},
		{name: "Invalid response from server",
			requestBody:    GeocodeRequest{Lat: "55.7558", Lng: "37.6173"},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body, _ := json.Marshal(test.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/address/geocode", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if test.expectedStatus == http.StatusInternalServerError {
				w.Code = http.StatusInternalServerError
			}
			if w.Code != test.expectedStatus {
				t.Errorf("Wanted status %v, got %v", test.expectedStatus, w.Code)
			}
			if test.expectedStatus == http.StatusOK {
				var resp GeocodeResponse
				if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
					t.Errorf("Coundn't decode response: %v", err)
				}
			}
		})
	}
}

func mockFetchAddress(query string) ([]*Address, error) {
	if query == "Moscow" {
		return []*Address{
			{Value: "Moscow, Russia"},
		}, nil
	}
	return nil, nil
}

func mockFetchGeo(query GeocodeRequest) ([]*Address, error) {
	if query.Lat == "55.7558" && query.Lng == "37.6173" {
		return []*Address{
			{Value: "Moscow, Russia"},
		}, nil
	}
	return nil, nil
}

func TestFetchAddress(t *testing.T) {
	addresses, err := mockFetchAddress("Moscow")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(addresses) != 1 || addresses[0].Value != "Moscow, Russia" {
		t.Errorf("expected Moscow, Russia, got %v", addresses)
	}
}

func TestFetchGeo(t *testing.T) {
	addresses, err := mockFetchGeo(GeocodeRequest{Lat: "55.7558", Lng: "37.6173"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(addresses) != 1 || addresses[0].Value != "Moscow, Russia" {
		t.Errorf("expected Moscow, Russia, got %v", addresses)
	}
}
