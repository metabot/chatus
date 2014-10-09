package chatus

import (
	"testing"
	"time"
)


func TestValidate(t *testing.T) {

	s := &WechatStation{
		defaultProcessor: DefaultProcessor,
		id: "test",
		token: "foobar",
	}

	err := s.IsValid("1232234123451", "adav2341asdf=-eq", "720fd5568eabf4cb9f9bd81dea494880d138253a")
	if err != nil {
		t.Fatal(err)
	}

	err = s.IsValid("1232234123451", "adav2341as", "720fd5568eabf4cb9f9bd81dea494880d138253a")
	if err == nil {
		t.Fatal("expected failure")
	}

}


func TestParseAccessToke(t *testing.T) {

	in := `{"access_token":"xyz", "expires_in":7200}`

	atk, err := parseAccessToken(in)
	if err != nil {
		t.Fatal(err)
	}

	if atk.Token != "xyz" {
		t.Fatal("expected: xyz  actual:", atk.Token)
	}

	if atk.ExpiresIn != 7200 {
		t.Fatal("expected: 7200 actual:", atk.ExpiresIn)
	}

	future := time.Now().Add(time.Duration(7000)*time.Second)
	if !atk.ExpiresAt.After(future) {
		t.Fatal("wrong expiration time")
	}

	future = time.Now().Add(time.Duration(7500)*time.Second)
	if atk.ExpiresAt.After(future) {
		t.Fatal("wrong expiration time")
	}
}

