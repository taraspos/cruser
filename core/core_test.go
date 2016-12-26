package core

import (
	"reflect"
	"testing"

	"os"

	"github.com/trane9991/cruser/core/keys"
	"github.com/trane9991/cruser/core/user"
	"github.com/trane9991/cruser/core/utils"
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
}, keys.SSHKey{
	Line:     "ssh-dss ssssssssssssssssss new@user.com",
	Key:      "ssssssssssssssssss",
	Type:     "ssh-dss",
	Email:    "new@user.com",
	Username: "new",
}}

var expectedUsers = users{user.Profile{
	Name:              "test",
	SSHAuthorizedKeys: []string{"ssh-rsa bbbbbbbbbbbbbbbbbbbbbb test@gmail.com", "ssh-dss cccccccccccccccccccccc test@outlook.com"},
	Comment:           "test@gmail.com;test@outlook.com;",
}, user.Profile{
	Name:              "hello",
	SSHAuthorizedKeys: []string{"ssh-rsa cccccccccccccccccccccc hello@gmail.com", "ssh-rsa cccccccccccccccccccccc hello@gmail.com"},
	Comment:           "hello@gmail.com;hello@gmail.com;",
}, user.Profile{
	Name:              "new",
	SSHAuthorizedKeys: []string{"ssh-dss ssssssssssssssssss new@user.com"},
	Comment:           "new@user.com",
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

func TestGetUsersFromFileWithKeys(t *testing.T) {
	us := getUsersFromFileWithKeys("../users")
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

func TestUsersCreate(t *testing.T) {
	if os.Getenv("IN_DOCKER") != "true" {
		t.Logf("No variable 'IN_DOCKER' specified. We don't want to create extra users on your system, so skipping this test...")
		t.Skip()
	}
	path := "../users"

	lines := utils.ReadLinesFromFile(path)
	sks := keys.Parse(lines)
	us := keys2users(sks)

	for _, u := range us {
		if u.Exists() {
			t.Logf("User '%s' exist, but he should... Fail", u.Name)
			t.Fail()
		}
	}

	dryRun := false
	Run(&path, &dryRun)

	for _, u := range us {
		if !u.Exists() {
			t.Logf("User '%s' not exist, but he should... Fail", u.Name)
			t.Fail()
		}
		t.Logf("")
		t.Logf("User '%s' exist", u.Name)
		lines := utils.ReadLinesFromFile(path)
		sks := keys.Parse(lines)
		t.Log("Checking SSH keys...")

		for _, sk := range sks {
			for i, usk := range u.SSHAuthorizedKeys {
				if sk.Line == usk {
					t.Logf("SSH key '%s' exist", usk)
					break
				}
				if i == len(usk) {
					t.Logf("SSH couldn't find match for SSH key: '%s", usk)
					t.Fail()
				}
			}
		}

	}
}
