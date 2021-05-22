package client

import (
	"bufio"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"syscall"

	"golang.org/x/term"
)

type ReqType uint8

const (
	post ReqType = iota
	get
	del
)

type credentialsRequest struct {
	Username   string
	Password   string
	SecretList map[string]string
}

func Login(rt uint8) error {
	creds, err := enterCredentials(false)
	if err != nil {
		return err
	}

	var req *http.Request
	if ReqType(rt) == post {
		req, err = postSecrets(creds)
		if err != nil {
			return err
		}
	} else if ReqType(rt) == get {
		req, err = getSecrets(creds)
		if err != nil {
			return err
		}
	} else if ReqType(rt) == del {
		req, err = delAccount(creds)
		if err != nil {
			return err
		}
	} else {
		return errors.New("unknown request type")
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusNoContent {
		log.Println("You do not have any secrets stored")
		return nil
	} else if resp.StatusCode != http.StatusOK {
		log.Fatal("Server responded with " + string(body))
		return nil
	}

	secrets := make(map[string][]byte)
	if ReqType(rt) == get {
		if err = json.Unmarshal(body, &secrets); err != nil {
			return err
		}
		fmt.Println("Your secrets:")
		for k, v := range secrets {
			secrets[k], err = decryptAESGCM([]byte(creds.Password), v)
			if err != nil {
				return err
			}
			fmt.Printf("%s: %s\n", k, secrets[k])
		}
	} else {
		log.Println("Server responded with " + string(body))
	}

	return nil
}

func SignUp() error {
	creds, err := enterCredentials(true)
	if err != nil {
		return err
	}
	mashalledCreds, err := json.Marshal(creds)
	if err != nil {
		return err
	}
	resp, err := http.Post("http://localhost:2000/signup", "application/json", bytes.NewBuffer(mashalledCreds))
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusOK {
		log.Println("Server responded with " + string(body))
	} else {
		log.Fatal("Server responded with " + string(body))
	}
	return nil
}

func enterCredentials(signup bool) (credentialsRequest, error) {
	username, err := getUserInput("Enter username: ")
	if err != nil {
		return credentialsRequest{}, err
	}

	fmt.Print("Enter password: ")
	masterPassword, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Print("\n")
	if err != nil {
		return credentialsRequest{}, err
	}

	if signup {
		fmt.Print("Confirm password: ")
		retryPassword, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Print("\n")
		if err != nil {
			return credentialsRequest{}, err
		}

		if string(retryPassword) != string(masterPassword) {
			return credentialsRequest{}, errors.New("entered passwords did not match")
		}
	}

	return credentialsRequest{
		Username: username,
		Password: string(masterPassword),
	}, nil

}

func postSecrets(creds credentialsRequest) (*http.Request, error) {
	secrets := make(map[string]string)
	for {
		service, err := getUserInput("\nEnter service name: ")
		if err != nil {
			return nil, err
		}

		password, err := getUserInput("Enter service password: ")
		if err != nil {
			return nil, err
		}

		secrets[service] = password

		fmt.Print("Press 1 to exit terminal and send secrets to server, ",
			"or any other key to continue adding secrets\n")

		exit, err := getUserInput("")
		if err != nil {
			return nil, err
		}
		if exit == "1" {
			break
		}
	}
	creds.SecretList = secrets

	mashalledCreds, err := json.Marshal(creds)
	if err != nil {
		return &http.Request{}, err
	}

	reader := strings.NewReader(string(mashalledCreds))

	req, err := http.NewRequest("POST", "http://localhost:2000/login", reader)
	if err != nil {
		return &http.Request{}, err
	}

	return req, nil
}

func getSecrets(creds credentialsRequest) (*http.Request, error) {
	mashalledCreds, err := json.Marshal(creds)
	if err != nil {
		return &http.Request{}, err
	}

	reader := strings.NewReader(string(mashalledCreds))

	req, err := http.NewRequest("GET", "http://localhost:2000/login", reader)
	if err != nil {
		return &http.Request{}, err
	}

	return req, nil
}

func delAccount(creds credentialsRequest) (*http.Request, error) {
	mashalledCreds, err := json.Marshal(creds)
	if err != nil {
		return &http.Request{}, err
	}

	reader := strings.NewReader(string(mashalledCreds))

	req, err := http.NewRequest("DELETE", "http://localhost:2000/login", reader)
	if err != nil {
		return &http.Request{}, err
	}

	return req, nil
}

func decryptAESGCM(masterpass, ciphertext []byte) ([]byte, error) {
	keyLength := 2 * aes.BlockSize
	key := make([]byte, keyLength)

	if len(masterpass) >= keyLength {
		copy(key, masterpass[0:keyLength])
	} else {
		copy(key, masterpass)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Get the nonce from the ciphertext
	nonce := ciphertext[:12]

	return aesgcm.Open(nil, nonce, ciphertext[12:], nil)
}

func getUserInput(help string) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(help)
	userInput, err := reader.ReadString('\n')
	if err != nil {
		return userInput, err
	}

	return strings.Replace(userInput, "\n", "", -1), nil
}
