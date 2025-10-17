package test

import (
	"fmt"
	"testing"
	"time"

	goemail "github.com/JiuYu77/go-email"
)

func TestCache(t *testing.T) {
	cache := goemail.NewCache(10 * time.Second)
	cache.Set("key", &goemail.VerificationCode{
		Email:     email,
		Code:      "123456",
		ExpiresAt: time.Now().Add(15 * time.Second),
		Used:      false,
	})
	code, ok := cache.Get("key")
	if !ok {
		t.Error("cache get failed")
	}
	value, ok := code.(*goemail.VerificationCode)
	if !ok {
		t.Error("cache get type failed")
	} else {
		if value.Email != email {
			t.Error("cache get email failed")
		}
		if value.Code != "123456" {
			t.Error("cache get code failed")
		}
		if value.Used {
			t.Error("cache get used failed")
		}
	}

	time.Sleep(20 * time.Second)

	_, ok = cache.Get("key")
	if !ok {
		fmt.Println(">>>>cache value has expired, success")
	} else {
		t.Errorf(">>>>cache test expired failed")
	}
}
