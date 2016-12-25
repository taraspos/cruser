package keys

import (
	"bufio"
	"log"
	"reflect"
	"strings"
	"testing"
)

const testKeys = `
asdas aaaaaaaaaaaaaaaaaaaaaa test@gmail.com
ssh-rsa bbbbbbbbbbbbbbbbbbbbbb test@gmail.com
ssh-dss cccccccccccccccccccccc test@outlook.com
ssh-rsa cccccccccccccccccccccc hello@gmail.com
ssh-rsa cccccccccccccccccccccc hello@gmail.com`

var expectedKeys = SSHKeys{SSHKey{
	Line:     "ssh-rsa bbbbbbbbbbbbbbbbbbbbbb test@gmail.com",
	Key:      "bbbbbbbbbbbbbbbbbbbbbb",
	Type:     "ssh-rsa",
	Email:    "test@gmail.com",
	Username: "test",
}, SSHKey{
	Line:     "ssh-dss cccccccccccccccccccccc test@outlook.com",
	Key:      "cccccccccccccccccccccc",
	Type:     "ssh-dss",
	Email:    "test@outlook.com",
	Username: "test",
}, SSHKey{
	Line:     "ssh-rsa cccccccccccccccccccccc hello@gmail.com",
	Key:      "cccccccccccccccccccccc",
	Type:     "ssh-rsa",
	Email:    "hello@gmail.com",
	Username: "hello",
}, SSHKey{
	Line:     "ssh-rsa cccccccccccccccccccccc hello@gmail.com",
	Key:      "cccccccccccccccccccccc",
	Type:     "ssh-rsa",
	Email:    "hello@gmail.com",
	Username: "hello",
}}

func TestParse(t *testing.T) {
	lines := ReadLines(testKeys)
	sks := Parse(lines)
	if !reflect.DeepEqual(sks, expectedKeys) {
		t.Logf("%+v", sks)
		t.Logf("%+v", expectedKeys)
		t.Fail()
	}
}

// ReadLines from multiline string
func ReadLines(multiline string) []string {
	var lines []string
	scanner := bufio.NewScanner(strings.NewReader(multiline))
	for scanner.Scan() {
		line := scanner.Text()
		lines = append(lines, line)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return lines
}
