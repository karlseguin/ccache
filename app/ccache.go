package main

import (
  "fmt"
  "time"
  "ccache"
)

func main() {
  c := ccache.New(ccache.Configure().PromoteDelay(time.Second * 0))
  fmt.Println(c.Get("abc"))
  c.Set("abc", new("xxx"))
  fmt.Println(c.Get("abc"))
  time.Sleep(time.Second)
  fmt.Println(c.Get("abc"))
  time.Sleep(time.Second)
}

type Item struct {
  value string
  expires time.Time
}

func (i *Item) Expires() time.Time {
  return i.expires
}

func new(value string) *Item {
  return &Item{value, time.Now().Add(time.Minute)}
}
