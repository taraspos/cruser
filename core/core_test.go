package core

import (
	"reflect"
	"testing"

	"github.com/trane9991/cruser/core/keys"
	"github.com/trane9991/cruser/core/user"
)

var testKeys = keys.SSHKeys{keys.SSHKey{
	Line:     "ssh-rsa bbbbbbbbbbbbbbbbbbbbbb test@gmail.com",
	Key:      "bbbbbbbbbbbbbbbbbbbbbb",
	Type:     "ssh-rsa",
	Email:    "test@gmail.com",
	Username: "test",
}, keys.SSHKey{
	Line:     "ssh-dss cccccccccccccccccccccc test@outlook.com",
	Key:      "cccccccccccccccccccccc",
	Type:     "ssh-dss",
	Email:    "test@outlook.com",
	Username: "test",
}, keys.SSHKey{
	Line:     "ssh-rsa cccccccccccccccccccccc hello@gmail.com",
	Key:      "cccccccccccccccccccccc",
	Type:     "ssh-rsa",
	Email:    "hello@gmail.com",
	Username: "hello",
}, keys.SSHKey{
	Line:     "ssh-rsa cccccccccccccccccccccc hello@gmail.com",
	Key:      "cccccccccccccccccccccc",
	Type:     "ssh-rsa",
	Email:    "hello@gmail.com",
	Username: "hello",
}}

var expectedUsers = users{user.Profile{
	Name:              "test",
	SSHAuthorizedKeys: []string{"ssh-rsa bbbbbbbbbbbbbbbbbbbbbb test@gmail.com", "ssh-dss cccccccccccccccccccccc test@outlook.com"},
	Comment:           "test@gmail.com;test@outlook.com;",
}, user.Profile{
	Name:              "hello",
	SSHAuthorizedKeys: []string{"ssh-rsa cccccccccccccccccccccc hello@gmail.com", "ssh-rsa cccccccccccccccccccccc hello@gmail.com"},
	Comment:           "hello@gmail.com;hello@gmail.com;",
}}

func TestKeys2Users(t *testing.T) {
	us := keys2users(testKeys)
	for _, u := range us {
		for i, expectedUser := range expectedUsers {
			if reflect.DeepEqual(u, expectedUser) {
				break
			}
			if i == len(expectedUsers) {
				t.Fail()
			}
		}
	}
}
