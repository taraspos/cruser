package keys

import (
	"log"
	"regexp"
)

type SSHKey struct {
	Line     string
	Type     string
	Key      string
	Username string
	Email    string
}

type SSHKeys []SSHKey

var Match = regexp.MustCompile(`(\S*) (\S*) ((\w*)@.*)`)

// parse ssh keys, and save next data from it. Example:
// ssh-rsa  aaaaaa...bbbbbbbb test@user.com
// 0. ssh-key       - ssh-rsa  aaaaaa...bbbbbbbb test@user.com
// 1. Type          - ssh-rsa
// 2. Key           - aaaaaaa...bbbbbbbb
// 3. Email address - test@user.com
// 4. Username      - test
func Parse(lines []string) SSHKeys {
	var sks SSHKeys
	for _, key := range lines {
		var sk SSHKey
		match := Match.FindStringSubmatch(key)

		if len(match) == 5 {
			sk.Line = match[0]
			sk.Type = match[1]
			sk.Key = match[2]
			sk.Email = match[3]
			sk.Username = match[4]

			if sk.Type != "ssh-rsa" && sk.Type != "ssh-dss" && sk.Type != "ssh-ed25519" && sk.Type != "ecdsa-sha2-nistp256" {
				log.Printf("'%s' - not known SSH Key type ", sk.Type)
			} else {
				sks = append(sks, sk)
			}
		}
	}
	return sks
}
