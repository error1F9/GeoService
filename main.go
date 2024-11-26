package main

import (
	_ "GeoSrvice/docs"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	dadata "github.com/ekomobile/dadata/v2"
	"github.com/ekomobile/dadata/v2/api/suggest"
	"github.com/ekomobile/dadata/v2/client"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
	"net/http"
)

type SearchRequest struct {
	Query string `json:"query"`
}
type SearchResponse struct {
	Addresses []*Address `json:"addresses"`
}

type GeocodeRequest struct {
	Lat string `json:"lat"`
	Lng string `json:"lng"`
}
type GeocodeResponse struct {
	Addresses []*Address `json:"suggestions"`
}

type Address struct {
	Value string `json:"value"`
}

const dataURL = "http://suggestions.dadata.ru"

// @title Todo App API
// @version 1.0
// @description This is a sample server Petstore server.
// @termsOfService http://swagger.io/terms/

// @host localhost:8080
// @BasePath /

func main() {
	r := chi.NewRouter()

	r.Use(middleware.Logger)

	r.Post("/api/address/search", handleAddressSearch)
	r.Post("/api/address/geocode", handleAddressGeocode)

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("swagger.json"), //The url pointing to API definition
	))

	http.ListenAndServe(":8080", r)
}

// @Summary receive Address by GeoCode
// @Tags GeoCode
// @Description Handle Address by GeoCode
// @ID geo
// @Accept json
// @Produce json
// @Param input body GeocodeResponse true "Handle Address by GeoCode"
// @Success 200 {object} SearchResponse
// @Success 400 {object} error
// @Success 500 {object} error
// @Router /api/address/geocode [post]
func handleAddressGeocode(w http.ResponseWriter, r *http.Request) {
	var geocodeReq GeocodeRequest
	if err := json.NewDecoder(r.Body).Decode(&geocodeReq); err != nil || geocodeReq.Lng == "" || geocodeReq.Lat == "" {
		http.Error(w, "Empty Query", http.StatusBadRequest)
		return
	}

	geo, err := fetchGeo(geocodeReq)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := SearchResponse{geo}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)

}

func fetchGeo(query GeocodeRequest) ([]*Address, error) {
	url := fmt.Sprintf("%s/suggestions/api/4_1/rs/geolocate/address", dataURL)
	body := map[string]string{"lat": query.Lat, "lon": query.Lng}
	return sendDadataGeoRequest(url, body)
}

func sendDadataGeoRequest(url string, body map[string]string) ([]*Address, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Token 90a5dd26d0ba58ea94f25f085aa113ad67f2af27")
	//req.Header.Set("X-Secret", "eb3066ce98823788c54dafb9e5e66d87a3c92d9d")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %s", resp.Status)
	}

	var geoResponse GeocodeResponse

	if err := json.NewDecoder(resp.Body).Decode(&geoResponse); err != nil {
		return nil, err
	}

	return geoResponse.Addresses, nil
}

func handleAddressSearch(w http.ResponseWriter, r *http.Request) {
	var req SearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Query == "" {
		http.Error(w, "Empty Query", http.StatusBadRequest)
		return
	}

	addresses, err := fetchAddress(req.Query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := SearchResponse{addresses}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
	//fmt.Println(resp.Addresses[0].Value)

}

func fetchAddress(query string) ([]*Address, error) {
	creds := client.Credentials{
		ApiKeyValue:    "90a5dd26d0ba58ea94f25f085aa113ad67f2af27",
		SecretKeyValue: "eb3066ce98823788c54dafb9e5e66d87a3c92d9d",
	}

	api := dadata.NewSuggestApi(client.WithCredentialProvider(&creds))

	params := suggest.RequestParams{
		Query: query,
	}

	suggestions, err := api.Address(context.Background(), &params)
	if err != nil {
		return nil, err
	}

	addresses := make([]*Address, len(suggestions))

	for i, s := range suggestions {
		addresses[i] = &Address{s.Value}
	}

	return addresses, nil
}
