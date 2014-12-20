package main

import (
  "net/http"
  "fmt"
  "time"
  "encoding/binary"
  "net"
  "net/smtp"
  calendar "code.google.com/p/google-api-go-client/calendar/v3"
  "code.google.com/p/goauth2/oauth"
)

var oauthconfig = &oauth.Config{
  ClientId: ClientId,
  ClientSecret: ClientSecret,
  Scope: calendar.CalendarScope,
  AuthURL: "https://accounts.google.com/o/oauth2/auth",
  TokenURL: "https://accounts.google.com/o/oauth2/token",
  TokenCache: oauth.CacheFile("cache.json"),
  RedirectURL: "http://max.uy/calefon/oauthcallback",
  AccessType: "offline",
}

var transport *oauth.Transport = &oauth.Transport{Config: oauthconfig}

var errorMessageSent bool = false

func Log(message string, a ...interface{}) {
  message = fmt.Sprintf(message, a...)
  fmt.Printf("%v %v\n", time.Now().Format(time.Stamp), message)
}

func SendEmailError(err error) {
  if errorMessageSent {
    return
  }
  errorMessageSent = true
  msg := `From: calefon@server.max.uy
To: max@max.uy
Subject: Calefon error

Error: ` + err.Error() + `

Login to calefon oauth again? http://max.uy/calefon`;
  err = smtp.SendMail("localhost:25", nil, "calefon@server.max.uy", []string{"max@max.uy"}, []byte(msg))
  if err != nil {
    Log("Error sending mail: %v", err)
  }
}

func handleConnection(conn net.Conn) {
  defer conn.Close()
  Log("Connection from %v", conn.RemoteAddr())
  referenceTime := time.Now()

  client := transport.Client()
  service, err := calendar.New(client)
  if err != nil {
    Log("calendar.New error %v", err)
    return
  }
  calendarId := "ppl11p5l4ulont1d4ogue1dqos@group.calendar.google.com"
  response, err := service.Freebusy.Query(&calendar.FreeBusyRequest{
    Items: []*calendar.FreeBusyRequestItem{&calendar.FreeBusyRequestItem{calendarId}},
    TimeMin: referenceTime.Format(time.RFC3339),
    TimeMax: referenceTime.Add(time.Hour*24*7).Format(time.RFC3339),
  }).Do()
  if err != nil {
    SendEmailError(err)
    Log("service.FreeBusy.Query.Do error %v", err)
    return
  }
  errorMessageSent = false
  var freeBusyCalendar = response.Calendars[calendarId]
  for _, period := range freeBusyCalendar.Busy {
    startTime, _ := time.Parse(time.RFC3339, period.Start)
    endTime, _ := time.Parse(time.RFC3339, period.End)
    startInt := int32(startTime.Sub(referenceTime)/time.Millisecond)
    endInt := int32(endTime.Sub(referenceTime)/time.Millisecond)
    binary.Write(conn, binary.LittleEndian, startInt)
    binary.Write(conn, binary.LittleEndian, endInt) 
  }
  var zero int32 = 0
  binary.Write(conn, binary.LittleEndian, zero)
  binary.Write(conn, binary.LittleEndian, zero)

  Log("Connection from %v complete", conn.RemoteAddr())
}

func tcpListen() {
  listen, err := net.Listen("tcp", ":9001")
  if err != nil {
    fmt.Println(err)
    return
  }
  Log("Waiting for TCP connections")
  for {
    conn, err := listen.Accept()
    if err != nil {
      fmt.Println(err)
      continue
    }
    go handleConnection(conn)
  }
}

func httpListen() {
  http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    if transport.Token == nil || transport.Token.Expired() {
      http.Redirect(w, r, oauthconfig.AuthCodeURL(""), http.StatusFound)
    } else {
      fmt.Fprintf(w, "Ok")
    }
  })
  http.HandleFunc("/oauthcallback", func(w http.ResponseWriter, r *http.Request) {
    code := r.FormValue("code")
    transport.Exchange(code)
    fmt.Fprintf(w, "Ok")
    Log("OAuth callback with code %v", code)
  })
  Log("Waiting for HTTP connections")
  err := http.ListenAndServe(":8081", nil)
  if err != nil {
    Log("%v", err)
  }
}

func main() {
  go tcpListen();
  httpListen();
}
