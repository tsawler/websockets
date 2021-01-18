package handlers

import (
	"fmt"
	"github.com/CloudyKit/jet/v6"
	"log"
	"net/http"
)

var views = jet.NewSet(
	jet.NewOSFileSystemLoader("./html"),
	jet.InDevelopmentMode(),
)

// Home displays the home page with some sample data
func Home(w http.ResponseWriter, r *http.Request) {
	data := make(jet.VarMap)

	data.Set("user_id", 1)

	data.Set("first", "Trevor")

	dow := []string{
		"Sunday",
		"Monday",
		"Tuesday",
		"Wednesday",
		"Thursday",
		"Friday",
		"Saturday",
	}
	data.Set("dow", dow)

	err := renderPage(w, "home.jet", data)
	if err != nil {
		_, _ = fmt.Fprint(w, "Error executing template:", err)
	}

}

// SendData displays the page for sending data via websockets
func SendData(w http.ResponseWriter, r *http.Request) {
	err := renderPage(w, "send.jet", nil)
	if err != nil {
		_, _ = fmt.Fprint(w, "Error executing template:", err)
	}
}

// renderPage renders the page using Jet templates
func renderPage(w http.ResponseWriter, tmpl string, data jet.VarMap) error {
	view, err := views.GetTemplate(tmpl)
	if err != nil {
		log.Println("Unexpected template err:", err.Error())
		return err
	}

	err = view.Execute(w, data, nil)
	if err != nil {
		log.Println("Error executing template:", err)
		return err
	}
	return nil
}
