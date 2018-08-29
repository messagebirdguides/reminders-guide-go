package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/messagebird/go-rest-api"
	"github.com/messagebird/go-rest-api/lookup"
	"github.com/messagebird/go-rest-api/sms"
)

// Global, because we need to share this with the handler functions
var (
	client *messagebird.Client
)

// Data structures
type booking struct {
	Name        string
	Treatment   string
	Phone       string
	BookingTime *time.Time
	MinDate     string
}

type bookingContainer struct {
	Booking booking
	Message string
}

func main() {
	client = messagebird.New("<enter-your-apikey>")

	// Routes
	http.HandleFunc("/", bbScheduler)

	// Serve
	port := ":8080"
	log.Println("Serving application on", port)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Println(err)
	}
}

// Routes
func bbScheduler(w http.ResponseWriter, r *http.Request) {
	var (
		loc          *time.Location
		reminderDiff time.Duration
		err          error
	)

	// Set locale. Hardcoding this because you're unlikely to set a beauty appointment across timezones.
	// Use list in /usr/local/Cellar/go/1.10.3/libexec/lib/time/zoneinfo.zip
	loc, err = time.LoadLocation("Europe/Amsterdam")
	if err != nil {
		log.Println(err)
	}

	// Set a time.Duration value for message scheduling.
	// Here, we set a 3 hour duration that we will use to subtract from the booking time.
	reminderDiff, err = time.ParseDuration("3h")
	if err != nil {
		log.Println(err)
	}

	// Initialize &booking with only MinDate values so that we can pass "min" value into <input type="date"/>
	BookingEmpty := booking{
		MinDate: time.Now().In(loc).Format("2006-01-02"),
	}

	// Handle form submission
	if r.Method == "POST" {
		r.ParseForm()

		// Convert r.FormValue("date") to time.Time type.
		bookingTime, err := time.ParseInLocation("2006-01-02 15:04", r.FormValue("date")+" "+r.FormValue("time"), loc)
		if err != nil {
			log.Println(err)
		}

		reminderTime := bookingTime.Add(-reminderDiff)

		// Populate ThisBooking with data to pass back into form.
		// We can also use this to pass data into a remote database.
		ThisBooking := booking{
			Name:        r.FormValue("name"),
			Treatment:   r.FormValue("treatment"),
			Phone:       r.FormValue("phone"),
			BookingTime: &bookingTime,
			MinDate:     time.Now().In(loc).Format("2006-01-02"),
		}

		// First things first: we'll check if the phone number is valid
		// We don't need the lookup object; we just need to check if we encounter an error.
		_, err = lookup.Read(client, r.FormValue("phone"), &lookup.Params{CountryCode: "NL"})
		if err != nil {
			RenderDefaultTemplate(w, "views/booking.gohtml", bookingContainer{ThisBooking, "Please enter a valid phone number."})
			return
		}

		status := checkTime(w, bookingTime, reminderDiff, loc)
		if status == "Success!" {

			// Set messages to display
			successStatus := "Done! We've set up an appointment for you at " + bookingTime.Format("Mon, 02 Jan 2006 3:04 PM") +
				" for " + r.FormValue("treatment") + ". We'll send a reminder to " + r.FormValue("phone") + " at " + reminderTime.Format("Mon, 02 Jan 2006 3:04 PM") + ". Thanks for using BeautyBird!"
			reminderMessage := "Gentle reminder: you've got an appointment with BeautyBird at " + bookingTime.Format("Mon, 02 Jan 2006 3:04 PM") + ". See you then!"

			// Create a new message, and schedule it to be sent 3 hours before the booking time.
			msg, err := sms.Create(
				client,
				"BeautyBird",
				[]string{r.FormValue("phone")},
				reminderMessage,
				// Use messagebird.MessageParams to set up a schedule for the reminder SMS.
				&sms.Params{
					ScheduledDatetime: reminderTime,
				},
			)
			// If the MessageBird API encounters an error, intercept and render error message instead of breaking the application.
			if err != nil {
				log.Println(err)
				RenderDefaultTemplate(w, "views/booking.gohtml", bookingContainer{ThisBooking, fmt.Sprintln(err) + ". Please check your details and try again!"})
				return
			}

			// For development logging
			log.Println(msg)

			RenderDefaultTemplate(w, "views/booking.gohtml", bookingContainer{ThisBooking, successStatus})
			return
		} else if status != "" {
			RenderDefaultTemplate(w, "views/booking.gohtml", bookingContainer{ThisBooking, status})
			return
		}
	}
	// By default, render page with BookingEmpty object with no message.
	RenderDefaultTemplate(w, "views/booking.gohtml", bookingContainer{BookingEmpty, ""})
}

// checkTime checks if the bookingTime is within an acceptable time range, set within this function itself.
func checkTime(w http.ResponseWriter, bookingTime time.Time, reminderDiff time.Duration, loc *time.Location) string {
	// Make these easy to access.
	bookingYear := bookingTime.Year()
	bookingMonth := bookingTime.Month()
	bookingDay := bookingTime.Day()

	// Set time references. We need these for time comparisons.
	openingTime := time.Date(bookingYear, bookingMonth, bookingDay, 9, 0, 0, 0, loc)
	closingTime := time.Date(bookingYear, bookingMonth, bookingDay, 18, 0, 0, 0, loc)

	// To make sure that we always get the local time
	now := time.Now().In(loc)

	//
	timeBeforeBooking := bookingTime.Sub(now)

	switch {
	// Check if bookingTime is earlier than the time now.
	case bookingTime.Before(now):
		return "Cannot make a booking before now. Please try again!"
	// Check if earlier than openingTime.
	case bookingTime.Before(openingTime):
		return "We're not open yet! Please book your appointment between " + openingTime.Format("03:04 PM") + " and " + closingTime.Format("03:04 PM") + "."
	// Check if later than closingTime.
	case bookingTime.After(closingTime):
		return "We're closed! Please book your appointment between " + openingTime.Format("03:04 PM") + " and " + closingTime.Format("03:04 PM") + "."
	// Check if earlier than reminderDiff before closingTime.
	// Just calling reminderDiff.String() produces output like "3h0m0s",
	// so we need to split the string and take whatever comes before h.
	// More granular reminderDiffs will require more complex logic.
	case timeBeforeBooking < reminderDiff:
		return "Please book an appointment " + strings.Split(reminderDiff.String(), "h")[0] + " hours in advance."
	// In all other cases, consider booking a success. Return simple string that can be easily checked for.
	default:
		return "Success!"
	}
}

// Helpers

// RenderDefaultTemplate takes:
// - a http.ResponseWriter
// - a string that's the path to your template file
// - data to render to the template. If no data, should enter 'nil'
func RenderDefaultTemplate(w http.ResponseWriter, thisView string, data interface{}) {
	renderthis := []string{thisView, "views/layouts/default.gohtml"}
	t, err := template.ParseFiles(renderthis...)
	if err != nil {
		log.Fatal(err)
	}
	err = t.ExecuteTemplate(w, "default", data)
	if err != nil {
		log.Fatal(err)
	}
}
