package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"sync"
)

var (
	flags     map[string]string
	flagMutex sync.RWMutex
)

// Flag keys
const (
	WelcomeFlag           = "welcome"
	TaskInjectionFlag     = "task_injection"
	CrossUserAccessFlag   = "cross_user_access"
	AdminConsoleFlag      = "admin_console"
	SystemCompromisedFlag = "system_compromised"
)

// InitFlags initializes the flags from the JSON file
func InitFlags() {
	flagData, err := ioutil.ReadFile("flags.json")
	if err != nil {
		log.Fatalf("Error reading flags file: %v", err)
	}

	flagMutex.Lock()
	defer flagMutex.Unlock()

	err = json.Unmarshal(flagData, &flags)
	if err != nil {
		log.Fatalf("Error unmarshaling flags: %v", err)
	}
}

// GetFlag retrieves a flag by its key
func GetFlag(key string) string {
	flagMutex.RLock()
	defer flagMutex.RUnlock()
	return flags[key]
}

// UpdateFlag updates a flag value (this could be used by an admin interface)
func UpdateFlag(key, value string) {
	flagMutex.Lock()
	defer flagMutex.Unlock()
	flags[key] = value

	// Write updated flags back to file
	flagData, err := json.MarshalIndent(flags, "", "  ")
	if err != nil {
		log.Printf("Error marshaling flags: %v", err)
		return
	}

	err = ioutil.WriteFile("flags.json", flagData, 0644)
	if err != nil {
		log.Printf("Error writing flags file: %v", err)
	}
}
