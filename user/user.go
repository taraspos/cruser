package user

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

var (
	errorNoSuchUser = errors.New("User not exist")
	DryRun          = false
)

// Profile contains general user configurations
type Profile struct {
	Name              string
	SSHAuthorizedKeys []string
	Homedir           string
	Comment           string
	NoCreateHome      bool
	PrimaryGroup      string
	Groups            []string
	NoUserGroup       bool
	System            bool
	NoLogInit         bool
	Shell             string
	Sudoer            bool
}

// Validating sudoers line with visudo. Just to be sure
func (u Profile) validateSudo() error {
	_, err := exec.LookPath("visudo")
	if err != nil {
		return err
	}

	var stdout, stderr bytes.Buffer

	cmd := exec.Command("visudo", "-c", "-f", "-")
	if DryRun {
		cmdToStdout(cmd)
		return nil
	}

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	io.WriteString(stdin, u.sudoersLine())
	stdin.Close()
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf(stderr.String())
	}
	log.Printf("Sudoers line validation successfully passed. visudo output: %s", stdout.String())
	return nil
}

// Adds the sudoers line to the /etc/sudoers.d/${user}
func (u Profile) addSudo() error {
	err := createFile(u.sudoersFile())
	if err != nil {
		return err
	}

	if DryRun {
		fmt.Fprintf(os.Stdout, "Adding next content: '%s' to the sudoers file %s \n", []byte(u.sudoersLine()), u.sudoersFile())
		return nil
	}

	err = ioutil.WriteFile(u.sudoersFile(), []byte(u.sudoersLine()), 0440)
	if err != nil {
		return err
	}

	log.Printf("Successfully added line '%s' to the file '%s'", u.sudoersLine(), u.sudoersFile())

	return nil
}

func (u Profile) sudoersLine() string {
	return fmt.Sprintf("%s ALL=(ALL) NOPASSWD:ALL", u.Name)
}

func (u Profile) sudoersFile() string {
	return fmt.Sprintf("/etc/sudoers.d/%s", u.Name)
}

func (u Profile) authorizedKeysDir() string {
	return fmt.Sprintf("/home/%s/.ssh", u.Name)
}

func (u Profile) authorizedKeysFile() string {
	return fmt.Sprintf("%s/authorized_keys", u.authorizedKeysDir())
}

func createDir(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {

		if DryRun {
			fmt.Fprintf(os.Stdout, "Creating dir '%s' \n", path)
			return nil
		}

		err := os.MkdirAll(path, 0600)
		if err != nil {
			return err
		}
	}
	return nil
}

// create file if not exists
func createFile(path string) error {
	var _, err = os.Stat(path)

	if os.IsNotExist(err) {

		if DryRun {
			fmt.Fprintf(os.Stdout, "Creating file '%s' \n", path)
			return nil
		}

		var file, err = os.Create(path)
		file.Chmod(0600)
		file.Close()
		return err
	}

	return nil
}

// seting ownership of files to the user
func setOwnership(u Profile, path string) error {
	uid, err := u.getUID()
	if err != nil {
		return err
	}
	gid, err := u.getGID()
	if err != nil {
		return err
	}

	if DryRun {
		fmt.Fprintf(os.Stdout, "Changing ownership of '%s', to the uid=%d and gid=%d \n", path, uid, gid)
		return nil
	}

	err = os.Chown(path, uid, gid)
	if err != nil {
		return err
	}
	return nil
}

// AuthorizeSSHKeys adds the provided SSH public keys to the user's list of authorized keys
func (u Profile) AuthorizeSSHKeys() error {
	for i, key := range u.SSHAuthorizedKeys {
		u.SSHAuthorizedKeys[i] = strings.TrimSpace(key)
	}

	// join all keys with newlines, ensuring the resulting string
	// also ends with a newline
	joined := fmt.Sprintf("%s\n", strings.Join(u.SSHAuthorizedKeys, "\n"))

	err := createDir(u.authorizedKeysDir())
	if err != nil {
		return err
	}
	err = setOwnership(u, u.authorizedKeysDir())
	if err != nil {
		return err
	}
	err = createFile(u.authorizedKeysFile())
	if err != nil {
		return err
	}
	err = setOwnership(u, u.authorizedKeysFile())
	if err != nil {
		return err
	}

	if DryRun {
		fmt.Fprintf(os.Stdout, "Writing SSH keys to the file %s \n", u.authorizedKeysFile())
		return nil
	}

	f, err := os.OpenFile(u.authorizedKeysFile(), os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}

	defer f.Close()

	if _, err = f.WriteString(joined); err != nil {
		return err
	}

	log.Printf("Public SSH keys successfully added to the file '%s'", u.authorizedKeysFile())
	return nil
}

// Create adds user to the Linux system via "useradd" command
func (u Profile) Create() error {
	_, err := exec.LookPath("useradd")
	if err != nil {
		return err
	}

	var args []string

	if u.Comment != "" {
		args = append(args, "--comment", fmt.Sprintf("%q", u.Comment))
	}

	if u.Homedir != "" {
		args = append(args, "--home-dir", u.Homedir)
	}

	if u.NoCreateHome {
		args = append(args, "--no-create-home")
	} else {
		args = append(args, "--create-home")
	}

	if u.PrimaryGroup != "" {
		args = append(args, "--gid", u.PrimaryGroup)
	}

	if len(u.Groups) > 0 {
		args = append(args, "--groups", strings.Join(u.Groups, ","))
	}

	if u.NoUserGroup {
		args = append(args, "--no-user-group")
	}

	if u.System {
		args = append(args, "--system")
	}

	if u.NoLogInit {
		args = append(args, "--no-log-init")
	}

	if u.Shell != "" {
		args = append(args, "--shell", u.Shell)
	}

	args = append(args, u.Name)

	cmd := exec.Command("useradd", args...)
	if DryRun {
		cmdToStdout(cmd)
		return nil
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Command 'useradd %s' failed: %v\n%s", strings.Join(args, " "), err, output)
	}
	return err
}

// AuthorizeSudo validates and adds sudo line to the /etc/sudoers.d/$(user)
func (u Profile) AuthorizeSudo() error {
	if err := u.validateSudo(); err != nil {
		return err
	}
	if err := u.addSudo(); err != nil {
		return err
	}
	return nil
}

// Exists check if user exists, and return true/false
func (u Profile) Exists() bool {
	_, err := exec.LookPath("id")
	if err != nil {
		log.Fatal(err)
	}
	cmd := exec.Command("id", u.Name)
	if DryRun {
		cmdToStdout(cmd)
		return true
	}

	_, err = cmd.CombinedOutput()
	return err == nil
}

func (u Profile) getUID() (int, error) {
	if !u.Exists() {
		return 0, errorNoSuchUser
	}

	cmd := exec.Command("id", "-u", u.Name)
	if DryRun {
		cmdToStdout(cmd)
		return 0, nil
	}

	uidStr, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	uid, err := strconv.Atoi(strings.Replace(string(uidStr), "\n", "", -1))
	if err != nil {
		return 0, err
	}
	return uid, nil
}

func (u Profile) getGID() (int, error) {
	if !u.Exists() {
		return 0, errorNoSuchUser
	}

	cmd := exec.Command("id", "-g", u.Name)
	if DryRun {
		cmdToStdout(cmd)
		return 0, nil
	}

	gidStr, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	gid, err := strconv.Atoi(strings.Replace(string(gidStr), "\n", "", -1))
	if err != nil {
		return 0, err
	}
	return gid, nil
}

func cmdToStdout(cmd *exec.Cmd) {
	fmt.Fprintf(os.Stdout, "%s %s\n", cmd.Path, strings.Join(cmd.Args[1:], " "))
}
