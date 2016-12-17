package main

import (
	"bufio"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"regexp"

	"fmt"

	"github.com/trane9991/cruser/user"
)

type sshKey struct {
	Line     string
	Type     string
	Key      string
	Username string
	Email    string
}

type sshKeys []sshKey

var re = regexp.MustCompile(`(\S*) (\S*) ((\w*)@.*)`)

type users []user.Profile

func main() {
	fileWithKeys := flag.String("file", "users", "File with the list of SSH-keys and user emails in format of '~/.ssh/authorized_keys' file")
	dryRun := flag.Bool("dry-run", false, "Do not execute commands, just print them.")
	flag.Parse()

	// Read provided ssh keys
	lines := readLines(*fileWithKeys)

	// Remove dublicated SSH keys
	lines = removeDuplicatesUnordered(lines)

	// parse componetns of ssh keys
	sks := parseKeys(lines)

	// Transform list of SSH keys to the list of users
	us := keys2users(sks, false)

	// Checking if users exist, then read existing SSH keys
	// from this usera nd add them to the list
	us.checkUsers()

	// Merge users again with SSH keys from existing users
	us = us.mergeUsers()
	user.DryRun = *dryRun
	if user.DryRun {
		log.SetOutput(ioutil.Discard)
	}
	for _, u := range us {
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
			log.Printf("User '%s' already exists. Appending SSH keys forcefully (rewriting completely SSH-keys file with old and new ssh keys)...", u.Name)
			if err := u.AuthorizeSSHKeys(); err != nil {
				log.Printf("Failed to add ssh key for user '%s': %v", u.Name, err)
			}
		}
	}
}

// parse ssh keys, and save next data from it. Example:
// ssh-rsa  aaaaaa...bbbbbbbb test@user.com
// 0. ssh-key       - ssh-rsa  aaaaaa...bbbbbbbb test@user.com
// 1. Type          - ssh-rsa
// 2. Key           - aaaaaaa...bbbbbbbb
// 3. Email address - test@user.com
// 4. Username      - test
func parseKeys(lines []string) sshKeys {
	var sks sshKeys
	for _, key := range lines {
		var sk sshKey
		match := re.FindStringSubmatch(key)

		if len(match) == 5 {
			sk.Line = match[0]
			sk.Type = match[1]
			sk.Key = match[2]
			sk.Email = match[3]
			sk.Username = match[4]

			if sk.Type != "ssh-rsa" && sk.Type != "ssh-dss" && sk.Type != "ssh-ed25519" && sk.Type != "ecdsa-sha2-nistp256" {
				fmt.Printf("'%s' - not known SSH Key type ", sk.Type)
			} else {
				sks = append(sks, sk)
			}
		}
	}
	return sks
}

// convert list of keys to the list of SSH users
func keys2users(sks sshKeys, existing bool) users {
	encountered := map[string]int{}
	// Count how many times user appears in the file
	for i := range sks {
		encountered[sks[i].Username]++
	}
	var result users

	for key, value := range encountered {
		switch {
		case value == 1: // if user appears only once in list, add him to the new list of users
			for _, sk := range sks {
				if key == sk.Username {
					var u user.Profile
					u.Name = sk.Username
					u.SSHAuthorizedKeys = append(u.SSHAuthorizedKeys, sk.Line)
					u.Comment = sk.Email
					result = append(result, u)
				}
			}
		case value > 1: // if user appears more then once on the list, merge them in one and add to the list of users
			var u user.Profile
			for _, sk := range sks {
				if key == sk.Username {
					u.Name = sk.Username
					u.SSHAuthorizedKeys = append(u.SSHAuthorizedKeys, sk.Line)
					u.Comment = fmt.Sprintf("%s%s;", u.Comment, sk.Email)
				}
			}
			result = append(result, u)
		}
	}
	return result
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
					mergedUser.SSHAuthorizedKeys = append(mergedUser.SSHAuthorizedKeys, u.SSHAuthorizedKeys...)
					mergedUser.Comment = fmt.Sprintf("%s%s;", mergedUser.Comment, u.Comment)
				}
			}
			result = append(result, mergedUser)
		}
	}

	return result
}

// read lines from file
func readLines(path string) []string {
	var lines []string
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		lines = append(lines, line)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return lines
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

// read and parse ssh keys of existing user
func getSSHKeysFromExistingUser(u user.Profile) users {
	// Read provided ssh keys
	lines := readLines(u.AuthorizedKeysFile())
	// Remove dublicated SSH keys
	lines = removeDuplicatesUnordered(lines)
	// Get username from ssh keys emails
	sks := parseKeys(lines)
	us := keys2users(sks, true)
	return us
}

func (us *users) checkUsers() {
	for _, u := range *us {
		if u.Exists() {
			*us = append(*us, getSSHKeysFromExistingUser(u)...)
		}
	}
}
