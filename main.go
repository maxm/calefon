package main

import (
  "net/http"
  "fmt"
  "io/ioutil"
  "encoding/xml"
  "time"
  "encoding/binary"
  "net"
)

type Feed struct {
  Entries []Entry `xml:"entry"`
}

type Entry struct {
  When When `xml:"http://schemas.google.com/g/2005 when"`
}

type When struct {
  StartTime string `xml:"startTime,attr"`
  EndTime string `xml:"endTime,attr"`
}

func handleConnection(conn net.Conn) {
  defer conn.Close()
  fmt.Printf("%v Connection from %v\n", time.Now().Format(time.Stamp), conn.RemoteAddr())
  referenceTime := time.Now()

  // I guess this URL shouldn't be in the repo, but I think no harm can be done with it. Be nice :)
  resp, err := http.Get("https://www.google.com/calendar/feeds/ppl11p5l4ulont1d4ogue1dqos%40group.calendar.google.com/private-7d768e7805fe231218a3dfb54977c41d/free-busy?singleevents=true&sortorder=a&orderby=starttime&futureevents=true")
  if err != nil {
    fmt.Println(err)
  } else {
    defer resp.Body.Close()
    body, _ := ioutil.ReadAll(resp.Body)
    
    var feed Feed
    xml.Unmarshal(body, &feed)
   
    for i:=0; i < len(feed.Entries) && i < 20; i++ {
      entry := feed.Entries[i]
      startTime, _ := time.Parse(time.RFC3339, entry.When.StartTime)
      endTime, _ := time.Parse(time.RFC3339, entry.When.EndTime)
      startInt := int32(startTime.Sub(referenceTime)/time.Millisecond)
      endInt := int32(endTime.Sub(referenceTime)/time.Millisecond)
      binary.Write(conn, binary.LittleEndian, startInt)
      binary.Write(conn, binary.LittleEndian, endInt) 
    }
    var zero int32 = 0
    binary.Write(conn, binary.LittleEndian, zero)
    binary.Write(conn, binary.LittleEndian, zero)
  }

  fmt.Printf("%v Connection from %v complete\n", time.Now().Format(time.Stamp), conn.RemoteAddr())
}

func main() {
  listen, err := net.Listen("tcp", ":9001")
  if err != nil {
    fmt.Println(err)
    return
  }
  fmt.Printf("%v Waiting for connections\n", time.Now().Format(time.Stamp))
  for {
    conn, err := listen.Accept()
    if err != nil {
      fmt.Println(err)
      continue
    }
    go handleConnection(conn)
  }
}
