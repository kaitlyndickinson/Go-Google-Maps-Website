package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path"
	"strconv"
	"time"

	"github.com/kelvins/geocoder"
	"googlemaps.github.io/maps"
)

type Address struct {
	StreetNumber string
	StreetName   string
	City         string
	State        string
	Country      string
}

// Struct that stores the directions
type DirectionsRequest struct {
	Origin      string
	Destination string
}

// Globals
var saveFlag = false
var latLongFlag = false
var address = geocoder.Address{}
var maplocation string
var zoomlevel string
var maptype string
var mapsize string
var startString = "<img src='https://maps.googleapis.com/maps/api/staticmap?center="
var zoomString = "&zoom="
var mapTypeString = "&maptype="
var mapSizeString = "&size="

func main() {

	// === ROUTE HANDLERS ===
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("styles")))) // For serving CSS
	http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("src"))))            // For serving html

	http.HandleFunc("/saveaddress", saveaddress)
	http.HandleFunc("/golatlong.html", latLong)

	// === Go Maps ===
	http.HandleFunc("/savemap", mapsSaveLocation)
	http.HandleFunc("/gomaps.html", goMaps)

	// === Go Directions ===
	http.HandleFunc("/savedirections", DirectionsSave)
	http.HandleFunc("/godirections.html", Directions)

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func saveaddress(w http.ResponseWriter, r *http.Request) {
	address.Number, _ = strconv.Atoi(r.FormValue("streetnumber"))
	address.Street = r.FormValue("streetname")
	address.City = r.FormValue("city")
	address.State = r.FormValue("state")
	address.Country = r.FormValue("country")
	saveFlag = true

	if latLongFlag {
		latLong(w, r)
		latLongFlag = false
	} else {
		timezone(w, r)
	}
}

func latLong(w http.ResponseWriter, r *http.Request) {
	latLongFlag = true

	fp := path.Join("src", "golatlong.html")
	tmpl, err := template.ParseFiles(fp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	if saveFlag {
		geocoder.ApiKey = "AIzaSyDnmYy2YrGXlYqQHOPHXyxRLOEDoaPFzok"

		//convert location to lat/long
		location, err := geocoder.Geocoding(address)
		if err != nil {
			fmt.Println("Could not get location: ", err)
		} else {
			fmt.Fprintln(w, "Latitude: ", location.Latitude)
			fmt.Fprintln(w, "Longitude: ", location.Longitude)
		}
	}
}

// === Go Maps Functionality ===
// Saves the form values from gomaps.html
func mapsSaveLocation(w http.ResponseWriter, r *http.Request) {
	maplocation = r.FormValue("maplocation")
	zoomlevel = r.FormValue("zoomlevel")
	mapsize = r.FormValue("mapsize")
	maptype = r.FormValue("maptype")
	http.Redirect(w, r, "/gomaps.html", http.StatusSeeOther)
}

// Takes the user input to generate one final string and print a static map to gomaps.html
func goMaps(w http.ResponseWriter, r *http.Request) {
	fp := path.Join("src", "gomaps.html")
	tmpl, err := template.ParseFiles(fp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	// User has provided input in all fields
	if maplocation != "" && zoomlevel != "" && mapsize != "" && maptype != "" {
		var endString = "&key=AIzaSyDnmYy2YrGXlYqQHOPHXyxRLOEDoaPFzok'>"

		// Create one final string to build the image and parameters needed to connect to Google Maps API
		var finalString = startString + maplocation + zoomString + zoomlevel + mapSizeString + mapsize + mapTypeString + maptype + endString

		// Display on gomaps.html page
		fmt.Fprintf(w, "%s", finalString)

		// User has provided input for first 3 fields
	} else if maplocation != "" && zoomlevel != "" && mapsize != "" {
		var endString = "&maptype=roadmap&key=AIzaSyDnmYy2YrGXlYqQHOPHXyxRLOEDoaPFzok'>"
		var finalString = startString + maplocation + zoomString + zoomlevel + mapSizeString + mapsize + endString

		fmt.Fprintf(w, "%s", finalString)

		// User has provided input for first 2 fields
	} else if maplocation != "" && zoomlevel != "" {
		var endString = "&size=600x300&maptype=roadmap&key=AIzaSyDnmYy2YrGXlYqQHOPHXyxRLOEDoaPFzok'>"
		var finalString = startString + maplocation + zoomString + zoomlevel + endString

		fmt.Fprintf(w, "%s", finalString)

		// User has only provided geographic location
	} else if maplocation != "" {
		var endString = "&zoom=15&size=600x300&maptype=roadmap&key=AIzaSyDnmYy2YrGXlYqQHOPHXyxRLOEDoaPFzok'>"
		var finalString = startString + maplocation + endString

		fmt.Fprintf(w, "%s", finalString)
	}
}

// === Go Directions Functionality ===
// So that we can store the data you get in DirectionsSave
var r DirectionsRequest

// So you can get the data out of godirections.html
func DirectionsSave(w http.ResponseWriter, d *http.Request) {
	r.Origin = d.FormValue("startLocation")
	r.Destination = d.FormValue("endLocation")
	Directions(w, d)
}

// go thought to get the string that is needed to make the url
func Directions(w http.ResponseWriter, d *http.Request) {
	fp := path.Join("src", "godirections.html")
	tmpl, err := template.ParseFiles(fp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	// Starting of the string for the URL
	var startString_d = "https://maps.googleapis.com/maps/api/directions/json?origin="
	// Making the URL
	startString_d += r.Origin + "MA&destination=" + r.Destination + "&key=AIzaSyDnmYy2YrGXlYqQHOPHXyxRLOEDoaPFzok"

	// This prints out the URL on the termial
	println(startString_d)
}

func timezone(w http.ResponseWriter, r *http.Request) {
	latLongFlag = false

	fp := path.Join("src", "timezone.html")
	tmpl, err := template.ParseFiles(fp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	geocoder.ApiKey = "AIzaSyDnmYy2YrGXlYqQHOPHXyxRLOEDoaPFzok"
	c, err := maps.NewClient(maps.WithAPIKey("AIzaSyDnmYy2YrGXlYqQHOPHXyxRLOEDoaPFzok"))
	if err != nil {
		log.Fatal(err)
	}

	//convert location to lat/long
	location, err := geocoder.Geocoding(address)
	if err != nil {
		fmt.Println("Could not get location: ", err)
	}

	// convert to LatLng slice
	ls := make([]maps.LatLng, 1)
	ls[0].Lat = location.Latitude
	ls[0].Lng = location.Longitude

	// get elevation request
	t := &maps.ElevationRequest{
		Locations: ls,
	}

	ctx := context.TODO()
	data, err := c.Elevation(ctx, t)
	if err != nil {
		fmt.Println("Could not get elevation: ", err)
	}

	tr := &maps.TimezoneRequest{
		Location:  &ls[0],
		Timestamp: time.Now(),
		Language:  "English",
	}

	tz, err := c.Timezone(ctx, tr)
	if err != nil {
		fmt.Println("Could not get timezone: ", err)
	}

	addresses, err := geocoder.GeocodingReverse(location)
	if err != nil {
		fmt.Println("Could not get the address: ", err)
	} else {
		address = addresses[0]
		fmt.Fprintln(w, "Timezone ID: ", tz.TimeZoneID)
		fmt.Fprintf(w, "%s", "<br><br>")
		fmt.Fprintln(w, "Timezone: ", tz.TimeZoneName)
		fmt.Fprintf(w, "%s", "<br><br>")
		fmt.Fprintln(w, "Elevation(in feet): ", int(data[0].Elevation*3.2808))
		fmt.Fprintf(w, "%s", "<br><br>")
		fmt.Fprintln(w, "District: ", address.District)
		fmt.Fprintf(w, "%s", "<br><br>")
		fmt.Fprintln(w, "County: ", address.County)
		fmt.Fprintf(w, "%s", "<br><br>")
		fmt.Fprintln(w, "Zip Code: ", address.PostalCode)
		fmt.Fprintf(w, "%s", "<br><br>")
		fmt.Fprintln(w, "Neighborhood: ", address.Neighborhood)
		fmt.Fprintf(w, "%s", "<br><br>")
		fmt.Fprintln(w, "Type: ", address.Types)
	}
}
