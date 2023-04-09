package main

import (
	"bufio"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"log"
	"net/rpc"
	"os"
	"strconv"
	"strings"
	"unicode"

	"example.com/common"
	"github.com/gofiber/fiber/v2"
	"github.com/streadway/amqp"
)

func RSA_OAEP_Encrypt(secretMessage string, key rsa.PublicKey) string {
	label := []byte("OAEP Encrypted")
	rng := rand.Reader
	ciphertext, err := rsa.EncryptOAEP(sha256.New(), rng, &key, []byte(secretMessage), label)
	if err != nil {
		log.Fatal(err)
	}
	return base64.StdEncoding.EncodeToString(ciphertext)
}

func RSA_OAEP_Decrypt(cipherText string, privKey rsa.PrivateKey) (string, error) {
	ct, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil {
		return "", err
	}
	label := []byte("OAEP Encrypted")
	rng := rand.Reader
	plaintext, err := rsa.DecryptOAEP(sha256.New(), rng, &privKey, ct, label)
	if err != nil {
		log.Fatal(err)
	}
	return string(plaintext), nil
}

type config struct {
	messaging_server string
	server_port      string
	queue_name       string
	rpc_server       string
	logstash         string
}

func readFromStdinLine() string {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		return scanner.Text()
	}
	return ""
}

func readClientName() string {
	var current string
	for current = ""; len(current) == 0 || len(strings.Split(current, " ")) != 1; {
		fmt.Print("Enter name= ")
		current = readFromStdinLine()
	}
	return current
}

func readInt(message string) int {
	for {
		if len(message) != 0 {
			fmt.Print(message)
		}
		current := readFromStdinLine()
		val, e := strconv.Atoi(current)
		if e != nil {
			continue
		}
		return val
	}

}

var clientName string

func logLine(asd string) {
	log.Printf("Client %s: %s", clientName, asd)
}

func main() {
	configServer := config{}
	log.Println("Client nou proneste")
	clientName = readClientName()
	var file *os.File
	var err error
	file, err = os.Open("config.txt")
	if err != nil {
		file, err = os.Open("../config.txt")
		if err != nil {
			sysvar := strings.ToLower(os.Getenv("NON_INTERACTIVE"))
			switch sysvar {
			case "true", "t", "yes", "1", "y":
				panic(err)
			}
			scanner := bufio.NewScanner(os.Stdin)
			for scanner.Scan() {
				path := scanner.Text()
				file, err = os.Open(path)
				if err == nil {
					break
				}
			}
		}
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		value := scanner.Text()
		words := strings.SplitN(value, "=", 2)
		switch words[0] {
		case "MESSAGING_SERVER":
			configServer.messaging_server = words[1]
		case "SERVER_PORT":
			configServer.server_port = words[1]
		case "QUEUE_NAME":
			configServer.queue_name = words[1]
		case "RPC_PORT":
			configServer.rpc_server = words[1]
		case "LOGSTASH":
			configServer.logstash = words[1]
		}

	}
	file.Close()

	log.SetOutput(common.GetHTTPLogger(configServer.logstash))

	logLine("Initializez conexiune RPC")
	rpc_client, err := rpc.Dial("tcp", fmt.Sprintf("localhost:%s", configServer.rpc_server))
	if err != nil {
		panic(err)
	}

	logLine("Initializez conexiune RabbitMQ")
	connection, err := amqp.Dial(fmt.Sprintf("amqp://guest:guest@%s/", configServer.messaging_server))

	if err != nil {
		panic(err)
	}
	defer connection.Close()

	ch, err := connection.Channel()
	if err != nil {
		panic(err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		configServer.queue_name, // name
		false,                   // durable
		false,                   // delete when unused
		false,                   // exclusive
		false,                   // no-wait
		nil,                     // arguments
	)
	if err != nil {
		panic(err)
	}

	privatekey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatalf("Cannot generate RSA key.")
	}
	publickey := privatekey.PublicKey

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)

	if err != nil {
		panic(err)
	}

	logLine("Successfully connected to RabbitMQ instance")

	tokenChannel := make(chan string)
	go func() {
		for d := range msgs {
			word := d.Body
			if decrypted, err := RSA_OAEP_Decrypt(string(word), *privatekey); (err == nil) && isASCII(decrypted) {
				tokenChannel <- decrypted
			}
		}
	}()

	body := fmt.Sprintf("%s %d %s", publickey.N, publickey.E, clientName)
	err = ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		})

	if err != nil {
		panic(err)
	}

	token := <-tokenChannel
	logLine("Am primit token")
	ch.Close()
	connection.Close()

optionmenu:
	for {
		printMenu()
		choice := readFromStdinLine()
		val, err := strconv.Atoi(choice)
		if err != nil {
			continue
		}
		switch val {
		case 1:
			fmt.Println("Enter words (one by one, stops at empty):")
			fmt.Print("Word: ")
			word := readFromStdinLine()
			if len(word) == 0 {
				break
			}
			words := make([]string, 0)
			words = append(words, word)
			for {
				fmt.Print("Word: ")
				current := readFromStdinLine()
				if len(current) == 0 {
					break
				}
				if len(current) != len(word) {
					fmt.Println("Wrong size")
					continue
				}
				words = append(words, current)
			}
			wordsRequest := strings.Join(words, ", ")
			data := make(map[string]string)
			data["words"] = wordsRequest
			logLine(fmt.Sprintf("Clientul face request cu datele %v\n", words))
			_, body2, _ := doJSONRequest(data, token, getURL(configServer, "exercitiul1"))
			logLine(fmt.Sprintf("Client primeste %s\n", string(body2)))
		case 2:
			words := readArrayOfStrings("Enter words:", "Word: ")
			reply := 0
			logLine(fmt.Sprintf("Clientul face request cu datele %v\n", words))
			err = rpc_client.Call("NetRPCListener.Exercitiul2", words, &reply)
			if err != nil {
				panic(err)
			}
			logLine(fmt.Sprintf("Received: %v\n", reply))
		case 3:
			fmt.Println("Enter numbers (one by one, stops at empty):")
			words := make([]string, 0)
			for {
				fmt.Print("Number: ")
				current := readFromStdinLine()
				if len(current) == 0 {
					break
				}
				_, e := strconv.Atoi(current)
				if e != nil {
					panic(e)
				}
				words = append(words, current)
			}
			log.Printf("Clientul face request cu datele %v\n", words)
			wordsRequest := strings.Join(words, ", ")
			data := make(map[string]string)
			data["numbers"] = wordsRequest
			_, body2, _ := doJSONRequest(data, token, getURL(configServer, "exercitiul3"))
			log.Printf("Client primeste %s\n", string(body2))
		case 4:
			a := readInt("a= ")
			b := readInt("b= ")
			words := make([]int, 0)
			words = append(words, a)
			words = append(words, b)
			for {
				fmt.Print("Number: ")
				current := readFromStdinLine()
				if len(current) == 0 {
					break
				}
				val, e := strconv.Atoi(current)
				if e != nil {
					panic(e)
				}
				words = append(words, val)
			}
			log.Printf("Clientul face request cu datele %v\n", words)
			data := make(map[string][]int)
			data["numbers"] = words
			_, body2, _ := doJSONRequest(data, token, getURL(configServer, "exercitiul4"))
			log.Printf("Client primeste %s\n", string(body2))
		case 5:
			fmt.Println("Enter strings (one by one, stops at empty):")
			words := make([]string, 0)
			for {
				fmt.Print("String: ")
				current := readFromStdinLine()
				if len(current) == 0 {
					break
				}
				words = append(words, current)
			}
			log.Printf("Clientul face request cu datele %v\n", words)
			data := make(map[string][]string)
			data["numbers"] = words
			_, body2, _ := doJSONRequest(data, token, getURL(configServer, "exercitiul5"))
			log.Printf("Client primeste %s\n", string(body2))
		case 6:
			words := readArrayOfStrings("Enter words(cuvant STANGA/DREAPTA 2):", "Word: ")
			log.Printf("Clientul face request cu datele %v\n", words)
			data := make(map[string][]string)
			data["words"] = words
			_, body2, _ := doJSONRequest(data, token, getURL(configServer, "exercitiul6"))
			log.Printf("Client primeste %s\n", string(body2))
		case 7:
			fmt.Print("Word: ")
			word := readFromStdinLine()
			log.Printf("Clientul face request cu datele %v\n", word)
			data := make(map[string]string)
			data["word"] = word
			_, body2, _ := doJSONRequest(data, token, getURL(configServer, "exercitiul7"))
			log.Printf("Client primeste %s\n", string(body2))
		case 8:
			words := readArrayOfInts("Enter numbers (one by one, stops at empty):", "Number: ")
			log.Printf("Clientul face request cu datele %v\n", words)
			data := make(map[string][]int)
			data["numbers"] = words
			_, body2, _ := doJSONRequest(data, token, getURL(configServer, "exercitiul8"))
			log.Printf("Client primeste %s\n", string(body2))
		case 9:
			words := readArrayOfStrings("Enter strings one by one", "String: ")
			log.Printf("Clientul face request cu datele %v\n", words)
			data := make(map[string][]string)
			data["words"] = words
			_, body2, _ := doJSONRequest(data, token, getURL(configServer, "exercitiul9"))
			log.Printf("Client primeste %s\n", string(body2))
		case 10:
			words := readArrayOfStrings("Read numbers or strings:", "Number: ")
			log.Printf("Clientul face request cu datele %v\n", words)
			data := make(map[string][]string)
			data["numbers"] = words
			_, body2, _ := doJSONRequest(data, token, getURL(configServer, "exercitiul10"))
			log.Printf("Client primeste %s\n", string(body2))
		case 11:
			words := readArrayOfInts("Enter numbers:", "Number: ")
			reply := 0
			log.Printf("Client sends %v", words)
			err = rpc_client.Call("NetRPCListener.Exercitiul11", words, &reply)
			if err != nil {
				panic(err)
			}
			log.Printf("Received: %v\n", reply)
		case 12:
			words := readArrayOfInts("Enter numbers:", "Number: ")
			reply := 0
			log.Printf("Client sends %v", words)
			err = rpc_client.Call("NetRPCListener.Exercitiul12", words, &reply)
			if err != nil {
				panic(err)
			}
			log.Printf("Received: %v\n", reply)
		case 13:
			words := readArrayOfInts("Enter numbers:", "Number: ")
			var reply common.ReturnFloats
			log.Printf("Client sends %v", words)
			err = rpc_client.Call("NetRPCListener.Exercitiul13", words, &reply)
			if err != nil {
				panic(err)
			}
			log.Printf("Received: %v\n", reply)
		case 14:
			words := readArrayOfStrings("Enter passwords:", "Password: ")
			log.Printf("Clientul face request cu datele %v\n", words)
			data := make(map[string][]string)
			data["words"] = words
			_, body2, _ := doJSONRequest(data, token, getURL(configServer, "exercitiul14"))
			log.Printf("Client primeste %s\n", string(body2))

		case 0:
			break optionmenu
		default:
			fmt.Println("Unknown or not done.")
		}
		fmt.Print("Press enter to continue...")
		readFromStdinLine()
	}
	ch.Close()
}

func readArrayOfStrings(initial_message string, inbetween string) []string {
	if len(initial_message) != 0 {
		fmt.Println(initial_message)
	}
	words := make([]string, 0)
	for {
		if len(inbetween) != 0 {
			fmt.Print(inbetween)
		}
		current := readFromStdinLine()
		if len(current) == 0 {
			break
		}
		words = append(words, current)
	}
	return words
}

func readArrayOfInts(initial_message string, inbetween string) []int {
	if len(initial_message) != 0 {
		fmt.Println(initial_message)
	}
	words := make([]int, 0)
	for {
		if len(inbetween) != 0 {
			fmt.Print(inbetween)
		}
		current := readFromStdinLine()
		if len(current) == 0 {
			break
		}
		val, e := strconv.Atoi(current)
		if e != nil {
			panic(e)
		}
		words = append(words, val)
	}
	return words
}

func printMenu() {
	fmt.Println("=====MENU======")
	for i := 1; i <= 15; i++ {
		fmt.Printf("%d. Exercise %d\n", i, i)
	}
	fmt.Println("0. Exit")
	fmt.Print("Option= ")
}

func doJSONRequest(data interface{}, token string, url string) (int, []byte, []error) {
	a := fiber.AcquireAgent()
	a.JSON(data)
	req := a.Request()
	req.Header.SetMethod(fiber.MethodPost)
	req.Header.Set("Authorization", token)
	req.SetRequestURI(url)

	if err := a.Parse(); err != nil {
		panic(err)
	}

	return a.Bytes()
}

func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > unicode.MaxASCII {
			return false
		}
	}
	return true
}

func getURL(configServer config, endpoint string) string {
	return fmt.Sprintf("http://localhost:%s/%s", configServer.server_port, endpoint)
}
