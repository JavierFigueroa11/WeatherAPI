package services

import (
	"TPFinal/pkg/domain"
	"TPFinal/pkg/utils"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	myip "github.com/polds/MyIP"
)

var weatherAPIKey string = "b64966af79891ad1f90c85de924bbe10"
var key string = "440d88bc9073b1"

var locations = make(map[string]domain.Locations)

func GetIP() (string, error) {
	ip, err := myip.GetMyIP()
	if err != nil {
		err = errors.New("Error al obtener su IP")
	}
	return ip[:len(ip)-2], err
}

func GetLocation() (domain.Location, error) {
	ip, err := GetIP()
	var location domain.Location
	if err != nil {
		err = errors.New("Error al obtener su IP")
		return location, err
	}
	resp, err2 := http.Get("http://ipvigilante.com/" + ip)
	data, err3 := ioutil.ReadAll(resp.Body)
	if err2 != nil || err3 != nil {
		err = errors.New("Error al obtener su ubicación")
		return location, err
	}
	json.Unmarshal(data, &location)
	if location.Status != "success" {
		err = errors.New("Error al obtener su ubicación")
		return location, err
	}
	return location, nil
}

func GetLocations() ([]domain.Locations, error) {
	client := utils.GetClient()
	locations, err := client.GetLocations()
	if err != nil {
		return locations, err
	}
	return locations, nil
}

func GetLocationID(id string) (domain.Locations, error) {
	client := utils.GetClient()
	location, err := client.GetLocationId(id)
	if err != nil {
		err = errors.New("Error al buscar la ubicación")
	}
	return location, err
}

func cleanQuery(params []string) string {
	var query string
	for _, param := range params {
		query += url.QueryEscape(param) + ","
	}
	return query
}

func CreateLocation(city, state, country string) (domain.Search, error) {
	var search []domain.Search
	cleaned := []string{city, state, country}
	query := cleanQuery(cleaned)
	url := "https://us1.locationiq.com/v1/search.php?key=" + key + "&q=" + query + "&format=json"
	resp, err := http.Get(url)
	if err != nil {
		err = errors.New("Error al buscar la ubicación")
		return search[0], err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			err = errors.New("Error al crear la ubicación")
			return search[0], err
		}
		json.Unmarshal(data, &search)
		_, err = GetLocationID(search[0].Id)
		if err == nil {
			return search[0], errors.New("la ubicación ya se encuentra en la lista")
		}
		aux := domain.Locations{
			Id:   search[0].Id,
			Name: search[0].Name,
			Lat:  search[0].Lat,
			Lon:  search[0].Lon,
		}
		locations[search[0].Id] = aux
		client := utils.GetClient()
		client.SaveLocation(aux)
		return search[0], nil
	}
	return search[0], errors.New("Error al buscar la ubicación")
}

func DeleteLocation(id string) error {
	_, err := GetLocationID(id)
	if err != nil {
		return errors.New("Error al borrar la ubicación")
	}
	client := utils.GetClient()
	err = client.DeleteLocation(id)
	if err != nil {
		return err
	}
	return errors.New("Error al borrar la ubicación")
}

func UpdateLocation(id, name, lat, lon string) (domain.Locations, error) {
	location, err := GetLocationID(id)
	if err != nil {
		var locAux domain.Locations
		return locAux, errors.New("Error al actualizar la ubicación")
	}
	client := utils.GetClient()
	new := domain.Locations{
		Id:   location.Id,
		Name: name,
		Lat:  lat,
		Lon:  lon,
	}
	newLocation, err := client.SaveLocation(new)
	if err != nil {
		return newLocation, errors.New("Couldn't Update Location")
	}
	return newLocation, nil
}

func GetWeather() (domain.Weather, error) {
	location, err := GetLocation()
	var weather domain.Weather
	if err != nil {
		err = errors.New("Error al obtener su ubicación")
		return weather, err
	}
	var url string = "http://api.openweathermap.org/data/2.5/weather?"
	url += "lat=" + location.Data.Latitud
	url += "&lon=" + location.Data.Longitud
	url += "&units=metric"
	url += "&appid=" + weatherAPIKey
	resp, err2 := http.Get(url)
	data, err3 := ioutil.ReadAll(resp.Body)
	if err2 != nil || err3 != nil {
		err = errors.New("Error al obtener el clima")
		return weather, err
	}
	json.Unmarshal(data, &weather)
	return weather, nil
}
