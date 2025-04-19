package pokecache

import (
	"testing"
	"time"
)

func TestAddGet(t *testing.T) {
    duration := 100 * time.Millisecond

    cache := NewCache(duration)
    _, ok := cache.Get("AnyKey")

    if ok {
        t.Errorf("Wrong key exists") 
    }

    k, v := "key", []byte("TEST")
    cache.Add(k, v)
    val, ok := cache.Get(k)
    if !ok {
        t.Errorf("Could not get existing key %s", k) 
    }

    if string(val) != string(v) {
        t.Errorf("Got wrong value") 
    }

    time.Sleep(duration * 4)

    _, ok = cache.Get(k)
    if ok {
        t.Errorf("Expired key was not deleted") 
    }
}
