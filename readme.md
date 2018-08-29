
# SMS Appointment Reminders
### ⏱ 30 min build time

## Why build SMS appointment reminders? 

Booking appointments online from a website or mobile app is quick and easy. Customers just have to select their desired date and time, enter their personal details and hit a button. The problem however, is that easy-to-book appointments are often just as easy to forget.

For appointment-based services, no-shows are annoying and costly because of the time and revenue lost waiting for a customer instead of serving them, or another customer. Timely SMS reminders act as a simple and discrete nudges, which can go a long way in the prevention of costly no-shows.

## Getting Started

In this MessageBird Developer Guide, we'll show you how to use the MessageBird SMS messaging API to build an SMS appointment reminder application in Go. 

This sample application represents the booking website of a fictitious online beauty salon called *BeautyBird*. To reduce the growing number of no-shows, BeautyBird now collects appointment bookings through a form on their website and schedules timely SMS reminders to be sent out three hours before the selected date and time.

To look at the full sample application or run it on your computer, go to the [MessageBird Developer Guides GitHub repository](https://github.com/messagebirdguides/reminders-guide-go) and clone it or download the source code as a ZIP archive. 

## Getting started

We'll be building our single-page web application with:

* the latest version of [Go](https://golang.org), and
* the [MessageBird's REST API package for Go](https://github.com/messagebird/go-rest-api)

### Structure of your application

We're building a single page application that takes user input from our BeautyBird booking page and schedules an SMS reminder to be sent three hours before the appointment. To do this, we need to do the following:

- **Build an appointment booking page**: Our appointment booking page should take in a name, treatment details, a phone number, and a date and time for the booking.
- **Check if the phone number provided is valid**: We need to know if the phone number provided works. We'll use MessageBird's [Lookup REST API](https://developers.messagebird.com/docs/lookup) to check if a phone number is valid.
- **Check if the booking time is valid**: We should only accept appointments that are within opening hours, and if they are made at least three hours in advance.
- **Finally, schedule an SMS reminder to be sent**: After we've verified that all the booking information we've received is valid, we schedule an SMS reminder to be sent three hours before the start of the appointment.

### Project Setup

Create a folder for your application. In this folder, create the 
following subfolders:

 - `views`
 - `views/layouts`

  We'll use the following packages from the Go standard library to build our application:

- `net/http`: A HTTP package for building our routes and a simple http server.
- `html/template`: A HTML template library for building views.
- `time`: The Go standard library for handling time.

From the MessageBird Go REST API package, we'll import the following packages:

- `github.com/messagebird/go-rest-api`: The MessageBird core client package.
- `github.com/messagebird/go-rest-api/sms`: The MessageBird SMS messaging package.
- `github.com/messagebird/go-rest-api/lookup`: The MessageBird phone number lookup package.


### Create your API Key 🔑

To start making API calls, we need to generate an access key. MessageBird provides keys in _live_ and _test_ modes. For this tutorial you will need to use a live key. Otherwise, you will not be able to test the complete flow. Read more about the difference between test and live API keys [here](https://support.messagebird.com/hc/en-us/articles/360000670709-What-is-the-difference-between-a-live-key-and-a-test-key-).

Go to the [MessageBird Dashboard](https://dashboard.messagebird.com/en/user/index); if you have already created an API key it will be shown right there. Click on the eye icon to make the access key visible, then select and copy it to your clipboard. If you do not see any key on the dashboard or if you're unsure whether this key is in _live_ mode, go to the _Developers_ section and open the [API access (REST) tab](https://dashboard.messagebird.com/en/developers/access). Here, you can create new keys and manage your existing ones.

If you are having any issues creating your API key, please don't hesitate to contact support at support@messagebird.com.

**Pro-tip:** To keep our demonstration code simple, we will be saving our API key in `main.go`. However, hardcoding your credentials in the code is a risky practice that should never be used in production applications. A better method, also recommended by the [Twelve-Factor App Definition](https://12factor.net/), is to use environment variables. You can use open source packages such as [GoDotEnv](https://github.com/joho/godotenv) to read your API key from a `.env` file into your Go application. Your `.env` file should be written as follows:

`````env
MESSAGEBIRD_API_KEY=YOUR-API-KEY
`````

To use [GoDotEnv](https://github.com/joho/godotenv) in your application, install it by running:

````bash
go get -u github.com/joho/godotenv
````

Then, import it in your application:

````go
import (
  // Other imported packages
  "os"

  "github.com/joho/godotenv"
)

func main(){
  // GoDotEnv loads any ".env" file located in the same directory as main.go
  err := godotenv.Load()
  if err != nil {
    log.Fatal("Error loading .env file")
  }

  // Store the value for the key "MESSAGEBIRD_API_KEY" in the loaded '.env' file.
  apikey := os.Getenv("MESSAGEBIRD_API_KEY")

  // The rest of your application ...
}
````

## Initialize the MessageBird Client

Install the [MessageBird's REST API package for Go](https://github.com/messagebird/go-rest-api) by running:

````go
go get -u github.com/messagebird/go-rest-api
````

In your project folder, create a `main.go` file, and write the following code:

````go
package main

import (
  "github.com/messagebird/go-rest-api"
)

var client *messagebird.Client

func main(){
  client = messagebird.New(<enter-your-apikey>)
}
````

## Building an appointment booking page

Our goal here is to set up a page to collect customer details and, most importantly, their phone number so that we can send them an SMS reminder three hours before their appointment.

First, we'll set up our templates and routes. Modify `main.go` to look like the following:

````go
package main

import (
    "log"
    "net/http"
    "html/template"

    "github.com/messagebird/go-rest-api"
    )

var client *messagebird.Client

func main(){
    client = messagebird.New(<enter-your-apikey>)

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

func bbScheduler(w http.ResponseWriter, r *http.Request){
    RenderDefaultTemplate(w,"views/booking.gohtml",nil)
}

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
````

Here, we've set up a route that's handled by the `bbScheduler` handler function. We've also set up a helper function `RenderDefaultTemplate()` that parses our `default.gohtml` template and one other template, and handles possible errors.

With that done, we'll set up our `default` template. Create `views/layouts/default.gohtml` and write the following code:

````html
{{ define "default" }}
<!DOCTYPE html>
  <head>
    <meta charset="utf-8">
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <title>MessageBird Verify Example</title>
    <meta name="description" content="">
    <meta name="viewport" content="width=device-width, initial-scale=1">
  </head>
  <body>
    <main>
    {{ template "yield" . }}
    </main>
  </body>
</html>
{{ end }}
````

And our `booking` template. Create `views/booking.gohtml` and add the following code to it:

````html
{{ define "yield" }}
<h1>BeautyBird &lt;3</h1>
<p>Book an appointment for a treatment in our salon, right here on our website!</p>
<form method="post" action="/">
    <div>
        <label>Your name:</label>
        <br />
        <input type="text" name="name" required/>
    </div>
    <div>
        <label>Your desired treatment:</label>
        <br />
        <input type="text" name="treatment" required/>
    </div>
    <div>
        <label>Your mobile number (e.g. +31624971134):</label>
        <br />
        <input type="tel" name="phone" required/>
    </div>
    <div>
        <label>Date and Time (<small>Please book at least 3 hours in advance.</small>):</label>
        <br/>
        <input type="date" name="date" required/>
        <input type="time" name="time" required/>
    </div>
    <div>
        <button type="submit">Book Now!</button>
    </div>
</form>
{{ end }}
````

Done! If you run `go run main.go` now and navigate to `http://localhost:8080`, you'll find your booking appointment page ready to accept new appointments (and immediately forget them; we'll deal with that in a moment).

## Storing Appointments & Scheduling Reminders

Let's write some code to pull date from the booking form.

First, we need to create data structures to contain our appointment information. Just above `main()`, add the following code:

````go
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
````

Here, we're setting up two structs to contain data we're parsing from the submitted booking form, and the text that we want to pass back into our template to display a state message that tells our customers if their booking was sucecssful. Because (1) we can only pass one data object into a template, and (2) we want to manage our booking data separately from the status message, we'll set up two separate struct types: `booking`, and `bookingContainer`.

**Note**: This guide shows you how to schedule an SMS reminder for a single appointment. In production, you'll need to save the appointment information in a persistent data store. To keep the guide straightforward, we've omitted those instructions.

### Parse the booking form

Let's get our customer's booking information out of that form. We'll add code to `bbScheduler()` to:

1. Parse the submitted form.
2. Extract data from the submitted form and add them to a `ThisBooking` struct.
3. Update the view with this new information.

Modify `bbScheduler()` to look like the following:

````go
func bbScheduler(w http.ResponseWriter, r *http.Request){

    if r.Method == "POST" {
        r.ParseForm()

        ThisBooking := booking{
            Name:        r.FormValue("name"),
            Treatment:   r.FormValue("treatment"),
            Phone:       r.FormValue("phone"),
            BookingTime: bookingTime,
            MinDate:     minDate,
        }
        RenderDefaultTemplate(w,"views/booking.gohtml",bookingContainer{ThisBooking,"Booking successful!"})
        return
    }

    RenderDefaultTemplate(w,"views/booking.gohtml",nil)
}
````

Here, we're parsing the form with `r.ParseForm()` and putting information that we've extracted from the booking form into a `ThisBooking` struct. Then we call `RenderDefaultTemplate()` again, and pass `ThisBooking` and a `"Booking successful!"` message to the template with a `bookingContainer{}` struct literal.

### Working with `time`

Notice that in our `booking` struct, `BookingTime` is of type `time.Time`. This is a data structure that Go uses to work with time values. Using the `time` package gives us the ability to set and compare time values, features that we'll need when writing code to check if a booking time is valid. We'll need to add code to `bbScheduler()` to:

1. Import the `time` package.
2. Define a time locale. If we don't define a time locale, then the `time` package will assume that its working with UTC.
2. convert the value of `r.FormValue("date")` and `r.FormValue("time")` to a format that is digestable by the `time` package.

Modify the `import` statement in `main.go` to add the `time` package:

````go
import (
    "log"
    "net/http"
    "html/template"
    "time"

    "github.com/messagebird/go-rest-api"
    )
````

Then, rewrite `bbScheduler()` so that it looks like this:

````go
func bbScheduler(w http.ResponseWriter, r *http.Request){

    loc, err = time.LoadLocation("Europe/Amsterdam")
    if err != nil {
        log.Println(err)
    }

    if r.Method == "POST" {
        r.ParseForm()

        bookingTime, err := time.ParseInLocation("2006-01-02 15:04", r.FormValue("date")+"  "+r.FormValue("time"), loc)
        if err != nil {
            log.Println(err)
        }

        ThisBooking := booking{
            Name:        r.FormValue("name"),
            Treatment:   r.FormValue("treatment"),
            Phone:       r.FormValue("phone"),
            BookingTime: &bookingTime,
            MinDate:     time.Now().In(loc).Format("2006-01-02"),
        }
        RenderDefaultTemplate(w,"views/booking.gohtml",ThisBooking)
        return
    }

    RenderDefaultTemplate(w,"views/booking.gohtml",nil)
````

There are several things happening here, so let's break it down.

#### a. Define a time locale

We're defining a location at the top of `bbScheduler()` to make sure that we're working in the correct time locale. Because we don't expect BeautyBird customers to cross timezones for their appointments, we can hardcode this value here as the `loc` variable.

#### b. Parse date and time from form input

To get the the date and time values extracted from our appointment form into a format that the `time` package understands, we need to tell `time` to parse those values. In the above code, we call 
`time.ParseInLocation()` which takes three parameters: a "layout" string, a "value" string that specifies a specific date and time, and a "location" (which we've defined as `loc`). 

The "layout" string is Go's way of allowing you to quickly specify in what format the date and time is written in "value". When we write:

````go
time.ParseInLocation("2006-01-02 15:04", r.FormValue("date")+"  "+r.FormValue("time"), loc)
````

we're telling Go to parse `r.FormValue("date") + " " + r.FormValue("time")` with the "layout": `"2006-01-02 15:04"`. 

This "layout" can be read like this:

- `2006`: tells Go to expect a four digit "year" value.
- `01`: A two digit "month" value.
- `02`: A two digit "day" value.
- `15`: A two digit "hour" value, in a 24-hour format. We can also tell Go to expect an "hour" value in a 12-hour format by writing `03` instead.
- `04`: A two digit "minute" value.
- `PM`: We're not using it here yet, but we tell Go to expect a `PM` or `AM` value by writing `PM` (or `pm` for lowercase).

#### c. Set a MinDate value

We're using `<input type=date/>` to allow our customers to enter the a date for their booking. To prevent them from selecting a date that's earlier than today, we can write a `min` attribute that contains today's date. In the code above, we get today's date and assign it to the `MinDate` field in `ThisBooking`, which we then pass to the template when we call `RenderDefaultTemplate()`.

Once we've parsed the values submitted through the booking form, we assign these values to their corresponding struct fields in `ThisBooking`:

````go
ThisBooking := booking{
        Name:        r.FormValue("name"),
        Treatment:   r.FormValue("treatment"),
        Phone:       r.FormValue("phone"),
        BookingTime: &bookingTime,
        MinDate:     time.Now().In(loc).Format("2006-01-02"),
    }
````

For the purposes of this guide, we'll use `ThisBooking` once to set up a scheduled SMS reminder, and then forget it once the next booking is received. In production, you'll want to push each new instance of `ThisBooking` to a persistent data store to save that information.

## Setting up templates to use booking data

Now, we can set up our template to use all the data that we're passing through the `RenderDefaultTemplate()` call.

Modify your `booking.gohtml` file to look like this:

````go
{{ define "yield" }}
    <h1>BeautyBird &lt;3</h1>
    <p>Book an appointment for a treatment in our salon, right here on our website!</p>
    <form method="post" action="/">
        <div>
            <label>Your name:</label>
            <br />
            <input type="text" name="name" {{ if .Booking.Name }} value="{{ .Booking.Name }}"{{ end }} required/>
        </div>
        <div>
            <label>Your desired treatment:</label>
            <br />
            <input type="text" name="treatment" {{ if .Booking.Treatment }} value="{{ .Booking.Treatment }}"{{ end }} required/>
        </div>
        <div>
            <label>Your mobile number (e.g. +31624971134):</label>
            <br />
            <input type="tel" name="phone" {{ if .Booking.Phone }} value="{{ .Booking.Phone }}"{{ end }} required/>
        </div>
        <div>
            <label>Date and Time (<small>Please book at least 3 hours in advance.</small>):</label>
            <br/>
            <input type="date" name="date" min="{{ .Booking.MinDate }}" required/>
            <input type="time" name="time" required/>
        </div>
        <div>
            <button type="submit">Book Now!</button>
        </div>
    </form>

    {{ if .Message }}
    <section>
    <strong>{{ .Message }}</strong>
    </section>
    {{ end }}
{{ end }}
````

For our "name", "treatment", and "phone" fields, we're adding `{{ if .FieldName }}value="{{ .FieldName }}"{{ end }}` blocks to tell our template to display a field value if it's been defined and available. This allows us to display field values entered for the previous form submissions. This allows us to handle a case where a submission fails — our customer can check and resubmit their booking details without having to re-enter information.

We've also added a `min` attribute to our `<input type="date"/>` line, so that customers cannot select a date before the present date.

Finally, we're displaying a status message (if any) at the bottom of the template with an `{{ if .Message }}{{ .Message }}{{ end }}` block.

## Process form input

Now that we've got our basic application structure and templates set up, we can start writing code to (1) check if we have enough valid information to schedule an appointment, and (2) schedule an SMS reminder for the appointment.

### Checking appointment information

The first check we need to do — that all the form fields are filled in on submission — is already handled by the `required` attribute we've added to all our `input` fields.

The other two checks that we need to do are:

- Checking if the entered phone number is valid, and
- Checking if the appointment is set for a valid date and time.

#### a. Checking phone number validity

We can use MessageBird's Lookup REST API to check if a phone number is valid.

First, we need to add the MessageBird Lookup package to our `import` statement:

````go
import (
    "log"
    "net/http"
    "html/template"
    "time"

    "github.com/messagebird/go-rest-api"
    "github.com/messagebird/go-rest-api/lookup"
    )
````

Then, in `bbScheduler()`, add this code just under `r.ParseForm()`:

````go
    _, err = lookup.Read(client, r.FormValue("phone"), &lookup.Params{CountryCode: "NL"})
    if err != nil {
        RenderDefaultTemplate(w, "views/booking.gohtml", bookingContainer{ThisBooking, "Please enter a valid phone number."})
        return
    }
````

Here, we're calling `lookup.Read()` and passing in the phone number we've gotten through the booking form, and a `CountryCode` field value that tells MessageBird which locality the phone number should belong to. If MessageBird cannot validate the phone number, `lookup.Read()` returns an error, which we catch in the following `if err != nil { ... }` block. We're discarding the resulting `lookup` object by assigning it to `_` because we don't need it in our application.

**Note**: To send a message to a phone number, you must have added that phone number to your MessageBird contact list. For more information on how to add a new phone number to your contact list, see the [MessageBird API Reference](https://developers.messagebird.com/docs/contacts#create-a-contact).

#### b. Checking appointment date and time

Next, we need to add code that checks if the date and time of the booking meets the following criteria:

- **Not before "now"**: We're preventing customers from selecting a date before today in our `booking.gohtml` template, but we also need to check if the time of the appointment is before the current time. We can't do this check in the template because date and time input values are checked separately. For example, adding a `min="15:00"` attribute value to `<input type="time"/>` would stop customers from booking appointments before 3 PM on any day, instead of just the current day. Instead, we'll check the date and time together in `bbScheduler()` when the form is submitted.
- **Within opening hours**: We don't want our customers to be able to book an appointment after hours, so we'll only allow bookings between 9 AM and 6 PM. If a customer tries to book an appointment after hours, we'll display a message asking them to book their appointment within opening hours.
- **There is at least 3 hours between the current time and the set appointment**: We want at least 3 hours of advanced notice for each appointment, so that the staff at BeautyBird have the time to prepare for it (and so our application can send that SMS reminder).

We'll write a helper function `checkTime()` to perform these checks. Add this code just under `main()`:

````go
func checkTime(w http.ResponseWriter, bookingTime time.Time, reminderDiff time.Duration, loc *time.Location) string {
    bookingYear := bookingTime.Year()
    bookingMonth := bookingTime.Month()
    bookingDay := bookingTime.Day()

    openingTime := time.Date(bookingYear, bookingMonth, bookingDay, 9, 0, 0, 0, loc)
    closingTime := time.Date(bookingYear, bookingMonth, bookingDay, 18, 0, 0, 0, loc)

    now := time.Now().In(loc)

    timeBeforeBooking := bookingTime.Sub(now)

    switch {
    case bookingTime.Before(now):
        return "Cannot make a booking before now. Please try again!"
    case bookingTime.Before(openingTime):
        return "We're not open yet! Please book your appointment between " + openingTime.Format("03:04 PM") + " and " + closingTime.Format("03:04 PM") + "."
    case bookingTime.After(closingTime):
        return "We're closed! Please book your appointment between " + openingTime.Format("03:04 PM") + " and " + closingTime.Format("03:04 PM") + "."
    case timeBeforeBooking < reminderDiff:
        return "Please book an appointment " + strings.Split(reminderDiff.String(), "h")[0] + " hours in advance."
    default:
        return "Success!"
    }
}
````

Then, in `bbScheduler()`, add the following lines of code after `ThisBooking := booking{ ... }`:

````go
var reminderDiff time.Duration

reminderDiff, err = time.ParseDuration("3h")
if err != nil {
    log.Println(err)
}

status := checkTime(w, bookingTime, reminderDiff, loc)
````

Remember the how we defined `bookingTime` earlier to help populate our `ThisBooking` struct? We're using it to:

- **Set a `reminderDiff` value**: To subtract three hours from bookingTime, we need to set a variable of `time.Duration` type. Here, we set a duration of three hours using the `time.ParseDuration()` function and assign it to `reminderDiff`. We then pass `reminderDiff` into our `checkTime()` helper function.
- **Check if our three time and date criteria above**: In `checkTime()`, we set up a switch statement that checks if `bookingTime` fulfils various conditions that makes the booked time and date invalid. If it meets any one of the conditions, then the switch statement returns a predefined string that we assign to the `status` variable, and can display as an error message in our `booking.gohtml` template. If it passes all the conditions, then we return the string `"Success!"`, which also gets assigned to the `status` variable.

## Checkpoint: What `bbScheduler()` should look like now

This is how your `bbScheduler` handler should look like now:

````go
func bbScheduler(w http.ResponseWriter, r *http.Request) {
    var (
        loc          *time.Location
        reminderDiff time.Duration
        err          error
    )

    loc, err = time.LoadLocation("Europe/Amsterdam")
    if err != nil {
        log.Println(err)
    }

    reminderDiff, err = time.ParseDuration("3h")
    if err != nil {
        log.Println(err)
    }

    BookingEmpty := booking{
        MinDate: time.Now().In(loc).Format("2006-01-02"),
    }

    if r.Method == "POST" {
        r.ParseForm()

        bookingTime, err := time.ParseInLocation("2006-01-02 15:04", r.FormValue("date")+" "+r.FormValue("time"), loc)
        if err != nil {
            log.Println(err)
        }

        reminderTime := bookingTime.Add(-reminderDiff)

        ThisBooking := booking{
            Name:        r.FormValue("name"),
            Treatment:   r.FormValue("treatment"),
            Phone:       r.FormValue("phone"),
            BookingTime: &bookingTime,
            MinDate:     time.Now().In(loc).Format("2006-01-02"),
        }

        _, err = lookup.Read(client, r.FormValue("phone"), &lookup.Params{CountryCode: "NL"})
        if err != nil {
            RenderDefaultTemplate(w, "views/booking.gohtml", bookingContainer{ThisBooking, "Please enter a valid phone number."})
            return
        }

        status := checkTime(w, bookingTime, reminderDiff, loc)
        if status == "Success!" {
            // =====
            // Display success state and actually schedule an SMS reminder.
            // =====
        } else if status != "" {
            RenderDefaultTemplate(w, "views/booking.gohtml", bookingContainer{ThisBooking, status})
            return
        }
    }
    RenderDefaultTemplate(w, "views/booking.gohtml", bookingContainer{BookingEmpty, ""})
}
````

We've cleaned up the code a little by grouping some variable assignments, as well added a `BookingEmpty` struct to send an initial `MinDate` value to our template when the booking page first loads.

We've also set up some logic to display the `status` returned by our `checkTime()` call for any other `status` than `"Success!"`. All that's left to do is to write the code for our `"Success!"` state.

## Writing the success state and scheduling an SMS reminder

We'll modify this block from the code in the above section:

````go
if status == "Success!" {
            // =====
            // Display success state and actually schedule an SMS reminder.
            // =====
}
````

First, we'll write the messages that (1) we should display as a success state on our booking page, and (2) the message that we want to send our as a scheduled SMS reminder. To the top of the `if status == "Success!" {...}` block, add:

````go
successStatus := "Done! We've set up an appointment for you at " + bookingTime.Format("Mon, 02 Jan 2006 3:04 PM") +
    " for " + r.FormValue("treatment") + ". We'll send a reminder to " + r.FormValue("phone") + " at " + reminderTime.Format("Mon, 02 Jan 2006 3:04 PM") + ". Thanks for using BeautyBird!"
reminderMessage := "Gentle reminder: you've got an appointment with BeautyBird at " + bookingTime.Format("Mon, 02 Jan 2006 3:04 PM") + ". See you then!"
````

Next, we finally get to schedule an SMS reminder. 

First, add the MessageBird SMS Message package to your `import` statement:

````go
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
````

Then, add the following code to the `if status == "Sucess!" { ...}` block:

````go
msg, err := sms.Create(
    client,
    "BeautyBird",
    []string{r.FormValue("phone")},
    reminderMessage,
    &sms.Params{
        ScheduledDatetime: reminderTime,
    },
)
if err != nil {
    log.Println(err)
    RenderDefaultTemplate(w, "views/booking.gohtml", bookingContainer{ThisBooking, fmt.Sprintln(err) + ". Please check your details and try again!"})
    return
}
log.Println(msg)
RenderDefaultTemplate(w, "views/booking.gohtml", bookingContainer{ThisBooking, successStatus})
return
````

Here, we call `sms.Create()` with the following parameters:

- `client` is our client object.
- `"BeautyBird"` as our "originator".
- `[]string{r.FormValue("phone")}` is a list of phone numbers, which we assign the phone number from our submitted form.
- `reminderMessage` as the "body" of our message.
- `&sms.Params{...}` is a list of message parameters, of which we only define the `ScheduledDatetime` field value.

If the call succceeds, we get a `*sms.Message` object that we save to the `msg` variable and log for development. (You can safely discard this object by replacing `msg` with `_` if you don't need to log the `*sms.Message` object.) We then call `RenderDefaultTemplate` and pass in `successStatus` in our `bookingContainer` struct, and tell our application to skip the rest of `bbScheduler()`.

If we encounter an error while creating a new message, we'll log and display the error, and skip the rest of `bbScheduler()` with:

````go
if err != nil {
    log.Println(err)
    RenderDefaultTemplate(w, "views/booking.gohtml", bookingContainer{ThisBooking, fmt.Sprintln(err) + ". Please check your details and try again!"})
    return
}
````

## Testing the Application

You're done! To test your application, navigate to your project folder in the terminal and run:

`go run main.go`

Then, point your browser at http://localhost:8080/ to see the form and schedule your appointment! If you've used a live API key, a message will arrive to your phone three hours before the appointment! But don't actually leave the house, this is just a demo :)


## Nice work!

You now have a running SMS appointment reminder application!

You can now use the flow, code snippets and UI examples from this tutorial as an inspiration to build your own SMS reminder system. Don't forget to download the code from the [MessageBird Developer Guides GitHub repository](https://github.com/messagebirdguides/reminders-guide).

## Next steps

Want to build something similar but not quite sure how to get started? Please feel free to let us know at support@messagebird.com, we'd love to help!
