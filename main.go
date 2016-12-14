package main

import (
	"bufio"
	"flag"
	"log"
	"os"
	"regexp"

	"fmt"

	"github.com/trane9991/cruser/user"
	"io/ioutil"
)

type sshKeys []string

var re = regexp.MustCompile(`ssh-rsa \S* ((\w*)@.*)`)

type users []user.Profile

func main() {
	var sk sshKeys
	var us users
	fileWithKeys := flag.String("file", "users", "File with the list of SSH-keys and user emails in format of '~/.ssh/authorized_keys' file")
	dryRun := flag.Bool("dry-run", false, "Do not execute commands, just print them.")
	flag.Parse()
	sk.readKeys(*fileWithKeys)
	us.parseKey(sk)
	user.DryRun = *dryRun
	if user.DryRun {
		log.SetOutput(ioutil.Discard)
	}
	for _, u := range us {
		// TODO: allow appending ssh-keys to existing users
		if !u.Exists() {
			u.Shell = "/bin/bash"
			if err := u.Create(); err != nil {
				log.Printf("Failed creating user '%s': %v", u.Name, err)
			} else {
				log.Printf("User '%s' Successfully created", u.Name)

				if err := u.AuthorizeSSHKeys(); err != nil {
					log.Printf("Failed to add ssh key for user '%s': %v", u.Name, err)
				}
				if err := u.AuthorizeSudo(); err != nil {
					log.Printf("Failed to authorize sudo rights for user '%s': %v", u.Name, err)
				}
			}
		} else {
			log.Printf("User '%s' already exists. Skipping...", u.Name)
		}
	}
}

// parse ssh keys, and save next data from it. Example:
// ssh-rsa  aaaaaa...bbbbbbbb test@user.com
// 1. ssh-key       - ssh-rsa  aaaaaa...bbbbbbbb test@user.com
// 2. Username      - test
// 3. Email address - test@user.com
func (us *users) parseKey(sk sshKeys) {
	for _, key := range sk {
		var u user.Profile
		match := re.FindStringSubmatch(key)
		if len(match) == 3 {
			u.SSHAuthorizedKeys = append(u.SSHAuthorizedKeys, match[0])
			u.Name = match[2]
			u.Comment = match[1]
			*us = append(*us, u)
		}
	}
	*us = us.mergeUsers()
}

// merge dublicated users in one with multiple ssh-keys
func (us users) mergeUsers() users {
	encountered := map[string]int{}
	// Count how many times user appears in the file
	for i := range us {
		encountered[us[i].Name]++
	}

	var result users

	for key, value := range encountered {
		switch {
		case value == 1: // if user appears only once in list, add him to the new list of users
			for _, u := range us {
				if key == u.Name {
					result = append(result, u)
				}
			}
		case value > 1: // if user appears more then once on the list, merge them in one and add to the list of users
			var mergedUser user.Profile
			for _, u := range us {
				if key == u.Name {
					mergedUser.Name = key
					mergedUser.SSHAuthorizedKeys = append(mergedUser.SSHAuthorizedKeys, u.SSHAuthorizedKeys[0])
					mergedUser.Comment = fmt.Sprintf("%s%s;", mergedUser.Comment, u.Comment)
				}
			}
			result = append(result, mergedUser)
		}
	}

	return result
}

// read keys form the authorized_keys file
func (sk *sshKeys) readKeys(path string) {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		*sk = append(*sk, line)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	// remove the duplicated ssh keys
	*sk = removeDuplicatesUnordered(*sk)
}

// https://www.dotnetperls.com/duplicates-go
func removeDuplicatesUnordered(elements []string) []string {
	encountered := map[string]bool{}

	// Create a map of all unique elements.
	for v := range elements {
		encountered[elements[v]] = true
	}

	// Place all keys from the map into a slice.
	result := []string{}
	for key := range encountered {
		result = append(result, key)
	}
	return result
}
