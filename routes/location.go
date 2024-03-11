package routes

import (
	"encoding/json"
	"io"
	"net/http"
	"os"

	"github.com/kataras/iris/v12"
)

func Autocomplete(ctx iris.Context) {
	limit := "10"
	location := ctx.URLParam("location")
	limitQuery := ctx.URLParam("limit")
	if limitQuery != "" {
		limit = limitQuery
	}

	apiKey := os.Getenv("LOCATION_TOKEN")
	countryCodes := os.Getenv("LOCATION_COUNTRY_CODES")
	url :=
		"https://api.locationiq.com/v1/autocomplete.php?key=" + apiKey + "&q=" + location + "&limit=" + limit + "&countrycodes=" + countryCodes

	fetchLocations(url, ctx)
}

func Search(ctx iris.Context) {
	location := ctx.URLParam("location")

	apiKey := os.Getenv("LOCATION_TOKEN")
	url :=
		"https://api.locationiq.com/v1/search.php?key=" + apiKey + "&q=" + location + "&format=json&dedupe=1&addressdetails=1&matchquality=1&normalizeaddress=1&normalizecity=1&countrycodes=bd"

	fetchLocations(url, ctx)
}

func fetchLocations(url string, ctx iris.Context) {
	var objMap []map[string]interface{}

	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	res, locationErr := client.Do(req)
	if locationErr != nil {
		ctx.JSON(objMap)
		return
	}

	defer res.Body.Close()

	body, bodyErr := io.ReadAll(res.Body)
	if bodyErr != nil {
		ctx.JSON(objMap)
		return
	}

	jsonErr := json.Unmarshal(body, &objMap)
	if jsonErr != nil {
		ctx.JSON(objMap)
		return
	}

	ctx.JSON(objMap)
}
