package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/net/publicsuffix"
)

// an HTTP client that automatically adds the session cookie
type Session struct{ http.Client }

// a device group, which in our case is a building
type Group struct {
	GroupID        int      `json:"groupId"`
	GroupName      string   `json:"groupName"`
	OrganizationID int      `json:"organizationId"`
	Devices        []Device `json:"devices"`
}

// an individual AirTame device
type Device struct {
	Platform          string `json:"platform"`
	Version           string `json:"version"`
	State             string `json:"state"`
	ScreenshotEnabled bool   `json:"screenshotEnabled"`
	SettingsState     string `json:"settingsState"`
	ID                int    `json:"id"`
	Ap24Enabled       bool   `json:"ap24Enabled"`
	Ap24Channel       int    `json:"ap24Channel"`
	Ap52Enabled       bool   `json:"ap52Enabled"`
	Ap52Channel       int    `json:"ap52Channel"`
	BackgroundType    string `json:"backgroundType"`
	DeviceName        string `json:"deviceName"`
	LastSeen          int    `json:"lastSeen"`
	LastConnected     int    `json:"lastConnected"`
	IsOnline          bool   `json:"isOnline"`
	NetworkState      struct {
		Online     bool `json:"online"`
		Interfaces []struct {
			Frequency      int    `json:"frequency"`
			IP             string `json:"ip"`
			Mac            string `json:"mac"`
			Mode           string `json:"mode"`
			Name           string `json:"name"`
			SignalStrength int    `json:"signal_strength"`
			Ssid           string `json:"ssid"`
			Status         string `json:"status"`
			Type           string `json:"type"`
		} `json:"interfaces"`
	} `json:"networkState"`
	UpdateAvailable       bool   `json:"updateAvailable"`
	UpdateChannel         string `json:"updateChannel"`
	UpdateProgress        int    `json:"updateProgress"`
	HomescreenOrientation string `json:"homescreenOrientation"`
}

// log in and get a session cookie
func Authenticate(email, password string) (Session, error) {
	// the email and password are passed as JSON
	// "Persist" is the "keep me logged in" checkbox
	reqBody, err := json.Marshal(struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Persist  bool   `json:"persist"`
	}{
		Email:    email,
		Password: password,
		Persist:  false,
	})
	if err != nil {
		return Session{}, err
	}

	// create a new HTTP client with a cookie jar
	jar, err := cookiejar.New(&cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	})
	if err != nil {
		return Session{}, err
	}
	client := http.Client{
		Jar: jar,
	}

	// POST the login request and fill in the cookie jar
	resp, err := client.Post(
		"https://airtame.cloud/api/authenticate",
		"application/json",
		bytes.NewReader(reqBody),
	)
	if err != nil {
		return Session{}, err
	}
	defer resp.Body.Close()

	// the most common error is that the email or password is incorrect.
	// for everything else, we just return "unexpected status code"
	if resp.StatusCode == http.StatusUnauthorized {
		return Session{}, errors.New("authentication failed")
	} else if resp.StatusCode != http.StatusOK {
		io.Copy(os.Stderr, resp.Body)
		return Session{}, errors.New("unexpected status code")
	}

	return Session{client}, nil
}

// just like the API call, this returns a list of all the devices grouped by
// building.
func (s Session) Devices() ([]Group, error) {
	resp, err := s.Get("https://airtame.cloud/api/devices")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		io.Copy(os.Stderr, resp.Body)
		return nil, errors.New("unexpected status code")
	}

	var groups []Group
	err = json.NewDecoder(resp.Body).Decode(&groups)
	if err != nil {
		return nil, err
	}

	return groups, nil
}

// this returns all devices, not grouped
func (s Session) FlatDevices() ([]Device, error) {
	// get devices by group
	groups, err := s.Devices()
	if err != nil {
		return nil, err
	}

	// flatten the groups into a single list of devices
	var devices []Device
	for _, group := range groups {
		devices = append(devices, group.Devices...)
	}

	return devices, nil
}

// this reboots all the devices you pass to it
func (s Session) BulkReboot(devices []Device) error {
	var deviceIDs []int
	for _, device := range devices {
		deviceIDs = append(deviceIDs, device.ID)
	}

	reqBody, err := json.Marshal(struct {
		DeviceIDs []int `json:"deviceIds"`
	}{
		DeviceIDs: deviceIDs,
	})
	if err != nil {
		return err
	}

	resp, err := s.Post(
		"https://airtame.cloud/api/devices/commands/reboot/bulk",
		"application/json",
		bytes.NewReader(reqBody),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		io.Copy(os.Stderr, resp.Body)
		return errors.New("unexpected status code")
	}

	// check that the "errors" array is empty
	var respObj struct {
		Errors []interface{} `json:"errors"`
	}
	err = json.NewDecoder(resp.Body).Decode(&respObj)
	if err != nil {
		return err
	}

	if len(respObj.Errors) > 0 {
		for _, err := range respObj.Errors {
			// these might be device IDs or error strings. I don't know.
			log.Println(err)
		}
		return errors.New("errors occurred")
	}

	return nil
}

func main() {
	var email, password string
	flag.StringVar(&email, "email", "", "email")
	flag.StringVar(&password, "password", "", "password")
	flag.Usage = func() {
		fmt.Fprintf(
			os.Stderr,
			"Usage: %s --email <email> --password <password>\n"+
				"\t{devices|flatdevices|reboot <device_id...>|rebootall}\n",
			filepath.Base(os.Args[0]),
		)
	}
	flag.Parse()

	// set up an encoder with indentation
	jsonOut := json.NewEncoder(os.Stdout)
	jsonOut.SetIndent("", "\t")

	switch strings.ToLower(flag.Arg(0)) {
	case "devices":
		s, err := Authenticate(email, password)
		if err != nil {
			log.Fatal(err)
		}
		devices, err := s.Devices()
		if err != nil {
			log.Fatal(err)
		}
		err = jsonOut.Encode(devices)
		if err != nil {
			log.Fatal(err)
		}
	case "flatdevices":
		s, err := Authenticate(email, password)
		if err != nil {
			log.Fatal(err)
		}
		devices, err := s.FlatDevices()
		if err != nil {
			log.Fatal(err)
		}
		err = jsonOut.Encode(devices)
		if err != nil {
			log.Fatal(err)
		}
	case "reboot":
		s, err := Authenticate(email, password)
		if err != nil {
			log.Fatal(err)
		}
		deviceIDs := flag.Args()[1:]
		if len(deviceIDs) == 0 {
			log.Fatal("no devices specified")
		}
		// convert the device IDs to "hollow" device objects
		var devices []Device
		for _, deviceID := range deviceIDs {
			idInt, err := strconv.Atoi(deviceID)
			if err != nil {
				log.Fatal(err)
			}
			devices = append(devices, Device{ID: idInt})
		}
		// reboot the devices
		err = s.BulkReboot(devices)
		if err != nil {
			log.Fatal(err)
		}
	case "rebootall":
		s, err := Authenticate(email, password)
		if err != nil {
			log.Fatal(err)
		}
		// get all devices
		devices, err := s.FlatDevices()
		if err != nil {
			log.Fatal(err)
		}
		// filter out the devices that are not online
		var onlineDevices []Device
		for _, device := range devices {
			if device.NetworkState.Online {
				onlineDevices = append(onlineDevices, device)
			} else {
				log.Printf("skipping %s (offline)", device.DeviceName)
			}
		}
		// reboot the devices
		err = s.BulkReboot(onlineDevices)
		if err != nil {
			log.Fatal(err)
		}
	default:
		flag.Usage()
	}
}
